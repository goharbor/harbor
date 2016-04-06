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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/astaxie/beego"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
)

// RepositoryAPI handles request to /api/repositories /api/repositories/tags /api/repositories/manifests, the parm has to be put
// in the query string as the web framework can not parse the URL if it contains veriadic sectors.
// For repostiories, we won't check the session in this API due to search functionality, querying manifest will be contorlled by
// the security of registry
const (
	SryunUserNamePri = "sryci"
	RepoInfoDir      = "templates"
	ComposeJsonFile  = "sry_compose.json"
	ComposeYamlFile  = "sry_compose.yml"
)

type RepositoryV3API struct {
	BaseAPI
	userID          int
	username        string
	project_name    string
	repository_name string
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
		fmt.Println("username: ", ra.username)
	}

	// userId  from token
	if ra.userID == dao.NonExistUserID {
		ra.userID = ra.ValidateUser()
		ra.username = fmt.Sprintf("%s%d", SryunUserNamePri, ra.userID)
		fmt.Println("username: ", ra.username)
	}

	project_name := ra.Ctx.Input.Param(":project_name")
	if len(project_name) != 0 {
		ra.project_name = project_name
	}
	repository_name := ra.Ctx.Input.Param(":repository_name")
	if len(repository_name) != 0 {
		ra.repository_name = repository_name
	}
}

// GET /api/v3/repositories/{project_name}/{respository_name}
func (ra *RepositoryV3API) GetRepository() {
	if ra.project_name == "" || ra.repository_name == "" {
		beego.Error("required project_name and repository_name")
		ra.RenderError(http.StatusInternalServerError, "required project_name and repository_name")
	}
	repository, err := dao.GetRepositoryByName(fmt.Sprintf("%s/%s",
		ra.project_name, ra.repository_name))
	if err != nil {
		beego.Error("Failed to get repository from DB: ", err)
		ra.RenderError(http.StatusInternalServerError, "Failed to get repository")
	}
	markDown, _ := GetMarkDown(ra.project_name, ra.repository_name)
	sry_compose, _ := GetCompose(ra.project_name, ra.repository_name)
	log.Println("markdown: ", markDown)
	log.Println("compose: ", sry_compose)
	repository.MarkDown = markDown
	repository.SryCompose = sry_compose
	repositoryResponse := models.RepositoryResponse{
		Code: 0,
		Data: repository,
	}

	ra.Data["json"] = repositoryResponse
	ra.ServeJSON()
}

// GET /api/v3/repositories/mine
func (ra *RepositoryV3API) GetMineRepositories() {
	repositories, err := dao.RepositoriesUnderNamespace(ra.username)
	if err != nil {
		beego.Error("Failed to get repositories from DB: ", err)
		ra.RenderError(http.StatusInternalServerError, "Failed to get repositories")
	}
	ra.Data["json"] = repositories
	ra.ServeJSON()
}

// GET /api/v3/repositories
func (ra *RepositoryV3API) GetRepositories() {
	repositories, err := dao.RepositoriesUnderNamespace("library")
	if err != nil {
		beego.Error("Failed to get repositories from DB: ", err)
		ra.RenderError(http.StatusInternalServerError, "Failed to get repositories")
	}
	repositoriesResponse := models.RepositoriesResponse{
		Code: 0,
		Data: repositories,
	}
	ra.Data["json"] = repositoriesResponse
	ra.ServeJSON()
}
func (ra *RepositoryV3API) GetTags() {
	tags, err := dao.TagsUnderNamespaceAndRepo(fmt.Sprintf("%s/%s", ra.project_name, ra.repository_name))
	if err != nil {
		beego.Error("Failed to get repo tags, repo name:", ra.repository_name, ", error: ", err)
		ra.RenderError(http.StatusInternalServerError, "Failed to get repo tags")
	}
	tagsResponse := models.TagsResponse{
		Code: 0,
		Data: tags,
	}
	ra.Data["json"] = tagsResponse
	ra.ServeJSON()
}

// PUT /api/v3/repositories/{project_name}/{respository_name}
// update respository category
func (ra *RepositoryV3API) UpdateRepository() {

	if ra.project_name == "" || ra.repository_name == "" {
		beego.Error("Project name or repository name is blank")
		ra.RenderError(http.StatusBadRequest, "Project name or repositoryName is black")
	}
	var repo models.Repository
	fmt.Println(string(ra.Ctx.Input.CopyBody(1 << 32)))
	ra.DecodeJSONReq(&repo)
	//if repo.Category == "" {
	//	beego.Error("Failed to request can't be empty")
	//	ra.RenderError(http.StatusInternalServerError, "Failed to request cat't be empty")
	//}
	repository, _ := dao.RepositoryExists(fmt.Sprintf("%s/%s", ra.project_name, ra.repository_name))
	if repository == nil {
		beego.Error("Failed to get repository, project name: ", ra.project_name)
		ra.RenderError(http.StatusNotFound, "Failed to get repository")
	}
	if repo.Category != "" {
		repository.Category = repo.Category
	}
	if repo.Description != "" {
		repository.Description = repo.Description
	}
	if repo.IsPublic != 0 {
		repository.IsPublic = repo.IsPublic
	}
	repository, err := dao.UpdateRepoInfo(repository)
	if err != nil {
		beego.Error("Failed to update repository error: ", err)
		ra.RenderError(http.StatusInternalServerError, "Failed to update repository")
	}
	ra.Data["json"] = repository
	ra.ServeJSON()
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
	var categories []string
	for _, v := range strings.Split(string(b), "\n") {
		if len(v) > 0 {
			categories = append(categories, v)
		}
	}
	categoriesResponse := models.CategoriesResponse{
		Code: 0,
		Data: categories,
	}
	ra.Data["json"] = categoriesResponse
	ra.ServeJSON()
}

func GetMarkDown(project_name string, repository_name string) (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/%s/%s.%s", RepoInfoDir, project_name, repository_name, repository_name, "md"))
	log.Println(fmt.Sprintf("%s/%s/%s.%s", RepoInfoDir, project_name, repository_name, "md"))
	if err != nil {
		return "", err
	}
	b64 := GetBaseEncoding()
	repoMarkdown := b64.EncodeToString(b)
	return repoMarkdown, nil
}

func GetCompose(project_name string, repository_name string) (string, error) {
	b, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/%s/%s", RepoInfoDir, project_name, repository_name, ComposeJsonFile))
	fmt.Println(fmt.Sprintf("%s/%s/%s/%s", RepoInfoDir, project_name, repository_name, ComposeJsonFile))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func GetBaseEncoding() *base64.Encoding {
	return base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
}
