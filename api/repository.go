/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/service/cache"
	svc_utils "github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils/log"
	"github.com/vmware/harbor/utils/registry"

	registry_error "github.com/vmware/harbor/utils/registry/error"

	"github.com/vmware/harbor/utils"
	"github.com/vmware/harbor/utils/registry/auth"
)

// RepositoryAPI handles request to /api/repositories /api/repositories/tags /api/repositories/manifests, the parm has to be put
// in the query string as the web framework can not parse the URL if it contains veriadic sectors.
type RepositoryAPI struct {
	BaseAPI
}

// Get ...
func (ra *RepositoryAPI) Get() {
	projectID, err := ra.GetInt64("project_id")
	if err != nil || projectID <= 0 {
		ra.CustomAbort(http.StatusBadRequest, "invalid project_id")
	}

	page, pageSize := ra.getPaginationParams()

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

		if svc_utils.VerifySecret(ra.Ctx.Request) {
			userID = 1
		} else {
			userID = ra.ValidateUser()
		}

		if !checkProjectPermission(userID, projectID) {
			ra.CustomAbort(http.StatusForbidden, "")
		}
	}

	repositories, err := getReposByProject(project.Name, ra.GetString("q"))
	if err != nil {
		log.Errorf("failed to get repository: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}

	total := int64(len(repositories))

	if (page-1)*pageSize > total {
		repositories = []string{}
	} else {
		repositories = repositories[(page-1)*pageSize:]
	}

	if page*pageSize <= total {
		repositories = repositories[:pageSize]
	}

	ra.setPaginationHeader(total, page, pageSize)

	ra.Data["json"] = repositories
	ra.ServeJSON()
}

// Delete ...
func (ra *RepositoryAPI) Delete() {
	repoName := ra.GetString("repo_name")
	if len(repoName) == 0 {
		ra.CustomAbort(http.StatusBadRequest, "repo_name is nil")
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
		if !hasProjectAdminRole(userID, project.ProjectID) {
			ra.CustomAbort(http.StatusForbidden, "")
		}
	}

	rc, err := ra.initRepositoryClient(repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	tags := []string{}
	tag := ra.GetString("tag")
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

	for _, t := range tags {
		if err := rc.DeleteTag(t); err != nil {
			if regErr, ok := err.(*registry_error.Error); ok {
				if regErr.StatusCode != http.StatusNotFound {
					ra.CustomAbort(regErr.StatusCode, regErr.Detail)
				}
			} else {
				log.Errorf("error occurred while deleting tag %s:%s: %v", repoName, t, err)
				ra.CustomAbort(http.StatusInternalServerError, "internal error")
			}
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

	go func() {
		log.Debug("refreshing catalog cache")
		if err := cache.RefreshCatalogCache(); err != nil {
			log.Errorf("error occurred while refresh catalog cache: %v", err)
		}
	}()
}

type tag struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// GetTags handles GET /api/repositories/tags
func (ra *RepositoryAPI) GetTags() {
	repoName := ra.GetString("repo_name")
	if len(repoName) == 0 {
		ra.CustomAbort(http.StatusBadRequest, "repo_name is nil")
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

	tags := []string{}

	ts, err := rc.ListTag()
	if err != nil {
		regErr, ok := err.(*registry_error.Error)
		if !ok {
			log.Errorf("error occurred while listing tags of %s: %v", repoName, err)
			ra.CustomAbort(http.StatusInternalServerError, "internal error")
		}
		// TODO remove the logic if the bug of registry is fixed
		// It's a workaround for a bug of registry: when listing tags of
		// a repository which is being pushed, a "NAME_UNKNOWN" error will
		// been returned, while the catalog API can list this repository.
		if regErr.StatusCode != http.StatusNotFound {
			ra.CustomAbort(regErr.StatusCode, regErr.Detail)
		}
	}

	tags = append(tags, ts...)

	sort.Strings(tags)

	ra.Data["json"] = tags
	ra.ServeJSON()
}

// GetManifests handles GET /api/repositories/manifests
func (ra *RepositoryAPI) GetManifests() {
	repoName := ra.GetString("repo_name")
	tag := ra.GetString("tag")

	if len(repoName) == 0 || len(tag) == 0 {
		ra.CustomAbort(http.StatusBadRequest, "repo_name or tag is nil")
	}

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

	result := struct {
		Manifest interface{} `json:"manifest"`
		Config   interface{} `json:"config,omitempty" `
	}{}

	mediaTypes := []string{}
	switch version {
	case "v1":
		mediaTypes = append(mediaTypes, schema1.MediaTypeManifest)
	case "v2":
		mediaTypes = append(mediaTypes, schema2.MediaTypeManifest)
	}

	_, mediaType, payload, err := rc.PullManifest(tag, mediaTypes)
	if err != nil {
		if regErr, ok := err.(*registry_error.Error); ok {
			ra.CustomAbort(regErr.StatusCode, regErr.Detail)
		}

		log.Errorf("error occurred while getting manifest of %s:%s: %v", repoName, tag, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	manifest, _, err := registry.UnMarshal(mediaType, payload)
	if err != nil {
		log.Errorf("an error occurred while parsing manifest of %s:%s: %v", repoName, tag, err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}

	result.Manifest = manifest

	deserializedmanifest, ok := manifest.(*schema2.DeserializedManifest)
	if ok {
		_, data, err := rc.PullBlob(deserializedmanifest.Target().Digest.String())
		if err != nil {
			log.Errorf("failed to get config of manifest %s:%s: %v", repoName, tag, err)
			ra.CustomAbort(http.StatusInternalServerError, "")
		}

		b, err := ioutil.ReadAll(data)
		if err != nil {
			log.Errorf("failed to read config of manifest %s:%s: %v", repoName, tag, err)
			ra.CustomAbort(http.StatusInternalServerError, "")
		}

		result.Config = string(b)
	}

	ra.Data["json"] = result
	ra.ServeJSON()
}

func (ra *RepositoryAPI) initRepositoryClient(repoName string) (r *registry.Repository, err error) {
	endpoint := os.Getenv("REGISTRY_URL")

	username, password, ok := ra.Ctx.Request.BasicAuth()
	if ok {
		return newRepositoryClient(endpoint, getIsInsecure(), username, password,
			repoName, "repository", repoName, "pull", "push", "*")
	}

	username, err = ra.getUsername()
	if err != nil {
		return nil, err
	}

	return cache.NewRepositoryClient(endpoint, getIsInsecure(), username, repoName,
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

//GetTopRepos handles request GET /api/repositories/top
func (ra *RepositoryAPI) GetTopRepos() {
	count, err := ra.GetInt("count", 10)
	if err != nil || count <= 0 {
		ra.CustomAbort(http.StatusBadRequest, "invalid count")
	}

	repos, err := dao.GetTopRepos(count)
	if err != nil {
		log.Errorf("failed to get top repos: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "internal server error")
	}
	ra.Data["json"] = repos
	ra.ServeJSON()
}

func newRepositoryClient(endpoint string, insecure bool, username, password, repository, scopeType, scopeName string,
	scopeActions ...string) (*registry.Repository, error) {

	credential := auth.NewBasicAuthCredential(username, password)
	authorizer := auth.NewStandardTokenAuthorizer(credential, insecure, scopeType, scopeName, scopeActions...)

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
