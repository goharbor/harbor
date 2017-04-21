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
	"path"
	"sort"
	"strings"
	"time"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/vmware/harbor/src/common/api"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/notary"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	registry_error "github.com/vmware/harbor/src/common/utils/registry/error"
	"github.com/vmware/harbor/src/ui/config"
	svc_utils "github.com/vmware/harbor/src/ui/service/utils"
)

// RepositoryAPI handles request to /api/repositories /api/repositories/tags /api/repositories/manifests, the parm has to be put
// in the query string as the web framework can not parse the URL if it contains veriadic sectors.
type RepositoryAPI struct {
	api.BaseAPI
}

type repoResp struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	OwnerID      int64     `json:"owner_id"`
	ProjectID    int64     `json:"project_id"`
	Description  string    `json:"description"`
	PullCount    int64     `json:"pull_count"`
	StarCount    int64     `json:"star_count"`
	TagsCount    int64     `json:"tags_count"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}

type detailedTagResp struct {
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
		ra.CustomAbort(http.StatusBadRequest, "invalid project_id")
	}

	project, err := dao.GetProjectByID(projectID)
	if err != nil {
		log.Errorf("failed to get project %d: %v", projectID, err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}

	if project == nil {
		ra.CustomAbort(http.StatusNotFound, fmt.Sprintf("project %d not found", projectID))
	}

	if project.Public == 0 {
		var userID int

		if svc_utils.VerifySecret(ra.Ctx.Request, config.JobserviceSecret()) {
			userID = 1
		} else {
			userID = ra.ValidateUser()
		}

		if !checkProjectPermission(userID, projectID) {
			ra.CustomAbort(http.StatusForbidden, "")
		}
	}

	keyword := ra.GetString("q")

	total, err := dao.GetTotalOfRepositoriesByProject(projectID, keyword)
	if err != nil {
		log.Errorf("failed to get total of repositories of project %d: %v", projectID, err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}

	page, pageSize := ra.GetPaginationParams()

	detail := ra.GetString("detail") == "1" || ra.GetString("detail") == "true"

	repositories, err := getRepositories(projectID,
		keyword, pageSize, pageSize*(page-1), detail)
	if err != nil {
		log.Errorf("failed to get repository: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}

	ra.SetPaginationHeader(total, page, pageSize)
	ra.Data["json"] = repositories
	ra.ServeJSON()
}

func getRepositories(projectID int64, keyword string,
	limit, offset int64, detail bool) (interface{}, error) {
	repositories, err := dao.GetRepositoriesByProject(projectID, keyword, limit, offset)
	if err != nil {
		return nil, err
	}

	//keep compatibility with old API
	if !detail {
		result := []string{}
		for _, repository := range repositories {
			result = append(result, repository.Name)
		}
		return result, nil
	}

	return populateTagsCount(repositories)
}

func populateTagsCount(repositories []*models.RepoRecord) ([]*repoResp, error) {
	result := []*repoResp{}
	for _, repository := range repositories {
		repo := &repoResp{
			ID:           repository.RepositoryID,
			Name:         repository.Name,
			OwnerID:      repository.OwnerID,
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
	project, err := dao.GetProjectByName(projectName)
	if err != nil {
		log.Errorf("failed to get project %s: %v", projectName, err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}

	if project == nil {
		ra.CustomAbort(http.StatusNotFound, fmt.Sprintf("project %s not found", projectName))
	}

	userID := ra.ValidateUser()
	if !hasProjectAdminRole(userID, project.ProjectID) {
		ra.CustomAbort(http.StatusForbidden, "")
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

	user, _, ok := ra.Ctx.Request.BasicAuth()
	if !ok {
		user, err = ra.getUsername()
		if err != nil {
			log.Errorf("failed to get user: %v", err)
		}
	}

	if config.WithNotary() {
		var digest string
		signedTags := make(map[string]struct{})
		targets, err := getNotaryTargets(user, repoName)
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
			if _, ok = signedTags[digest]; ok {
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
			if err := dao.AccessLog(user, projectName, repoName, tag, "delete"); err != nil {
				log.Errorf("failed to add access log: %v", err)
			}
		}(t)
	}

	exist, err := repositoryExist(repoName, rc)
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

type tag struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// GetTags returns tags of a repository
func (ra *RepositoryAPI) GetTags() {
	repoName := ra.GetString(":splat")
	detail := ra.GetString("detail") == "1" || ra.GetString("detail") == "true"

	projectName, _ := utils.ParseRepository(repoName)
	project, err := dao.GetProjectByName(projectName)
	if err != nil {
		log.Errorf("failed to get project %s: %v", projectName, err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}

	if project == nil {
		ra.CustomAbort(http.StatusNotFound, fmt.Sprintf("project %s not found", projectName))
	}

	if project.Public == 0 {
		userID := ra.ValidateUser()
		if !checkProjectPermission(userID, project.ProjectID) {
			ra.CustomAbort(http.StatusForbidden, "")
		}
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

	if !detail {
		ra.Data["json"] = tags
		ra.ServeJSON()
		return
	}

	result := []detailedTagResp{}

	for _, tag := range tags {
		manifest, err := getManifest(client, tag, "v1")
		if err != nil {
			if regErr, ok := err.(*registry_error.Error); ok {
				ra.CustomAbort(regErr.StatusCode, regErr.Detail)
			}

			log.Errorf("failed to get manifest of %s:%s: %v", repoName, tag, err)
			ra.CustomAbort(http.StatusInternalServerError, "internal error")
		}

		result = append(result, detailedTagResp{
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
	project, err := dao.GetProjectByName(projectName)
	if err != nil {
		log.Errorf("failed to get project %s: %v", projectName, err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}

	if project == nil {
		ra.CustomAbort(http.StatusNotFound, fmt.Sprintf("project %s not found", projectName))
	}

	if project.Public == 0 {
		userID := ra.ValidateUser()
		if !checkProjectPermission(userID, project.ProjectID) {
			ra.CustomAbort(http.StatusForbidden, "")
		}
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

	verify, err := config.VerifyRemoteCert()
	if err != nil {
		return nil, err
	}

	username, password, ok := ra.Ctx.Request.BasicAuth()
	if ok {
		return newRepositoryClient(endpoint, !verify, username, password,
			repoName, "repository", repoName, "pull", "push", "*")
	}

	username, err = ra.getUsername()
	if err != nil {
		return nil, err
	}

	return NewRepositoryClient(endpoint, !verify, username, repoName,
		"repository", repoName, "pull", "push", "*")
}

func (ra *RepositoryAPI) getUsername() (string, error) {
	// get username from session
	sessionUsername := ra.GetSession("username")
	if sessionUsername != nil {
		username, ok := sessionUsername.(string)
		if ok {
			return username, nil
		}
	}

	// if username does not exist in session, try to get userId from sessiion
	// and then get username from DB according to the userId
	sessionUserID := ra.GetSession("userId")
	if sessionUserID != nil {
		userID, ok := sessionUserID.(int)
		if ok {
			u := models.User{
				UserID: userID,
			}
			user, err := dao.GetUser(u)
			if err != nil {
				return "", err
			}

			return user.Username, nil
		}
	}

	return "", nil
}

//GetTopRepos returns the most populor repositories
func (ra *RepositoryAPI) GetTopRepos() {
	count, err := ra.GetInt("count", 10)
	if err != nil || count <= 0 {
		ra.CustomAbort(http.StatusBadRequest, "invalid count")
	}

	userID, _, ok := ra.GetUserIDForRequest()
	if !ok {
		userID = dao.NonExistUserID
	}

	repos, err := dao.GetTopRepos(userID, count)
	if err != nil {
		log.Errorf("failed to get top repos: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "internal server error")
	}

	detail := ra.GetString("detail") == "1" || ra.GetString("detail") == "true"
	if !detail {
		result := []*models.TopRepo{}

		for _, repo := range repos {
			result = append(result, &models.TopRepo{
				RepoName:    repo.Name,
				AccessCount: repo.PullCount,
			})
		}

		ra.Data["json"] = result
		ra.ServeJSON()
		return
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
	//use this func to init session.
	ra.GetUserIDForRequest()
	username, err := ra.getUsername()
	if err != nil {
		log.Warningf("Error when getting username: %v", err)
	}
	repoName := ra.GetString(":splat")

	targets, err := getNotaryTargets(username, repoName)
	if err != nil {
		log.Errorf("Error while fetching signature from notary: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}
	ra.Data["json"] = targets
	ra.ServeJSON()
}

func getNotaryTargets(username string, repo string) ([]notary.Target, error) {
	ext, err := config.ExtEndpoint()
	if err != nil {
		log.Errorf("Error while reading external endpoint: %v", err)
		return nil, err
	}
	endpoint := strings.Split(ext, "//")[1]
	fqRepo := path.Join(endpoint, repo)
	return notary.GetTargets(config.InternalNotaryEndpoint(), username, fqRepo)
}

func newRepositoryClient(endpoint string, insecure bool, username, password, repository, scopeType, scopeName string,
	scopeActions ...string) (*registry.Repository, error) {

	credential := auth.NewBasicAuthCredential(username, password)

	authorizer := auth.NewStandardTokenAuthorizer(credential, insecure,
		config.InternalTokenServiceEndpoint(), scopeType, scopeName, scopeActions...)

	store, err := auth.NewAuthorizerStore(endpoint, insecure, authorizer)
	if err != nil {
		return nil, err
	}

	client, err := registry.NewRepositoryWithModifiers(repository, endpoint, insecure, store)
	if err != nil {
		return nil, err
	}
	return client, nil
}
