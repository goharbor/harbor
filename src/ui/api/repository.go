// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	registry_error "github.com/vmware/harbor/src/common/utils/error"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/notary"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/ui/config"
)

// RepositoryAPI handles request to /api/repositories /api/repositories/tags /api/repositories/manifests, the parm has to be put
// in the query string as the web framework can not parse the URL if it contains veriadic sectors.
type RepositoryAPI struct {
	BaseController
}

type repoResp struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	ProjectID    int64     `json:"project_id"`
	Description  string    `json:"description"`
	PullCount    int64     `json:"pull_count"`
	StarCount    int64     `json:"star_count"`
	TagsCount    int64     `json:"tags_count"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}

type tag struct {
	Digest        string    `json:"digest"`
	Name          string    `json:"name"`
	Architecture  string    `json:"architecture"`
	OS            string    `json:"os"`
	DockerVersion string    `json:"docker_version"`
	Author        string    `json:"author"`
	Created       time.Time `json:"created"`
}

type tagResp struct {
	tag
	Signature    *notary.Target          `json:"signature"`
	ScanOverview *models.ImgScanOverview `json:"scan_overview,omitempty"`
}

type manifestResp struct {
	Manifest interface{} `json:"manifest"`
	Config   interface{} `json:"config,omitempty" `
}

// Get ...
func (ra *RepositoryAPI) Get() {
	projectID, err := ra.GetInt64("project_id")
	if err != nil || projectID <= 0 {
		ra.HandleBadRequest(fmt.Sprintf("invalid project_id %s", ra.GetString("project_id")))
		return
	}

	exist, err := ra.ProjectMgr.Exist(projectID)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to check the existence of project %d: %v",
			projectID, err))
		return
	}

	if !exist {
		ra.HandleNotFound(fmt.Sprintf("project %d not found", projectID))
		return
	}

	if !ra.SecurityCtx.HasReadPerm(projectID) {
		if !ra.SecurityCtx.IsAuthenticated() {
			ra.HandleUnauthorized()
			return
		}
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}

	keyword := ra.GetString("q")

	total, err := dao.GetTotalOfRepositoriesByProject(
		[]int64{projectID}, keyword)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to get total of repositories of project %d: %v",
			projectID, err))
		return
	}

	page, pageSize := ra.GetPaginationParams()

	repositories, err := getRepositories(projectID,
		keyword, pageSize, pageSize*(page-1))
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to get repository: %v", err))
		return
	}

	ra.SetPaginationHeader(total, page, pageSize)
	ra.Data["json"] = repositories
	ra.ServeJSON()
}

func getRepositories(projectID int64, keyword string,
	limit, offset int64) ([]*repoResp, error) {
	repositories, err := dao.GetRepositoriesByProject(projectID, keyword, limit, offset)
	if err != nil {
		return nil, err
	}

	return populateTagsCount(repositories)
}

func populateTagsCount(repositories []*models.RepoRecord) ([]*repoResp, error) {
	result := []*repoResp{}
	for _, repository := range repositories {
		repo := &repoResp{
			ID:           repository.RepositoryID,
			Name:         repository.Name,
			ProjectID:    repository.ProjectID,
			Description:  repository.Description,
			PullCount:    repository.PullCount,
			StarCount:    repository.StarCount,
			CreationTime: repository.CreationTime,
			UpdateTime:   repository.UpdateTime,
		}

		tags, err := getTags(repository.Name)
		if err != nil {
			return nil, err
		}
		repo.TagsCount = int64(len(tags))
		result = append(result, repo)
	}
	return result, nil
}

// Delete ...
func (ra *RepositoryAPI) Delete() {
	// using :splat to get * part in path
	repoName := ra.GetString(":splat")

	projectName, _ := utils.ParseRepository(repoName)
	exist, err := ra.ProjectMgr.Exist(projectName)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to check the existence of project %s: %v",
			projectName, err))
		return
	}

	if !exist {
		ra.HandleNotFound(fmt.Sprintf("project %s not found", projectName))
		return
	}

	if !ra.SecurityCtx.IsAuthenticated() {
		ra.HandleUnauthorized()
		return
	}

	if !ra.SecurityCtx.HasAllPerm(projectName) {
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}

	rc, err := ra.initRepositoryClient(repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	tags := []string{}
	tag := ra.GetString(":tag")
	if len(tag) == 0 {
		tagList, err := rc.ListTag()
		if err != nil {
			if regErr, ok := err.(*registry_error.Error); ok {
				ra.CustomAbort(regErr.StatusCode, regErr.Detail)
			}

			log.Errorf("error occurred while listing tags of %s: %v", repoName, err)
			ra.CustomAbort(http.StatusInternalServerError, "internal error")
		}

		// TODO remove the logic if the bug of registry is fixed
		if len(tagList) == 0 {
			ra.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}

		tags = append(tags, tagList...)
	} else {
		tags = append(tags, tag)
	}

	if config.WithNotary() {
		signedTags, err := getSignatures(ra.SecurityCtx.GetUsername(), repoName)
		if err != nil {
			ra.HandleInternalServerError(fmt.Sprintf(
				"failed to get signatures for repository %s: %v", repoName, err))
			return
		}

		for _, t := range tags {
			digest, _, err := rc.ManifestExist(t)
			if err != nil {
				log.Errorf("Failed to Check the digest of tag: %s, error: %v", t, err.Error())
				ra.CustomAbort(http.StatusInternalServerError, err.Error())
			}
			log.Debugf("Tag: %s, digest: %s", t, digest)
			if _, ok := signedTags[digest]; ok {
				log.Errorf("Found signed tag, repostory: %s, tag: %s, deletion will be canceled", repoName, t)
				ra.CustomAbort(http.StatusPreconditionFailed, fmt.Sprintf("tag %s is signed", t))
			}
		}
	}

	for _, t := range tags {
		if err = rc.DeleteTag(t); err != nil {
			if regErr, ok := err.(*registry_error.Error); ok {
				if regErr.StatusCode == http.StatusNotFound {
					continue
				}
				ra.CustomAbort(regErr.StatusCode, regErr.Detail)
			}
			log.Errorf("error occurred while deleting tag %s:%s: %v", repoName, t, err)
			ra.CustomAbort(http.StatusInternalServerError, "internal error")
		}
		log.Infof("delete tag: %s:%s", repoName, t)
		go TriggerReplicationByRepository(repoName, []string{t}, models.RepOpDelete)

		go func(tag string) {
			project, err := ra.ProjectMgr.Get(projectName)
			if err != nil {
				log.Errorf("failed to get the project %s: %v",
					projectName, err)
				return
			}

			if project == nil {
				log.Errorf("project %s not found", projectName)
				return
			}

			if err := dao.AddAccessLog(models.AccessLog{
				Username:  ra.SecurityCtx.GetUsername(),
				ProjectID: project.ProjectID,
				RepoName:  repoName,
				RepoTag:   tag,
				Operation: "delete",
				OpTime:    time.Now(),
			}); err != nil {
				log.Errorf("failed to add access log: %v", err)
			}
		}(t)
	}

	exist, err = repositoryExist(repoName, rc)
	if err != nil {
		log.Errorf("failed to check the existence of repository %s: %v", repoName, err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}
	if !exist {
		if err = dao.DeleteRepository(repoName); err != nil {
			log.Errorf("failed to delete repository %s: %v", repoName, err)
			ra.CustomAbort(http.StatusInternalServerError, "")
		}
	}
}

// GetTags returns tags of a repository
func (ra *RepositoryAPI) GetTags() {
	repoName := ra.GetString(":splat")

	projectName, _ := utils.ParseRepository(repoName)
	exist, err := ra.ProjectMgr.Exist(projectName)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to check the existence of project %s: %v",
			projectName, err))
		return
	}

	if !exist {
		ra.HandleNotFound(fmt.Sprintf("project %s not found", projectName))
		return
	}

	if !ra.SecurityCtx.HasReadPerm(projectName) {
		if !ra.SecurityCtx.IsAuthenticated() {
			ra.HandleUnauthorized()
			return
		}
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}

	client, err := ra.initRepositoryClient(repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	// get tags
	tags, err := getDetailedTags(client)
	if err != nil {
		regErr, ok := err.(*registry_error.Error)
		if !ok {
			ra.HandleInternalServerError(fmt.Sprintf(
				"failed to list tags of repository %s: %v", repoName, err))
			return
		}
		ra.RenderError(regErr.StatusCode, regErr.Detail)
		return
	}

	// get signatures
	signatures := map[string]*notary.Target{}
	if config.WithNotary() {
		signatures, err = getSignatures(repoName, ra.SecurityCtx.GetUsername())
		if err != nil {
			ra.HandleInternalServerError(fmt.Sprintf(
				"failed to get signatures of repository %s: %v", repoName, err))
			return
		}
	}

	// assemble the response
	tagResps := []*tagResp{}
	for _, tag := range tags {
		tagResp := &tagResp{
			tag: *tag,
		}
		//TODO: only when deployed with Clair...
		tagResp.ScanOverview = getScanOverview(tag.Digest, tag.Name)

		// compare both digest and tag
		if signature, ok := signatures[tag.Digest]; ok {
			if tag.Name == signature.Tag {
				tagResp.Signature = signature
			}
		}

		tagResps = append(tagResps, tagResp)
	}

	ra.Data["json"] = tagResps
	ra.ServeJSON()
}

// get tags of the repository, read manifest for every tag
// and assemble necessary attrs(os, architecture, etc.) into
// one struct
func getDetailedTags(client *registry.Repository) ([]*tag, error) {
	ts, err := getSimpleTags(client)
	if err != nil {
		return nil, err
	}

	list := []*tag{}

	for _, t := range ts {
		// the ignored manifest can be used to calculate the image size
		digest, _, config, err := getV2Manifest(client, t)
		if err != nil {
			return nil, err
		}

		tag := &tag{}
		if err = json.Unmarshal(config, tag); err != nil {
			return nil, err
		}

		tag.Name = t
		tag.Digest = digest

		list = append(list, tag)
	}

	return list, nil
}

// get v2 manifest of tag, returns digest, manifest,
// manifest config and error. The manifest config contains
// architecture, os, author, etc.
func getV2Manifest(client *registry.Repository, tag string) (
	string, *schema2.DeserializedManifest, []byte, error) {
	digest, _, payload, err := client.PullManifest(tag, []string{schema2.MediaTypeManifest})
	if err != nil {
		return "", nil, nil, err
	}

	manifest := &schema2.DeserializedManifest{}
	if err = manifest.UnmarshalJSON(payload); err != nil {
		return "", nil, nil, err
	}

	_, reader, err := client.PullBlob(manifest.Target().Digest.String())
	if err != nil {
		return "", nil, nil, err
	}

	config, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", nil, nil, err
	}
	return digest, manifest, config, nil
}

// return tag name list for the repository
func getSimpleTags(client *registry.Repository) ([]string, error) {
	tags := []string{}

	ts, err := client.ListTag()
	if err != nil {
		// TODO remove the logic if the bug of registry is fixed
		// It's a workaround for a bug of registry: when listing tags of
		// a repository which is being pushed, a "NAME_UNKNOWN" error will
		// been returned, while the catalog API can list this repository.

		if regErr, ok := err.(*registry_error.Error); ok &&
			regErr.StatusCode == http.StatusNotFound {
			return tags, nil
		}

		return nil, err
	}

	tags = append(tags, ts...)
	sort.Strings(tags)

	return tags, nil
}

// GetManifests returns the manifest of a tag
func (ra *RepositoryAPI) GetManifests() {
	repoName := ra.GetString(":splat")
	tag := ra.GetString(":tag")

	version := ra.GetString("version")
	if len(version) == 0 {
		version = "v2"
	}

	if version != "v1" && version != "v2" {
		ra.CustomAbort(http.StatusBadRequest, "version should be v1 or v2")
	}

	projectName, _ := utils.ParseRepository(repoName)
	exist, err := ra.ProjectMgr.Exist(projectName)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to check the existence of project %s: %v",
			projectName, err))
		return
	}

	if !exist {
		ra.HandleNotFound(fmt.Sprintf("project %s not found", projectName))
		return
	}

	if !ra.SecurityCtx.HasReadPerm(projectName) {
		if !ra.SecurityCtx.IsAuthenticated() {
			ra.HandleUnauthorized()
			return
		}

		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}

	rc, err := ra.initRepositoryClient(repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	manifest, err := getManifest(rc, tag, version)
	if err != nil {
		if regErr, ok := err.(*registry_error.Error); ok {
			ra.CustomAbort(regErr.StatusCode, regErr.Detail)
		}

		log.Errorf("error occurred while getting manifest of %s:%s: %v", repoName, tag, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	ra.Data["json"] = manifest
	ra.ServeJSON()
}

func getManifest(client *registry.Repository,
	tag, version string) (*manifestResp, error) {
	result := &manifestResp{}

	mediaTypes := []string{}
	switch version {
	case "v1":
		mediaTypes = append(mediaTypes, schema1.MediaTypeManifest)
	case "v2":
		mediaTypes = append(mediaTypes, schema2.MediaTypeManifest)
	}

	_, mediaType, payload, err := client.PullManifest(tag, mediaTypes)
	if err != nil {
		return nil, err
	}

	manifest, _, err := registry.UnMarshal(mediaType, payload)
	if err != nil {
		return nil, err
	}

	result.Manifest = manifest

	deserializedmanifest, ok := manifest.(*schema2.DeserializedManifest)
	if ok {
		_, data, err := client.PullBlob(deserializedmanifest.Target().Digest.String())
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadAll(data)
		if err != nil {
			return nil, err
		}

		result.Config = string(b)
	}

	return result, nil
}

func (ra *RepositoryAPI) initRepositoryClient(repoName string) (r *registry.Repository, err error) {
	endpoint, err := config.RegistryURL()
	if err != nil {
		return nil, err
	}

	return NewRepositoryClient(endpoint, true, ra.SecurityCtx.GetUsername(),
		repoName, "repository", repoName, "pull", "push", "*")
}

//GetTopRepos returns the most populor repositories
func (ra *RepositoryAPI) GetTopRepos() {
	count, err := ra.GetInt("count", 10)
	if err != nil || count <= 0 {
		ra.CustomAbort(http.StatusBadRequest, "invalid count")
	}

	projectIDs := []int64{}
	projects, err := ra.ProjectMgr.GetPublic()
	if err != nil {
		log.Errorf("failed to get the public projects: %v", err)
		return
	}
	if ra.SecurityCtx.IsAuthenticated() {
		list, err := ra.ProjectMgr.GetByMember(ra.SecurityCtx.GetUsername())
		if err != nil {
			log.Errorf("failed to get projects which the user %s is a member of: %v",
				ra.SecurityCtx.GetUsername(), err)
			return
		}
		projects = append(projects, list...)
	}

	for _, project := range projects {
		projectIDs = append(projectIDs, project.ProjectID)
	}

	repos, err := dao.GetTopRepos(projectIDs, count)
	if err != nil {
		log.Errorf("failed to get top repos: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "internal server error")
	}

	result, err := populateTagsCount(repos)
	if err != nil {
		log.Errorf("failed to popultate tags count to repositories: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "internal server error")
	}

	ra.Data["json"] = result
	ra.ServeJSON()
}

//GetSignatures returns signatures of a repository
func (ra *RepositoryAPI) GetSignatures() {
	repoName := ra.GetString(":splat")
	targets, err := notary.GetInternalTargets(config.InternalNotaryEndpoint(),
		ra.SecurityCtx.GetUsername(), repoName)
	if err != nil {
		log.Errorf("Error while fetching signature from notary: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}
	ra.Data["json"] = targets
	ra.ServeJSON()
}

//ScanImage handles request POST /api/repository/$repository/tags/$tag/scan to trigger image scan manually.
func (ra *RepositoryAPI) ScanImage() {
	if !config.WithClair() {
		log.Warningf("Harbor is not deployed with Clair, scan is disabled.")
		ra.RenderError(http.StatusServiceUnavailable, "")
		return
	}
	repoName := ra.GetString(":splat")
	tag := ra.GetString(":tag")
	projectName, _ := utils.ParseRepository(repoName)
	exist, err := ra.ProjectMgr.Exist(projectName)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to check the existence of project %s: %v",
			projectName, err))
		return
	}
	if !exist {
		ra.HandleNotFound(fmt.Sprintf("project %s not found", projectName))
		return
	}
	if !ra.SecurityCtx.IsAuthenticated() {
		ra.HandleUnauthorized()
		return
	}
	if !ra.SecurityCtx.HasAllPerm(projectName) {
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}
	err = TriggerImageScan(repoName, tag)
	//TODO better check existence
	if err != nil {
		log.Errorf("Error while calling job service to trigger image scan: %v", err)
		ra.HandleInternalServerError("Failed to scan image, please check log for details")
		return
	}
}

func getSignatures(repository, username string) (map[string]*notary.Target, error) {
	targets, err := notary.GetInternalTargets(config.InternalNotaryEndpoint(),
		username, repository)
	if err != nil {
		return nil, err
	}

	signatures := map[string]*notary.Target{}
	for _, tgt := range targets {
		digest, err := notary.DigestFromTarget(tgt)
		if err != nil {
			return nil, err
		}
		signatures[digest] = &tgt
	}

	return signatures, nil
}

//will return nil when it failed to get data.  The parm "tag" is for logging only.
func getScanOverview(digest string, tag string) *models.ImgScanOverview {
	data, err := dao.GetImgScanOverview(digest)
	if err != nil {
		log.Errorf("Failed to get scan result for tag:%s, digest: %s, error: %v", tag, digest, err)
	}
	if data == nil {
		return nil
	}
	job, err := dao.GetScanJob(data.JobID)
	if err != nil {
		log.Errorf("Failed to get scan job for id:%d, error: %v", data.JobID, err)
		return nil
	} else if job == nil { //job does not exist
		log.Errorf("The scan job with id: %d does not exist, returning nil", data.JobID)
		return nil
	}
	data.Status = job.Status
	if data.Status != models.JobFinished {
		log.Debugf("Unsetting vulnerable related historical values, job status: %s", data.Status)
		data.Sev = 0
		data.CompOverview = nil
		data.DetailsKey = ""
	}
	return data
}
