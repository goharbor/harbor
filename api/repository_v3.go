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
	"strings"

	"github.com/vmware/harbor/dao"
	svc_utils "github.com/vmware/harbor/service/utils"

	"github.com/astaxie/beego"
)

// RepositoryAPI handles request to /api/repositories /api/repositories/tags /api/repositories/manifests, the parm has to be put
// in the query string as the web framework can not parse the URL if it contains veriadic sectors.
// For repostiories, we won't check the session in this API due to search functionality, querying manifest will be contorlled by
// the security of registry
type RepositoryV3API struct {
	BaseAPI
	userID   int
	username string
}

// Prepare will set a non existent user ID in case the request tries to view repositories under a project he doesn't has permission.
func (ra *RepositoryV3API) Prepare() {
	userID, ok := ra.GetSession("userId").(int)
	if !ok {
		ra.userID = dao.NonExistUserID
	} else {
		ra.userID = userID
	}
	username, ok := ra.GetSession("username").(string)
	if !ok {
		beego.Warning("failed to get username from session")
		ra.username = ""
	} else {
		ra.username = username
	}
}

// Get ...
func (ra *RepositoryV3API) Get() {
	projectID, err0 := ra.GetInt64("project_id")
	if err0 != nil {
		beego.Error("Failed to get project id, error:", err0)
		ra.RenderError(http.StatusBadRequest, "Invalid project id")
		return
	}
	p, err := dao.GetProjectByID(projectID)
	if err != nil {
		beego.Error("Error occurred in GetProjectById:", err)
		ra.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if p == nil {
		beego.Warning("Project with Id:", projectID, ", does not exist")
		ra.RenderError(http.StatusNotFound, "")
		return
	}
	if p.Public == 0 && !checkProjectPermission(ra.userID, projectID) {
		ra.RenderError(http.StatusForbidden, "")
		return
	}
	repoList, err := svc_utils.GetRepoFromCache()
	if err != nil {
		beego.Error("Failed to get repo from cache, error:", err)
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

func (ra *RepositoryV3API) GetTags() {

	var tags []string

	repoName := ra.GetString("repo_name")
	result, err := svc_utils.RegistryAPIGet(svc_utils.BuildRegistryURL(repoName, "tags", "list"), ra.username)
	if err != nil {
		beego.Error("Failed to get repo tags, repo name:", repoName, ", error: ", err)
		ra.RenderError(http.StatusInternalServerError, "Failed to get repo tags")
	} else {
		t := tag{}
		json.Unmarshal(result, &t)
		tags = t.Tags
	}
	ra.Data["json"] = tags
	ra.ServeJSON()
}

// GET /api/v3/repositories
func (ra *RepositoryV3API) GetRepositories() {
	queryRepository := model.Repository{}
	repositories, err := dao.GetRepository(queryReposetory)
	if err != nil {
		beego.Error("Failed to get repositories from DB")
		beego.RenderError(http.StatusInternalServerError, "Failed to get repositories")
	}
	ra.Data["json"] = repositories
	ra.ServeJSON()
}

// GET /api/v3/repositories/{project_name}/{respository_name}
func (ra *RepositoryV3API) GetRepository() {
	repository_name := ra.GetString("respository_name", "")
	repositroy, err := dao.GetRepositoryByName(repositoryName)
	if err != nil {
		beego.Error("Failed to get repository from DB")
		beego.RenderError(http.StatusInternalServerError, "Failed to get repository")
	}
	ra.Data["json"] = Repository
	ra.ServeJSON()
}

// PUT /api/v3/repositories/{project_name}/{respository_name}
//
// update respository category
func (ra *RepositoryV3API) UpdateRepository() {
}

// PUT /api/v3/repositories/categories
//
// return list of repository categories, category are stored in /path/to/project/root/CATEGORIES
func (ra *RepositoryV3API) GetCategories() {
}
