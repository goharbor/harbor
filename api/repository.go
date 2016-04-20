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
	"strconv"
	"strings"
	"time"

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
	userID   int
	username string
	registry *registry.Registry
}

// Prepare will set a non existent user ID in case the request tries to view repositories under a project he doesn't has permission.
func (ra *RepositoryAPI) Prepare() {
	userID, ok := ra.GetSession("userId").(int)
	if !ok {
		ra.userID = dao.NonExistUserID
	} else {
		ra.userID = userID
	}
	username, ok := ra.GetSession("username").(string)
	if !ok {
		log.Warning("failed to get username from session")
		ra.username = ""
	} else {
		ra.username = username
	}

	var client *http.Client

	//no session, initialize a standard auth handler
	if ra.userID == dao.NonExistUserID && len(ra.username) == 0 {
		username, password, _ := ra.Ctx.Request.BasicAuth()

		credential := auth.NewBasicAuthCredential(username, password)
		client = registry.NewClientStandardAuthHandlerEmbeded(credential)
		log.Debug("initializing standard auth handler")

	} else {
		// session works, initialize a username auth handler
		username := ra.username
		if len(username) == 0 {
			user, err := dao.GetUser(models.User{
				UserID: ra.userID,
			})
			if err != nil {
				log.Errorf("error occurred whiling geting user for initializing a username auth handler: %v", err)
				return
			}

			username = user.Username
		}

		client = registry.NewClientUsernameAuthHandlerEmbeded(username)
		log.Debug("initializing username auth handler: %s", username)
	}

	endpoint := os.Getenv("REGISTRY_URL")
	r, err := registry.New(endpoint, client)
	if err != nil {
		log.Fatalf("error occurred while initializing auth handler for repository API: %v", err)
	}

	ra.registry = r
}

// Get ...
func (ra *RepositoryAPI) Get() {
	projectID, err0 := ra.GetInt64("project_id")
	if err0 != nil {
		log.Errorf("Failed to get project id, error: %v", err0)
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
	if p.Public == 0 && !checkProjectPermission(ra.userID, projectID) {
		ra.RenderError(http.StatusForbidden, "")
		return
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

	tags := []string{}
	tag := ra.GetString("tag")
	if len(tag) == 0 {
		tagList, err := ra.registry.ListTag(repoName)
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
		if err := ra.registry.DeleteTag(repoName, t); err != nil {
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

type histroyItem struct {
	V1Compatibility string `json:"v1Compatibility"`
}

type manifest struct {
	Name          string        `json:"name"`
	Tag           string        `json:"tag"`
	Architecture  string        `json:"architecture"`
	SchemaVersion int           `json:"schemaVersion"`
	History       []histroyItem `json:"history"`
}

// GetTags handles GET /api/repositories/tags
func (ra *RepositoryAPI) GetTags() {

	var tags []string

	repoName := ra.GetString("repo_name")

	tags, err := ra.registry.ListTag(repoName)
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

	ra.Data["json"] = tags
	ra.ServeJSON()
}

// GetManifests handles GET /api/repositories/manifests
func (ra *RepositoryAPI) GetManifests() {
	repoName := ra.GetString("repo_name")
	tag := ra.GetString("tag")

	item := models.RepoItem{}

	_, _, payload, err := ra.registry.PullManifest(repoName, tag, registry.ManifestVersion1)
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

	mani := manifest{}
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
