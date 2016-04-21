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
	"strconv"
	"strings"
	"time"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	svc_utils "github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils/log"
)

// RepositoryAPI handles request to /api/repositories /api/repositories/tags /api/repositories/manifests, the parm has to be put
// in the query string as the web framework can not parse the URL if it contains veriadic sectors.
// For repostiories, we won't check the session in this API due to search functionality, querying manifest will be contorlled by
// the security of registry
type RepositoryAPI struct {
	BaseAPI
	userID   int
	username string
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

type tag struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// GetTags handles GET /api/repositories/tags
func (ra *RepositoryAPI) GetTags() {

	var tags []string

	repoName := ra.GetString("repo_name")
	result, err := svc_utils.RegistryAPIGet(svc_utils.BuildRegistryURL(repoName, "tags", "list"), ra.username)
	if err != nil {
		log.Errorf("Failed to get repo tags, repo name: %s, error: %v", repoName, err)
		ra.RenderError(http.StatusInternalServerError, "Failed to get repo tags")
	} else {
		t := tag{}
		json.Unmarshal(result, &t)
		tags = t.Tags
	}
	ra.Data["json"] = tags
	ra.ServeJSON()
}

// GetManifests handles GET /api/repositories/manifests
func (ra *RepositoryAPI) GetManifests() {
	repoName := ra.GetString("repo_name")
	tag := ra.GetString("tag")

	item := models.RepoItem{}

	result, err := svc_utils.RegistryAPIGet(svc_utils.BuildRegistryURL(repoName, "manifests", tag), ra.username)
	if err != nil {
		log.Errorf("Failed to get manifests for repo, repo name: %s, tag: %s, error: %v", repoName, tag, err)
		ra.RenderError(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	mani := models.Manifest{}
	err = json.Unmarshal(result, &mani)
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
	item.CreatedStr = item.Created.Format("2006-01-02 15:04:05")
	item.DurationDays = strconv.Itoa(int(time.Since(item.Created).Hours()/24)) + " days"

	ra.Data["json"] = item
	ra.ServeJSON()
}
