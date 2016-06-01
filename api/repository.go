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
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	svc_utils "github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils/log"
	"github.com/vmware/harbor/utils/registry"
	"github.com/vmware/harbor/utils/registry/auth"
	"github.com/vmware/harbor/utils/registry/errors"
)

// RepositoryAPI handles request to /api/repositories /api/repositories/tags /api/repositories/manifests, the parm has to be put
// in the query string as the web framework can not parse the URL if it contains veriadic sectors.
// For repostiories, we won't check the session in this API due to search functionality, querying manifest will be contorlled by
// the security of registry
type RepositoryAPI struct {
	BaseAPI
}

// Get ...
func (ra *RepositoryAPI) Get() {
	projectID, err := ra.GetInt64("project_id")
	if err != nil {
		log.Errorf("Failed to get project id, error: %v", err)
		ra.RenderError(http.StatusBadRequest, "Invalid project id")
		return
	}
	p, err := dao.GetProjectByID(projectID)
	if err != nil {
		log.Errorf("Error occurred in GetProjectById, error: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if p == nil {
		log.Warningf("Project with Id: %d does not exist", projectID)
		ra.RenderError(http.StatusNotFound, "")
		return
	}

	if p.Public == 0 {
		userID := ra.ValidateUser()

		if !checkProjectPermission(userID, projectID) {
			ra.RenderError(http.StatusForbidden, "")
			return
		}
	}

	repoList, err := svc_utils.GetRepoFromCache()
	if err != nil {
		log.Errorf("Failed to get repo from cache, error: %v", err)
		ra.RenderError(http.StatusInternalServerError, "internal sever error")
	}

	projectName := p.Name
	q := ra.GetString("q")
	var resp []string
	if len(q) > 0 {
		for _, r := range repoList {
			if strings.Contains(r, "/") && strings.Contains(r[strings.LastIndex(r, "/")+1:], q) && r[0:strings.LastIndex(r, "/")] == projectName {
				resp = append(resp, r)
			}
		}
		ra.Data["json"] = resp
	} else if len(projectName) > 0 {
		for _, r := range repoList {
			if strings.Contains(r, "/") && r[0:strings.LastIndex(r, "/")] == projectName {
				resp = append(resp, r)
			}
		}
		ra.Data["json"] = resp
	} else {
		ra.Data["json"] = repoList
	}
	ra.ServeJSON()
}

// Delete ...
func (ra *RepositoryAPI) Delete() {
	repoName := ra.GetString("repo_name")
	if len(repoName) == 0 {
		ra.CustomAbort(http.StatusBadRequest, "repo_name is nil")
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
			e, ok := errors.ParseError(err)
			if ok {
				log.Info(e)
				ra.CustomAbort(e.StatusCode, e.Message)
			} else {
				log.Error(err)
				ra.CustomAbort(http.StatusInternalServerError, "internal error")
			}
		}
		tags = append(tags, tagList...)
	} else {
		tags = append(tags, tag)
	}

	for _, t := range tags {
		if err := rc.DeleteTag(t); err != nil {
			e, ok := errors.ParseError(err)
			if ok {
				ra.CustomAbort(e.StatusCode, e.Message)
			} else {
				log.Error(err)
				ra.CustomAbort(http.StatusInternalServerError, "internal error")
			}
		}
		log.Infof("delete tag: %s %s", repoName, t)
	}

	go func() {
		log.Debug("refreshing catalog cache")
		if err := svc_utils.RefreshCatalogCache(); err != nil {
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

	rc, err := ra.initRepositoryClient(repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	tags := []string{}

	ts, err := rc.ListTag()
	if err != nil {
		e, ok := errors.ParseError(err)
		if ok {
			ra.CustomAbort(e.StatusCode, e.Message)
		} else {
			log.Error(err)
			ra.CustomAbort(http.StatusInternalServerError, "internal error")
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

	rc, err := ra.initRepositoryClient(repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	item := models.RepoItem{}

	mediaTypes := []string{schema1.MediaTypeManifest}
	_, _, payload, err := rc.PullManifest(tag, mediaTypes)
	if err != nil {
		e, ok := errors.ParseError(err)
		if ok {
			ra.CustomAbort(e.StatusCode, e.Message)
		} else {
			log.Error(err)
			ra.CustomAbort(http.StatusInternalServerError, "internal error")
		}
	}
	mani := models.Manifest{}
	err = json.Unmarshal(payload, &mani)
	if err != nil {
		log.Errorf("Failed to decode json from response for manifests, repo name: %s, tag: %s, error: %v", repoName, tag, err)
		ra.RenderError(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	v1Compatibility := mani.History[0].V1Compatibility

	err = json.Unmarshal([]byte(v1Compatibility), &item)
	if err != nil {
		log.Errorf("Failed to decode V1 field for repo, repo name: %s, tag: %s, error: %v", repoName, tag, err)
		ra.RenderError(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	item.DurationDays = strconv.Itoa(int(time.Since(item.Created).Hours()/24)) + " days"

	ra.Data["json"] = item
	ra.ServeJSON()
}

func (ra *RepositoryAPI) initRepositoryClient(repoName string) (r *registry.Repository, err error) {
	endpoint := os.Getenv("REGISTRY_URL")

	username, password, ok := ra.Ctx.Request.BasicAuth()
	if ok {
		credential := auth.NewBasicAuthCredential(username, password)
		return registry.NewRepositoryWithCredential(repoName, endpoint, credential)
	}

	username, err = ra.getUsername()
	if err != nil {
		return nil, err
	}

	return registry.NewRepositoryWithUsername(repoName, endpoint, username)
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

//GetTopTenRepos handles request GET /api/repositories/toprepos
func (ra *RepositoryAPI) GetTopTenRepos() {
	repos, err := dao.GetTop10Repos()
	if err != nil {
		log.Errorf("error occured in get top 10 repos: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "internal server error")
	}
	ra.Data["json"] = repos
	ra.ServeJSON()
}
