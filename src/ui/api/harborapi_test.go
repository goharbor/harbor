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
//These APIs provide services for manipulating Harbor project.

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/tests/apitests/apilib"
	//	"strconv"
	//	"strings"

	"github.com/astaxie/beego"
	"github.com/dghubble/sling"

	//for test env prepare
	_ "github.com/vmware/harbor/src/ui/auth/db"
	_ "github.com/vmware/harbor/src/ui/auth/ldap"
)

const (
	jsonAcceptHeader = "application/json"
	testAcceptHeader = "text/plain"
	adminName        = "admin"
	adminPwd         = "Harbor12345"
)

var admin, unknownUsr, testUser *usrInfo

type testapi struct {
	basePath string
}

func newHarborAPI() *testapi {
	return &testapi{
		basePath: "",
	}
}

func newHarborAPIWithBasePath(basePath string) *testapi {
	return &testapi{
		basePath: basePath,
	}
}

type usrInfo struct {
	Name   string
	Passwd string
}

func init() {
	if err := config.Init(); err != nil {
		log.Fatalf("failed to initialize configurations: %v", err)
	}
	database, err := config.Database()
	if err != nil {
		log.Fatalf("failed to get database configurations: %v", err)
	}
	dao.InitDatabase(database)
	_, file, _, _ := runtime.Caller(1)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".."+string(filepath.Separator))))
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.TestBeegoInit(apppath)

	beego.Router("/api/search/", &SearchAPI{})
	beego.Router("/api/projects/", &ProjectAPI{}, "get:List;post:Post;head:Head")
	beego.Router("/api/projects/:id", &ProjectAPI{}, "delete:Delete;get:Get")
	beego.Router("/api/users/?:id", &UserAPI{})
	beego.Router("/api/users/:id([0-9]+)/password", &UserAPI{}, "put:ChangePassword")
	beego.Router("/api/users/:id/sysadmin", &UserAPI{}, "put:ToggleUserAdminRole")
	beego.Router("/api/projects/:id/publicity", &ProjectAPI{}, "put:ToggleProjectPublic")
	beego.Router("/api/projects/:id([0-9]+)/logs/filter", &ProjectAPI{}, "post:FilterAccessLog")
	beego.Router("/api/projects/:pid([0-9]+)/members/?:mid", &ProjectMemberAPI{}, "get:Get;post:Post;delete:Delete;put:Put")
	beego.Router("/api/repositories", &RepositoryAPI{})
	beego.Router("/api/statistics", &StatisticAPI{})
	beego.Router("/api/users/?:id", &UserAPI{})
	beego.Router("/api/logs", &LogAPI{})
	beego.Router("/api/repositories/*/tags/?:tag", &RepositoryAPI{}, "delete:Delete")
	beego.Router("/api/repositories/*/tags", &RepositoryAPI{}, "get:GetTags")
	beego.Router("/api/repositories/*/tags/:tag/manifest", &RepositoryAPI{}, "get:GetManifests")
	beego.Router("/api/repositories/*/signatures", &RepositoryAPI{}, "get:GetSignatures")
	beego.Router("/api/repositories/top", &RepositoryAPI{}, "get:GetTopRepos")
	beego.Router("/api/targets/", &TargetAPI{}, "get:List")
	beego.Router("/api/targets/", &TargetAPI{}, "post:Post")
	beego.Router("/api/targets/:id([0-9]+)", &TargetAPI{})
	beego.Router("/api/targets/:id([0-9]+)/policies/", &TargetAPI{}, "get:ListPolicies")
	beego.Router("/api/targets/ping", &TargetAPI{}, "post:Ping")
	beego.Router("/api/targets/:id([0-9]+)/ping", &TargetAPI{}, "post:PingByID")
	beego.Router("/api/policies/replication/:id([0-9]+)", &RepPolicyAPI{})
	beego.Router("/api/policies/replication", &RepPolicyAPI{}, "get:List")
	beego.Router("/api/policies/replication", &RepPolicyAPI{}, "post:Post;delete:Delete")
	beego.Router("/api/policies/replication/:id([0-9]+)/enablement", &RepPolicyAPI{}, "put:UpdateEnablement")
	beego.Router("/api/systeminfo", &SystemInfoAPI{}, "get:GetGeneralInfo")
	beego.Router("/api/systeminfo/volumes", &SystemInfoAPI{}, "get:GetVolumeInfo")
	beego.Router("/api/systeminfo/getcert", &SystemInfoAPI{}, "get:GetCert")
	beego.Router("/api/ldap/ping", &LdapAPI{}, "post:Ping")
	beego.Router("/api/configurations", &ConfigAPI{})
	beego.Router("/api/configurations/reset", &ConfigAPI{}, "post:Reset")
	beego.Router("/api/email/ping", &EmailAPI{}, "post:Ping")

	_ = updateInitPassword(1, "Harbor12345")

	//syncRegistry
	if err := SyncRegistry(); err != nil {
		log.Fatalf("failed to sync repositories from registry: %v", err)
	}

	//Init user Info
	admin = &usrInfo{adminName, adminPwd}
	unknownUsr = &usrInfo{"unknown", "unknown"}
	testUser = &usrInfo{TestUserName, TestUserPwd}

}

func request(_sling *sling.Sling, acceptHeader string, authInfo ...usrInfo) (int, []byte, error) {
	_sling = _sling.Set("Accept", acceptHeader)
	req, err := _sling.Request()
	if err != nil {
		return 400, nil, err
	}
	if len(authInfo) > 0 {
		req.SetBasicAuth(authInfo[0].Name, authInfo[0].Passwd)
	}
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, req)
	body, err := ioutil.ReadAll(w.Body)
	return w.Code, body, err
}

//Search for projects and repositories
//Implementation Notes
//The Search endpoint returns information about the projects and repositories
//offered at public status or related to the current logged in user.
//The response includes the project and repository list in a proper display order.
//@param q Search parameter for project and repository name.
//@return []Search
func (a testapi) SearchGet(q string, authInfo ...usrInfo) (int, apilib.Search, error) {
	var httpCode int
	var body []byte
	var err error

	_sling := sling.New().Get(a.basePath)

	// create path and map variables
	path := "/api/search"
	_sling = _sling.Path(path)

	type QueryParams struct {
		Query string `url:"q,omitempty"`
	}

	_sling = _sling.QueryStruct(&QueryParams{Query: q})

	if len(authInfo) > 0 {
		httpCode, body, err = request(_sling, jsonAcceptHeader, authInfo[0])
	} else {
		httpCode, body, err = request(_sling, jsonAcceptHeader)
	}

	var successPayload = new(apilib.Search)
	err = json.Unmarshal(body, &successPayload)
	return httpCode, *successPayload, err
}

//Create a new project.
//Implementation Notes
//This endpoint is for user to create a new project.
//@param project New created project.
//@return void
//func (a testapi) ProjectsPost (prjUsr usrInfo, project apilib.Project) (int, error) {
func (a testapi) ProjectsPost(prjUsr usrInfo, project apilib.ProjectReq) (int, error) {

	_sling := sling.New().Post(a.basePath)

	// create path and map variables
	path := "/api/projects/"

	_sling = _sling.Path(path)

	// body params
	_sling = _sling.BodyJSON(project)
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, prjUsr)
	return httpStatusCode, err
}

func (a testapi) StatisticGet(user usrInfo) (int, apilib.StatisticMap, error) {
	_sling := sling.New().Get(a.basePath)

	// create path and map variables
	path := "/api/statistics/"

	_sling = _sling.Path(path)
	var successPayload apilib.StatisticMap
	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, user)

	if err == nil && httpStatusCode == 200 {
		err = json.Unmarshal(body, &successPayload)
	}
	return httpStatusCode, successPayload, err
}

func (a testapi) LogGet(user usrInfo, startTime, endTime, lines string) (int, []apilib.AccessLog, error) {
	_sling := sling.New().Get(a.basePath)

	// create path and map variables
	path := "/api/logs/"
	fmt.Printf("logs path: %s\n", path)
	_sling = _sling.Path(path)
	type QueryParams struct {
		StartTime string `url:"start_time,omitempty"`
		EndTime   string `url:"end_time,omitempty"`
		Lines     string `url:"lines,omitempty"`
	}

	_sling = _sling.QueryStruct(&QueryParams{StartTime: startTime, EndTime: endTime, Lines: lines})
	var successPayload []apilib.AccessLog
	code, body, err := request(_sling, jsonAcceptHeader, user)
	if 200 == code && nil == err {
		err = json.Unmarshal(body, &successPayload)
	}
	return code, successPayload, err
}

////Delete a repository or a tag in a repository.
////Delete a repository or a tag in a repository.
////This endpoint let user delete repositories and tags with repo name and tag.\n
////@param repoName The name of repository which will be deleted.
////@param tag Tag of a repository.
////@return void
////func (a testapi) RepositoriesDelete(prjUsr UsrInfo, repoName string, tag string) (int, error) {
//func (a testapi) RepositoriesDelete(prjUsr UsrInfo, repoName string, tag string) (int, error) {
//	_sling := sling.New().Delete(a.basePath)

//	// create path and map variables
//	path := "/api/repositories"

//	_sling = _sling.Path(path)

//	type QueryParams struct {
//		RepoName string `url:"repo_name,omitempty"`
//		Tag      string `url:"tag,omitempty"`
//	}

//	_sling = _sling.QueryStruct(&QueryParams{RepoName: repoName, Tag: tag})
//	// accept header
//	accepts := []string{"application/json", "text/plain"}
//	for key := range accepts {
//		_sling = _sling.Set("Accept", accepts[key])
//		break // only use the first Accept
//	}

//	req, err := _sling.Request()
//	req.SetBasicAuth(prjUsr.Name, prjUsr.Passwd)
//	//fmt.Printf("request %+v", req)

//	client := &http.Client{}
//	httpResponse, err := client.Do(req)
//	defer httpResponse.Body.Close()

//	if err != nil {
//		// handle error
//	}
//	return httpResponse.StatusCode, err
//}

//Delete project by projectID
func (a testapi) ProjectsDelete(prjUsr usrInfo, projectID string) (int, error) {
	_sling := sling.New().Delete(a.basePath)

	//create api path
	path := "api/projects/" + projectID
	_sling = _sling.Path(path)
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, prjUsr)
	return httpStatusCode, err
}

//Check if the project name user provided already exists
func (a testapi) ProjectsHead(prjUsr usrInfo, projectName string) (int, error) {
	_sling := sling.New().Head(a.basePath)

	//create api path
	path := "api/projects"
	_sling = _sling.Path(path)
	type QueryParams struct {
		ProjectName string `url:"project_name,omitempty"`
	}
	_sling = _sling.QueryStruct(&QueryParams{ProjectName: projectName})

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, prjUsr)
	return httpStatusCode, err
}

//Return specific project detail infomation
func (a testapi) ProjectsGetByPID(projectID string) (int, apilib.Project, error) {
	_sling := sling.New().Get(a.basePath)

	//create api path
	path := "api/projects/" + projectID
	_sling = _sling.Path(path)

	var successPayload apilib.Project

	httpStatusCode, body, err := request(_sling, jsonAcceptHeader)
	if err == nil && httpStatusCode == 200 {
		err = json.Unmarshal(body, &successPayload)
	}
	return httpStatusCode, successPayload, err
}

//Search projects by projectName and isPublic
func (a testapi) ProjectsGet(projectName string, isPublic int32, authInfo ...usrInfo) (int, []apilib.Project, error) {
	_sling := sling.New().Get(a.basePath)

	//create api path
	path := "api/projects"
	_sling = _sling.Path(path)
	type QueryParams struct {
		ProjectName string `url:"project_name,omitempty"`
		IsPubilc    int32  `url:"is_public,omitempty"`
	}
	_sling = _sling.QueryStruct(&QueryParams{ProjectName: projectName, IsPubilc: isPublic})

	var successPayload []apilib.Project

	var httpStatusCode int
	var err error
	var body []byte
	if len(authInfo) > 0 {
		httpStatusCode, body, err = request(_sling, jsonAcceptHeader, authInfo[0])
	} else {
		httpStatusCode, body, err = request(_sling, jsonAcceptHeader)
	}

	if err == nil && httpStatusCode == 200 {
		err = json.Unmarshal(body, &successPayload)
	}

	return httpStatusCode, successPayload, err
}

//Update properties for a selected project.
func (a testapi) ToggleProjectPublicity(prjUsr usrInfo, projectID string, ispublic int32) (int, error) {
	// create path and map variables
	path := "/api/projects/" + projectID + "/publicity/"
	_sling := sling.New().Put(a.basePath)

	_sling = _sling.Path(path)

	type QueryParams struct {
		Public int32 `json:"public,omitempty"`
	}

	_sling = _sling.BodyJSON(&QueryParams{Public: ispublic})

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, prjUsr)
	return httpStatusCode, err

}

//Get access logs accompany with a relevant project.
func (a testapi) ProjectLogsFilter(prjUsr usrInfo, projectID string, accessLog apilib.AccessLogFilter) (int, []byte, error) {
	//func (a testapi) ProjectLogsFilter(prjUsr usrInfo, projectID string, accessLog apilib.AccessLog) (int, apilib.AccessLog, error) {
	_sling := sling.New().Post(a.basePath)

	path := "/api/projects/" + projectID + "/logs/filter"

	_sling = _sling.Path(path)

	// body params
	_sling = _sling.BodyJSON(accessLog)

	//var successPayload []apilib.AccessLog

	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, prjUsr)
	/*
		if err == nil && httpStatusCode == 200 {
			err = json.Unmarshal(body, &successPayload)
		}
	*/
	return httpStatusCode, body, err
	//	return httpStatusCode, successPayload, err
}

//-------------------------Member Test---------------------------------------//

//Return relevant role members of projectID
func (a testapi) GetProjectMembersByProID(prjUsr usrInfo, projectID string) (int, []apilib.User, error) {
	_sling := sling.New().Get(a.basePath)

	path := "/api/projects/" + projectID + "/members/"

	_sling = _sling.Path(path)

	var successPayload []apilib.User

	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, prjUsr)
	if err == nil && httpStatusCode == 200 {
		err = json.Unmarshal(body, &successPayload)
	}
	return httpStatusCode, successPayload, err
}

//Add project role member accompany with  projectID
func (a testapi) AddProjectMember(prjUsr usrInfo, projectID string, roles apilib.RoleParam) (int, error) {
	_sling := sling.New().Post(a.basePath)

	path := "/api/projects/" + projectID + "/members/"
	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(roles)
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, prjUsr)
	return httpStatusCode, err

}

//Delete project role member accompany with  projectID
func (a testapi) DeleteProjectMember(authInfo usrInfo, projectID string, userID string) (int, error) {
	_sling := sling.New().Delete(a.basePath)

	path := "/api/projects/" + projectID + "/members/" + userID
	_sling = _sling.Path(path)
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err

}

//Get role memberInfo by projectId and UserId
func (a testapi) GetMemByPIDUID(authInfo usrInfo, projectID string, userID string) (int, error) {
	_sling := sling.New().Get(a.basePath)

	path := "/api/projects/" + projectID + "/members/" + userID

	_sling = _sling.Path(path)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//Put:update current project role members accompany with relevant project and user
func (a testapi) PutProjectMember(authInfo usrInfo, projectID string, userID string, roles apilib.RoleParam) (int, error) {
	_sling := sling.New().Put(a.basePath)
	path := "/api/projects/" + projectID + "/members/" + userID

	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(roles)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//-------------------------Repositories Test---------------------------------------//
//Return relevant repos of projectID
func (a testapi) GetRepos(authInfo usrInfo, projectID,
	keyword, detail string) (int, interface{}, error) {
	_sling := sling.New().Get(a.basePath)

	path := "/api/repositories/"

	_sling = _sling.Path(path)

	type QueryParams struct {
		ProjectID string `url:"project_id"`
		Detail    string `url:"detail"`
		Keyword   string `url:"q"`
	}

	_sling = _sling.QueryStruct(&QueryParams{
		ProjectID: projectID,
		Detail:    detail,
		Keyword:   keyword,
	})
	code, body, err := request(_sling, jsonAcceptHeader, authInfo)
	if err != nil {
		return 0, nil, err
	}

	if code == http.StatusOK {
		if detail == "1" || detail == "true" {
			repositories := []repoResp{}
			if err = json.Unmarshal(body, &repositories); err != nil {
				return 0, nil, err
			}
			return code, repositories, nil
		}

		repositories := []string{}
		if err = json.Unmarshal(body, &repositories); err != nil {
			return 0, nil, err
		}
		return code, repositories, nil
	}

	return code, nil, nil
}

//Get tags of a relevant repository
func (a testapi) GetReposTags(authInfo usrInfo, repoName,
	detail string) (int, interface{}, error) {
	_sling := sling.New().Get(a.basePath)

	path := fmt.Sprintf("/api/repositories/%s/tags", repoName)

	_sling = _sling.Path(path)

	type QueryParams struct {
		Detail string `url:"detail"`
	}

	_sling = _sling.QueryStruct(&QueryParams{
		Detail: detail,
	})
	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo)
	if err != nil {
		return 0, nil, err
	}

	if httpStatusCode != http.StatusOK {
		return httpStatusCode, body, nil
	}

	if detail == "true" || detail == "1" {
		result := []detailedTagResp{}
		if err := json.Unmarshal(body, &result); err != nil {
			return 0, nil, err
		}
		return http.StatusOK, result, nil
	}

	result := []string{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, nil, err
	}
	return http.StatusOK, result, nil
}

//Get manifests of a relevant repository
func (a testapi) GetReposManifests(authInfo usrInfo, repoName string, tag string) (int, error) {
	_sling := sling.New().Get(a.basePath)

	path := fmt.Sprintf("/api/repositories/%s/tags/%s/manifest", repoName, tag)

	_sling = _sling.Path(path)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//Get public repositories which are accessed most
func (a testapi) GetReposTop(authInfo usrInfo, count,
	detail string) (int, interface{}, error) {
	_sling := sling.New().Get(a.basePath)

	path := "/api/repositories/top"

	_sling = _sling.Path(path)

	type QueryParams struct {
		Count  string `url:"count"`
		Detail string `url:"detail"`
	}

	_sling = _sling.QueryStruct(&QueryParams{
		Count:  count,
		Detail: detail,
	})
	code, body, err := request(_sling, jsonAcceptHeader, authInfo)
	if err != nil {
		return 0, nil, err
	}

	if code != http.StatusOK {
		return code, body, err
	}

	if detail == "true" || detail == "1" {
		result := []*repoResp{}
		if err = json.Unmarshal(body, &result); err != nil {
			return 0, nil, err
		}
		return http.StatusOK, result, nil
	}

	result := []*models.TopRepo{}
	if err = json.Unmarshal(body, &result); err != nil {
		return 0, nil, err
	}
	return http.StatusOK, result, nil
}

//-------------------------Targets Test---------------------------------------//
//Create a new replication target
func (a testapi) AddTargets(authInfo usrInfo, repTarget apilib.RepTargetPost) (int, string, error) {
	_sling := sling.New().Post(a.basePath)

	path := "/api/targets"

	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(repTarget)

	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, string(body), err
}

//List filters targets by name
func (a testapi) ListTargets(authInfo usrInfo, targetName string) (int, []apilib.RepTarget, error) {
	_sling := sling.New().Get(a.basePath)

	path := "/api/targets?name=" + targetName

	_sling = _sling.Path(path)

	var successPayload []apilib.RepTarget

	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo)
	if err == nil && httpStatusCode == 200 {
		err = json.Unmarshal(body, &successPayload)
	}

	return httpStatusCode, successPayload, err
}

//Ping target
func (a testapi) PingTarget(authInfo usrInfo, body interface{}) (int, error) {
	_sling := sling.New().Post(a.basePath)

	path := "/api/targets/ping"

	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(body)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//PingTargetByID ...
func (a testapi) PingTargetByID(authInfo usrInfo, id int) (int, error) {
	_sling := sling.New().Post(a.basePath)

	path := fmt.Sprintf("/api/targets/%d/ping", id)

	_sling = _sling.Path(path)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//Get target by targetID
func (a testapi) GetTargetByID(authInfo usrInfo, targetID string) (int, error) {
	_sling := sling.New().Get(a.basePath)

	path := "/api/targets/" + targetID

	_sling = _sling.Path(path)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)

	return httpStatusCode, err
}

//Update target by targetID
func (a testapi) PutTargetByID(authInfo usrInfo, targetID string, repTarget apilib.RepTargetPost) (int, error) {
	_sling := sling.New().Put(a.basePath)

	path := "/api/targets/" + targetID

	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(repTarget)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)

	return httpStatusCode, err
}

//List the target relevant policies by targetID
func (a testapi) GetTargetPoliciesByID(authInfo usrInfo, targetID string) (int, error) {
	_sling := sling.New().Get(a.basePath)

	path := "/api/targets/" + targetID + "/policies/"

	_sling = _sling.Path(path)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)

	return httpStatusCode, err
}

//Delete target by targetID
func (a testapi) DeleteTargetsByID(authInfo usrInfo, targetID string) (int, error) {
	_sling := sling.New().Delete(a.basePath)

	path := "/api/targets/" + targetID

	_sling = _sling.Path(path)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//--------------------Replication_Policy Test--------------------------------//

//Create a new replication policy
func (a testapi) AddPolicy(authInfo usrInfo, repPolicy apilib.RepPolicyPost) (int, error) {
	_sling := sling.New().Post(a.basePath)

	path := "/api/policies/replication/"

	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(repPolicy)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//List policies by policyName and projectID
func (a testapi) ListPolicies(authInfo usrInfo, policyName string, proID string) (int, []apilib.RepPolicy, error) {
	_sling := sling.New().Get(a.basePath)

	path := "/api/policies/replication/"

	_sling = _sling.Path(path)

	type QueryParams struct {
		PolicyName string `url:"name"`
		ProjectID  string `url:"project_id"`
	}
	_sling = _sling.QueryStruct(&QueryParams{PolicyName: policyName, ProjectID: proID})

	var successPayload []apilib.RepPolicy

	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo)

	if err == nil && httpStatusCode == 200 {
		err = json.Unmarshal(body, &successPayload)
	}
	return httpStatusCode, successPayload, err
}

//Get replication policy by policyID
func (a testapi) GetPolicyByID(authInfo usrInfo, policyID string) (int, error) {
	_sling := sling.New().Get(a.basePath)

	path := "/api/policies/replication/" + policyID

	_sling = _sling.Path(path)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)

	return httpStatusCode, err
}

//Update policyInfo by policyID
func (a testapi) PutPolicyInfoByID(authInfo usrInfo, policyID string, policyUpdate apilib.RepPolicyUpdate) (int, error) {
	_sling := sling.New().Put(a.basePath)

	path := "/api/policies/replication/" + policyID

	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(policyUpdate)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//Update policy enablement flag by policyID
func (a testapi) PutPolicyEnableByID(authInfo usrInfo, policyID string, policyEnable apilib.RepPolicyEnablementReq) (int, error) {
	_sling := sling.New().Put(a.basePath)

	path := "/api/policies/replication/" + policyID + "/enablement"

	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(policyEnable)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//Delete policy by policyID
func (a testapi) DeletePolicyByID(authInfo usrInfo, policyID string) (int, error) {
	_sling := sling.New().Delete(a.basePath)

	path := "/api/policies/replication/" + policyID

	_sling = _sling.Path(path)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//Return projects created by Harbor
//func (a HarborApi) ProjectsGet (projectName string, isPublic int32) ([]Project, error) {
//    }

//Check if the project name user provided already exists.
//func (a HarborApi) ProjectsHead (projectName string) (error) {
//}

//Get access logs accompany with a relevant project.
//func (a HarborApi) ProjectsProjectIdLogsFilterPost (projectID int32, accessLog AccessLog) ([]AccessLog, error) {
//}

//Return a project&#39;s relevant role members.
//func (a HarborApi) ProjectsProjectIdMembersGet (projectID int32) ([]Role, error) {
//}

//Add project role member accompany with relevant project and user.
//func (a HarborApi) ProjectsProjectIdMembersPost (projectID int32, roles RoleParam) (error) {
//}

//Delete project role members accompany with relevant project and user.
//func (a HarborApi) ProjectsProjectIdMembersUserIdDelete (projectID int32, userId int32) (error) {
//}

//Return role members accompany with relevant project and user.
//func (a HarborApi) ProjectsProjectIdMembersUserIdGet (projectID int32, userId int32) ([]Role, error) {
//}

//Update project role members accompany with relevant project and user.
//func (a HarborApi) ProjectsProjectIdMembersUserIdPut (projectID int32, userId int32, roles RoleParam) (error) {
//}

//Update properties for a selected project.
//func (a HarborApi) ProjectsProjectIdPut (projectID int32, project Project) (error) {
//}

//Get repositories accompany with relevant project and repo name.
//func (a HarborApi) RepositoriesGet (projectID int32, q string) ([]Repository, error) {
//}

//Get manifests of a relevant repository.
//func (a HarborApi) RepositoriesManifestGet (repoName string, tag string) (error) {
//}

//Get tags of a relevant repository.
//func (a HarborApi) RepositoriesTagsGet (repoName string) (error) {
//}

//Get registered users of Harbor.
func (a testapi) UsersGet(userName string, authInfo usrInfo) (int, []apilib.User, error) {
	_sling := sling.New().Get(a.basePath)
	// create path and map variables
	path := "/api/users/"
	_sling = _sling.Path(path)
	// body params
	type QueryParams struct {
		UserName string `url:"username, omitempty"`
	}
	_sling = _sling.QueryStruct(&QueryParams{UserName: userName})
	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo)
	var successPayLoad []apilib.User
	if 200 == httpStatusCode && nil == err {
		err = json.Unmarshal(body, &successPayLoad)
	}
	return httpStatusCode, successPayLoad, err
}

//Get registered users by userid.
func (a testapi) UsersGetByID(userName string, authInfo usrInfo, userID int) (int, apilib.User, error) {
	_sling := sling.New().Get(a.basePath)
	// create path and map variables
	path := "/api/users/" + fmt.Sprintf("%d", userID)
	_sling = _sling.Path(path)
	// body params
	type QueryParams struct {
		UserName string `url:"username, omitempty"`
	}
	_sling = _sling.QueryStruct(&QueryParams{UserName: userName})
	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo)
	var successPayLoad apilib.User
	if 200 == httpStatusCode && nil == err {
		err = json.Unmarshal(body, &successPayLoad)
	}
	return httpStatusCode, successPayLoad, err
}

//Creates a new user account.
func (a testapi) UsersPost(user apilib.User, authInfo ...usrInfo) (int, error) {
	_sling := sling.New().Post(a.basePath)

	// create path and map variables
	path := "/api/users/"

	_sling = _sling.Path(path)

	// body params
	_sling = _sling.BodyJSON(user)
	var httpStatusCode int
	var err error
	if len(authInfo) > 0 {
		httpStatusCode, _, err = request(_sling, jsonAcceptHeader, authInfo[0])
	} else {
		httpStatusCode, _, err = request(_sling, jsonAcceptHeader)
	}
	return httpStatusCode, err

}

//Update a registered user to change profile.
func (a testapi) UsersPut(userID int, profile apilib.UserProfile, authInfo usrInfo) (int, error) {
	_sling := sling.New().Put(a.basePath)
	// create path and map variables
	path := "/api/users/" + fmt.Sprintf("%d", userID)
	_sling = _sling.Path(path)

	// body params
	_sling = _sling.BodyJSON(profile)
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//Update a registered user to be an administrator of Harbor.
func (a testapi) UsersToggleAdminRole(userID int, authInfo usrInfo, hasAdminRole int32) (int, error) {
	_sling := sling.New().Put(a.basePath)
	// create path and map variables
	path := "/api/users/" + fmt.Sprintf("%d", userID) + "/sysadmin"
	_sling = _sling.Path(path)
	type QueryParams struct {
		HasAdminRole int32 `json:"has_admin_role,omitempty"`
	}

	_sling = _sling.BodyJSON(&QueryParams{HasAdminRole: hasAdminRole})
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//Update password of a registered user.
func (a testapi) UsersUpdatePassword(userID int, password apilib.Password, authInfo usrInfo) (int, error) {
	_sling := sling.New().Put(a.basePath)
	// create path and map variables
	path := "/api/users/" + fmt.Sprintf("%d", userID) + "/password"
	_sling = _sling.Path(path)
	// body params
	_sling = _sling.BodyJSON(password)
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

//Mark a registered user as be removed.
func (a testapi) UsersDelete(userID int, authInfo usrInfo) (int, error) {
	_sling := sling.New().Delete(a.basePath)
	// create path and map variables
	path := "/api/users/" + fmt.Sprintf("%d", userID)
	_sling = _sling.Path(path)
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}
func updateInitPassword(userID int, password string) error {
	queryUser := models.User{UserID: userID}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		return fmt.Errorf("Failed to get user, userID: %d %v", userID, err)
	}
	if user == nil {
		return fmt.Errorf("user id: %d does not exist", userID)
	}
	if user.Salt == "" {
		user.Salt = utils.GenerateRandomString()
		user.Password = password
		err = dao.ChangeUserPassword(*user)
		if err != nil {
			return fmt.Errorf("Failed to update user encrypted password, userID: %d, err: %v", userID, err)
		}

	} else {
	}
	return nil
}

//Get system volume info
func (a testapi) VolumeInfoGet(authInfo usrInfo) (int, apilib.SystemInfo, error) {
	_sling := sling.New().Get(a.basePath)
	path := "/api/systeminfo/volumes"
	_sling = _sling.Path(path)
	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo)
	var successPayLoad apilib.SystemInfo
	if 200 == httpStatusCode && nil == err {
		err = json.Unmarshal(body, &successPayLoad)
	}

	return httpStatusCode, successPayLoad, err
}

func (a testapi) GetGeneralInfo() (int, []byte, error) {
	_sling := sling.New().Get(a.basePath).Path("/api/systeminfo")
	return request(_sling, jsonAcceptHeader)
}

//Get system cert
func (a testapi) CertGet(authInfo usrInfo) (int, []byte, error) {
	_sling := sling.New().Get(a.basePath)
	path := "/api/systeminfo/getcert"
	_sling = _sling.Path(path)
	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, body, err
}

//Post ldap test
func (a testapi) LdapPost(authInfo usrInfo, ldapConf apilib.LdapConf) (int, error) {

	_sling := sling.New().Post(a.basePath)

	// create path and map variables
	path := "/api/ldap/ping"

	_sling = _sling.Path(path)

	// body params
	_sling = _sling.BodyJSON(ldapConf)
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

func (a testapi) GetConfig(authInfo usrInfo) (int, map[string]*value, error) {
	_sling := sling.New().Base(a.basePath).Get("/api/configurations")

	cfg := map[string]*value{}

	code, body, err := request(_sling, jsonAcceptHeader, authInfo)
	if err == nil && code == 200 {
		err = json.Unmarshal(body, &cfg)
	}
	return code, cfg, err
}

func (a testapi) PutConfig(authInfo usrInfo, cfg map[string]string) (int, error) {
	_sling := sling.New().Base(a.basePath).Put("/api/configurations").BodyJSON(cfg)

	code, _, err := request(_sling, jsonAcceptHeader, authInfo)

	return code, err
}

func (a testapi) ResetConfig(authInfo usrInfo) (int, error) {
	_sling := sling.New().Base(a.basePath).Post("/api/configurations/reset")

	code, _, err := request(_sling, jsonAcceptHeader, authInfo)

	return code, err
}

func (a testapi) PingEmail(authInfo usrInfo, settings []byte) (int, string, error) {
	_sling := sling.New().Base(a.basePath).Post("/api/email/ping").Body(bytes.NewReader(settings))

	code, body, err := request(_sling, jsonAcceptHeader, authInfo)

	return code, string(body), err
}
