// Copyright 2018 Project Harbor Authors
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
// These APIs provide services for manipulating Harbor project.

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/dghubble/sling"
	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/job/test"
	testutils "github.com/goharbor/harbor/src/common/utils/test"
	_ "github.com/goharbor/harbor/src/core/auth/db"
	_ "github.com/goharbor/harbor/src/core/auth/ldap"
	"github.com/goharbor/harbor/src/lib/config"
	libOrm "github.com/goharbor/harbor/src/lib/orm"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/orm"
	"github.com/goharbor/harbor/src/server/middleware/security"
	"github.com/goharbor/harbor/src/testing/apitests/apilib"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	TestUserName     = "testUser0001"
	TestUserPwd      = "testUser0001"
	jsonAcceptHeader = "application/json"
	testAcceptHeader = "text/plain"
	adminName        = "admin"
	adminPwd         = "Harbor12345"
)

var admin, unknownUsr, testUser *usrInfo
var handler http.Handler

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
	testutils.InitDatabaseFromEnv()
	config.Init()
	dao.PrepareTestData([]string{"delete from harbor_user where user_id >2", "delete from project where owner_id >2"}, []string{})
	config.Upload(testutils.GetUnitTestConfig())

	allCfgs, _ := config.GetSystemCfg(libOrm.Context())
	testutils.TraceCfgMap(allCfgs)

	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	dir = filepath.Join(dir, "..")
	apppath, _ := filepath.Abs(dir)
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.TestBeegoInit(apppath)

	beego.Router("/api/email/ping", &EmailAPI{}, "post:Ping")

	// Charts are controlled under projects
	chartRepositoryAPIType := &ChartRepositoryAPI{}
	beego.Router("/api/chartrepo/health", chartRepositoryAPIType, "get:GetHealthStatus")
	beego.Router("/api/chartrepo/:repo/charts", chartRepositoryAPIType, "get:ListCharts")
	beego.Router("/api/chartrepo/:repo/charts/:name", chartRepositoryAPIType, "get:ListChartVersions")
	beego.Router("/api/chartrepo/:repo/charts/:name", chartRepositoryAPIType, "delete:DeleteChart")
	beego.Router("/api/chartrepo/:repo/charts/:name/:version", chartRepositoryAPIType, "get:GetChartVersion")
	beego.Router("/api/chartrepo/:repo/charts/:name/:version", chartRepositoryAPIType, "delete:DeleteChartVersion")
	beego.Router("/api/chartrepo/:repo/charts", chartRepositoryAPIType, "post:UploadChartVersion")
	beego.Router("/api/chartrepo/:repo/prov", chartRepositoryAPIType, "post:UploadChartProvFile")
	beego.Router("/api/chartrepo/charts", chartRepositoryAPIType, "post:UploadChartVersion")

	// Repository services
	beego.Router("/chartrepo/:repo/index.yaml", chartRepositoryAPIType, "get:GetIndexByRepo")
	beego.Router("/chartrepo/index.yaml", chartRepositoryAPIType, "get:GetIndex")
	beego.Router("/chartrepo/:repo/charts/:filename", chartRepositoryAPIType, "get:DownloadChart")
	// Labels for chart
	chartLabelAPIType := &ChartLabelAPI{}
	beego.Router("/api/"+api.APIVersion+"/chartrepo/:repo/charts/:name/:version/labels", chartLabelAPIType, "get:GetLabels;post:MarkLabel")
	beego.Router("/api/"+api.APIVersion+"/chartrepo/:repo/charts/:name/:version/labels/:id([0-9]+)", chartLabelAPIType, "delete:RemoveLabel")

	beego.Router("/api/internal/syncquota", &InternalAPI{}, "post:SyncQuota")

	// Init user Info
	admin = &usrInfo{adminName, adminPwd}
	unknownUsr = &usrInfo{"unknown", "unknown"}
	testUser = &usrInfo{TestUserName, TestUserPwd}

	// Init mock jobservice
	mockServer := test.NewJobServiceServer()
	defer mockServer.Close()

	chain := middleware.Chain(orm.Middleware(), security.Middleware(), security.UnauthorizedMiddleware())
	handler = chain(beego.BeeApp.Handlers)
}

func request0(_sling *sling.Sling, acceptHeader string, authInfo ...usrInfo) (int, http.Header, []byte, error) {
	_sling = _sling.Set("Accept", acceptHeader)
	req, err := _sling.Request()
	if err != nil {
		return 400, nil, nil, err
	}
	if len(authInfo) > 0 {
		req.SetBasicAuth(authInfo[0].Name, authInfo[0].Passwd)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Body)
	return w.Code, w.Header(), body, err
}

func request(_sling *sling.Sling, acceptHeader string, authInfo ...usrInfo) (int, []byte, error) {
	code, _, body, err := request0(_sling, acceptHeader, authInfo...)
	return code, body, err
}

// Search for projects and repositories
// Implementation Notes
// The Search endpoint returns information about the projects and repositories
// offered at public status or related to the current logged in user.
// The response includes the project and repository list in a proper display order.
// @param q Search parameter for project and repository name.
// @return []Search
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

// Create a new project.
// Implementation Notes
// This endpoint is for user to create a new project.
// @param project New created project.
// @return void
// func (a testapi) ProjectsPost (prjUsr usrInfo, project apilib.Project) (int, error) {
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

// Delete project by projectID
func (a testapi) ProjectsDelete(prjUsr usrInfo, projectID string) (int, error) {
	_sling := sling.New().Delete(a.basePath)

	// create api path
	path := "api/projects/" + projectID
	_sling = _sling.Path(path)
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, prjUsr)
	return httpStatusCode, err
}

// Check if the project name user provided already exists
func (a testapi) ProjectsHead(prjUsr usrInfo, projectName string) (int, error) {
	_sling := sling.New().Head(a.basePath)

	// create api path
	path := "api/projects"
	_sling = _sling.Path(path)
	type QueryParams struct {
		ProjectName string `url:"project_name,omitempty"`
	}
	_sling = _sling.QueryStruct(&QueryParams{ProjectName: projectName})

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, prjUsr)
	return httpStatusCode, err
}

// Return specific project detail information
func (a testapi) ProjectsGetByPID(projectID string) (int, apilib.Project, error) {
	_sling := sling.New().Get(a.basePath)

	// create api path
	path := "api/projects/" + projectID
	_sling = _sling.Path(path)

	var successPayload apilib.Project

	httpStatusCode, body, err := request(_sling, jsonAcceptHeader)
	if err == nil && httpStatusCode == 200 {
		err = json.Unmarshal(body, &successPayload)
	}
	return httpStatusCode, successPayload, err
}

// Search projects by projectName and isPublic
func (a testapi) ProjectsGet(query *apilib.ProjectQuery, authInfo ...usrInfo) (int, []apilib.Project, error) {
	_sling := sling.New().Get(a.basePath).
		Path("api/projects").
		QueryStruct(query)

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
	} else {
		log.Println(string(body))
	}

	return httpStatusCode, successPayload, err
}

// Update properties for a selected project.
func (a testapi) ProjectsPut(prjUsr usrInfo, projectID string,
	project *proModels.Project) (int, error) {
	path := "/api/projects/" + projectID
	_sling := sling.New().Put(a.basePath).Path(path).BodyJSON(project)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, prjUsr)
	return httpStatusCode, err

}

// ProjectDeletable check whether a project can be deleted
func (a testapi) ProjectDeletable(prjUsr usrInfo, projectID int64) (int, bool, error) {
	_sling := sling.New().Get(a.basePath).
		Path("/api/projects/" + strconv.FormatInt(projectID, 10) + "/_deletable")

	code, body, err := request(_sling, jsonAcceptHeader, prjUsr)
	if err != nil {
		return 0, false, err
	}

	if code != http.StatusOK {
		return code, false, nil
	}

	deletable := struct {
		Deletable bool `json:"deletable"`
	}{}
	if err = json.Unmarshal(body, &deletable); err != nil {
		return 0, false, err
	}

	return code, deletable.Deletable, nil
}

// ProjectSummary returns summary for the project
func (a testapi) ProjectSummary(prjUsr usrInfo, projectID string) (int, apilib.ProjectSummary, error) {
	_sling := sling.New().Get(a.basePath)

	// create api path
	path := "api/projects/" + projectID + "/summary"
	_sling = _sling.Path(path)

	var successPayload apilib.ProjectSummary

	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, prjUsr)
	if err == nil && httpStatusCode == 200 {
		err = json.Unmarshal(body, &successPayload)
	}
	return httpStatusCode, successPayload, err
}

// -------------------------Member Test---------------------------------------//

// Return relevant role members of projectID
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

// Delete project role member accompany with  projectID
func (a testapi) DeleteProjectMember(authInfo usrInfo, projectID string, memberID string) (int, error) {
	_sling := sling.New().Delete(a.basePath)

	path := "/api/projects/" + projectID + "/members/" + memberID
	_sling = _sling.Path(path)
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err

}

// Get role memberInfo by projectId and UserId
func (a testapi) GetMemByPIDUID(authInfo usrInfo, projectID string, userID string) (int, error) {
	_sling := sling.New().Get(a.basePath)

	path := "/api/projects/" + projectID + "/members/" + userID

	_sling = _sling.Path(path)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

// Put:update current project role members accompany with relevant project and user
func (a testapi) PutProjectMember(authInfo usrInfo, projectID string, userID string, roles apilib.RoleParam) (int, error) {
	_sling := sling.New().Put(a.basePath)
	path := "/api/projects/" + projectID + "/members/" + userID

	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(roles)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

// --------------------Replication_Policy Test--------------------------------//

// Create a new replication policy
func (a testapi) AddPolicy(authInfo usrInfo, repPolicy apilib.RepPolicyPost) (int, error) {
	_sling := sling.New().Post(a.basePath)

	path := "/api/policies/replication/"

	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(repPolicy)

	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo)
	if httpStatusCode != http.StatusCreated {
		log.Println(string(body))
	}
	return httpStatusCode, err
}

// List policies by policyName and projectID
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

// Get replication policy by policyID
func (a testapi) GetPolicyByID(authInfo usrInfo, policyID string) (int, error) {
	_sling := sling.New().Get(a.basePath)

	path := "/api/policies/replication/" + policyID

	_sling = _sling.Path(path)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)

	return httpStatusCode, err
}

// Update policyInfo by policyID
func (a testapi) PutPolicyInfoByID(authInfo usrInfo, policyID string, policyUpdate apilib.RepPolicyUpdate) (int, error) {
	_sling := sling.New().Put(a.basePath)

	path := "/api/policies/replication/" + policyID

	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(policyUpdate)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

// Update policy enablement flag by policyID
func (a testapi) PutPolicyEnableByID(authInfo usrInfo, policyID string, policyEnable apilib.RepPolicyEnablementReq) (int, error) {
	_sling := sling.New().Put(a.basePath)

	path := "/api/policies/replication/" + policyID + "/enablement"

	_sling = _sling.Path(path)
	_sling = _sling.BodyJSON(policyEnable)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

// Delete policy by policyID
func (a testapi) DeletePolicyByID(authInfo usrInfo, policyID string) (int, error) {
	_sling := sling.New().Delete(a.basePath)

	path := "/api/policies/replication/" + policyID

	_sling = _sling.Path(path)

	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

// Get registered users of Harbor.
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

// Search registered users of Harbor.
func (a testapi) UsersSearch(userName string, authInfo ...usrInfo) (int, []apilib.UserSearch, error) {
	_sling := sling.New().Get(a.basePath)
	// create path and map variables
	path := "/api/users/search"
	_sling = _sling.Path(path)
	// body params
	type QueryParams struct {
		UserName string `url:"username, omitempty"`
	}
	_sling = _sling.QueryStruct(&QueryParams{UserName: userName})
	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo...)
	var successPayLoad []apilib.UserSearch
	if 200 == httpStatusCode && nil == err {
		err = json.Unmarshal(body, &successPayLoad)
	}
	return httpStatusCode, successPayLoad, err
}

// Get registered users by userid.
func (a testapi) UsersGetByID(userID int, authInfo usrInfo) (int, apilib.User, error) {
	_sling := sling.New().Get(a.basePath)
	// create path and map variables
	path := "/api/users/" + fmt.Sprintf("%d", userID)
	_sling = _sling.Path(path)
	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo)
	var successPayLoad apilib.User
	if 200 == httpStatusCode && nil == err {
		err = json.Unmarshal(body, &successPayLoad)
	}
	return httpStatusCode, successPayLoad, err
}

// Creates a new user account.
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

// Update a registered user to change profile.
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

// Update a registered user to be an administrator of Harbor.
func (a testapi) UsersToggleAdminRole(userID int, authInfo usrInfo, hasAdminRole bool) (int, error) {
	_sling := sling.New().Put(a.basePath)
	// create path and map variables
	path := "/api/users/" + fmt.Sprintf("%d", userID) + "/sysadmin"
	_sling = _sling.Path(path)
	type QueryParams struct {
		HasAdminRole bool `json:"sysadmin_flag,omitempty"`
	}

	_sling = _sling.BodyJSON(&QueryParams{HasAdminRole: hasAdminRole})
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

// Update password of a registered user.
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

func (a testapi) UsersGetPermissions(userID interface{}, scope string, authInfo usrInfo) (int, []apilib.Permission, error) {
	_sling := sling.New().Get(a.basePath)
	// create path and map variables
	path := fmt.Sprintf("/api/users/%v/permissions", userID)
	_sling = _sling.Path(path)
	type QueryParams struct {
		Scope string `url:"scope,omitempty"`
	}
	_sling = _sling.QueryStruct(&QueryParams{Scope: scope})
	httpStatusCode, body, err := request(_sling, jsonAcceptHeader, authInfo)
	var successPayLoad []apilib.Permission
	if 200 == httpStatusCode && nil == err {
		err = json.Unmarshal(body, &successPayLoad)
	}
	return httpStatusCode, successPayLoad, err
}

// Mark a registered user as be removed.
func (a testapi) UsersDelete(userID int, authInfo usrInfo) (int, error) {
	_sling := sling.New().Delete(a.basePath)
	// create path and map variables
	path := "/api/users/" + fmt.Sprintf("%d", userID)
	_sling = _sling.Path(path)
	httpStatusCode, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return httpStatusCode, err
}

func (a testapi) PingEmail(authInfo usrInfo, settings []byte) (int, string, error) {
	_sling := sling.New().Base(a.basePath).Post("/api/email/ping").Body(bytes.NewReader(settings))

	code, body, err := request(_sling, jsonAcceptHeader, authInfo)

	return code, string(body), err
}

func (a testapi) PostMeta(authInfor usrInfo, projectID int64, metas map[string]string) (int, string, error) {
	_sling := sling.New().Base(a.basePath).
		Post(fmt.Sprintf("/api/projects/%d/metadatas/", projectID)).
		BodyJSON(metas)

	code, body, err := request(_sling, jsonAcceptHeader, authInfor)
	return code, string(body), err
}

func (a testapi) PutMeta(authInfor usrInfo, projectID int64, name string,
	metas map[string]string) (int, string, error) {
	_sling := sling.New().Base(a.basePath).
		Put(fmt.Sprintf("/api/projects/%d/metadatas/%s", projectID, name)).
		BodyJSON(metas)

	code, body, err := request(_sling, jsonAcceptHeader, authInfor)
	return code, string(body), err
}

func (a testapi) GetMeta(authInfor usrInfo, projectID int64, name ...string) (int, map[string]string, error) {
	_sling := sling.New().Base(a.basePath).
		Get(fmt.Sprintf("/api/projects/%d/metadatas/", projectID))
	if len(name) > 0 {
		_sling = _sling.Path(name[0])
	}

	code, body, err := request(_sling, jsonAcceptHeader, authInfor)
	if err == nil && code == http.StatusOK {
		metas := map[string]string{}
		if err := json.Unmarshal(body, &metas); err != nil {
			return 0, nil, err
		}
		return code, metas, nil
	}
	return code, nil, err
}

func (a testapi) DeleteMeta(authInfor usrInfo, projectID int64, name string) (int, string, error) {
	_sling := sling.New().Base(a.basePath).
		Delete(fmt.Sprintf("/api/projects/%d/metadatas/%s", projectID, name))

	code, body, err := request(_sling, jsonAcceptHeader, authInfor)
	return code, string(body), err
}

type pingReq struct {
	ID             *int64  `json:"id"`
	Type           *string `json:"type"`
	URL            *string `json:"url"`
	CredentialType *string `json:"credential_type"`
	AccessKey      *string `json:"access_key"`
	AccessSecret   *string `json:"access_secret"`
	Insecure       *bool   `json:"insecure"`
}

func (a testapi) RegistryPing(authInfo usrInfo, registry *pingReq) (int, error) {
	_sling := sling.New().Base(a.basePath).Post("/api/registries/ping").BodyJSON(registry)
	code, _, err := request(_sling, jsonAcceptHeader, authInfo)
	return code, err
}

func (a testapi) RegistryDelete(authInfo usrInfo, registryID int64) (int, error) {
	_sling := sling.New().Base(a.basePath).Delete(fmt.Sprintf("/api/registries/%d", registryID))
	code, _, err := request(_sling, jsonAcceptHeader, authInfo)
	if err != nil || code != http.StatusOK {
		return code, fmt.Errorf("delete registry error: %v", err)
	}

	return code, nil
}
