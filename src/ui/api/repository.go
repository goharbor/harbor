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
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/notary"
	"github.com/vmware/harbor/src/common/utils/registry"
	registry_error "github.com/vmware/harbor/src/common/utils/registry/error"
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

type tagResp struct {
	Tag      string      `json:"tag"`
	Manifest interface{} `json:"manifest"`
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

	total, err := dao.GetTotalOfRepositoriesByProject(projectID, keyword)
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
		var digest string
		signedTags := make(map[string]struct{})
		targets, err := notary.GetInternalTargets(config.InternalNotaryEndpoint(),
			ra.SecurityCtx.GetUsername(), repoName)
		if err != nil {
			log.Errorf("Failed to get Notary targets for repository: %s, error: %v", repoName, err)
			log.Warningf("Failed to check signature status of repository: %s for deletion, there maybe orphaned targets in Notary.", repoName)
		}
		for _, tgt := range targets {
			digest, err = notary.DigestFromTarget(tgt)
			if err != nil {
				log.Errorf("Failed to get disgest from target, error: %v", err)
				ra.CustomAbort(http.StatusInternalServerError, err.Error())
			}
			signedTags[digest] = struct{}{}
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
				log.Error("project %s not found", projectName)
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
	tags, err := listTag(client)
	if err != nil {
		regErr, ok := err.(*registry_error.Error)
		if !ok {
			log.Errorf("error occurred while listing tags of %s: %v", repoName, err)
			ra.CustomAbort(http.StatusInternalServerError, "internal error")
		}

		ra.CustomAbort(regErr.StatusCode, regErr.Detail)
	}

	result := []tagResp{}

	for _, tag := range tags {
		manifest, err := getManifest(client, tag, "v1")
		if err != nil {
			if regErr, ok := err.(*registry_error.Error); ok {
				ra.CustomAbort(regErr.StatusCode, regErr.Detail)
			}

			log.Errorf("failed to get manifest of %s:%s: %v", repoName, tag, err)
			ra.CustomAbort(http.StatusInternalServerError, "internal error")
		}

		result = append(result, tagResp{
			Tag:      tag,
			Manifest: manifest.Manifest,
		})
	}

	ra.Data["json"] = result
	ra.ServeJSON()
}

func listTag(client *registry.Repository) ([]string, error) {
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
