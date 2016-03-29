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
	"io/ioutil"
	"log"
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
}

// GET /api/v3/repositories/{project_name}/{respository_name}
func (ra *RepositoryV3API) GetRepository() {
}

// PUT /api/v3/repositories/{project_name}/{respository_name}
//
// update respository category
func (ra *RepositoryV3API) UpdateRepository() {
	projectName := ra.Ctx.Input.Param(":project_name")
	//respositoryName := ra.Ctx.Input.Param(":respository_name")
	if projectName == "" {
		beego.Error("Project name is blank")
		ra.CustomAbort(http.StatusBadRequest, "Project name is blank")
	}
	var repo models.Repository
	err := json.Unmarshal(ra.Ctx.Input.RequestBody, &repo)
	if err != nil {
		beego.Error("Failed to request body conver to json err: ", err)
		ra.RenderError(http.StatusInternalServerError, "Failed to request body conver to json")
	}
	repository, err := dao.GetRepositoryByName(projectName)
	if err != nil {
		beego.Error("Failed to get repository, project name: ", projectName, ", error: ", err)
		ra.RenderError(http.StatusInternalServerError, "Failed to get repository")
	}
	repository.Category = repo.Category
	repository.Description = repo.Description
	/*err = dao.UpdateRepository(repository)
	if err != nil {
		beego.Error("Failed to update repository error: ", err)
		ra.RenderError(http.StatusInternalServerError, "Failed to update repository")
	}
	jstr, _ := json.Marshal(repository)
	ra.Data["json"] = jstr
	ra.ServerJSON()*/
}

// PUT /api/v3/repositories/categories
//
// return list of repository categories, category are stored in /path/to/project/root/CATEGORIES
func (ra *RepositoryV3API) GetCategories() {
	b, err := ioutil.ReadFile("CATEGORIES")
	if err != nil {
		beego.Error("Ftailed to get CATEGORIES errors: ", err)
		ra.RenderError(http.StatusInternalServerError, "Failed to get repo CATEGORIES")
	}
	var categorys []string
	for _, v := range strings.Split(string(b), "\n") {
		if len(v) > 0 {
			categorys = append(categorys, v)
		}
	}
	ra.Data["json"] = categorys
	ra.ServeJSON()
}
