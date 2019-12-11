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
package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common"

	"github.com/astaxie/beego"
	"github.com/dghubble/sling"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/project"
	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	nonSysAdminID, projAdminID, projDeveloperID, projGuestID, projAdminRobotID int64
	projAdminPMID, projDeveloperPMID, projGuestPMID, projAdminRobotPMID        int
	// The following users/credentials are registered and assigned roles at the beginning of
	// running testing and cleaned up at the end.
	// Do not try to change the system and project roles that the users have during
	// the testing. Creating a new one in your own case if needed.
	// The project roles that the users have are for project library.
	sysAdmin = &usrInfo{
		Name:   "admin",
		Passwd: "Harbor12345",
	}
	nonSysAdmin = &usrInfo{
		Name:   "non_admin",
		Passwd: "Harbor12345",
	}
	projAdmin = &usrInfo{
		Name:   "proj_admin",
		Passwd: "Harbor12345",
	}
	projDeveloper = &usrInfo{
		Name:   "proj_developer",
		Passwd: "Harbor12345",
	}
	projGuest = &usrInfo{
		Name:   "proj_guest",
		Passwd: "Harbor12345",
	}
	projAdmin4Robot = &usrInfo{
		Name:   "proj_admin_robot",
		Passwd: "Harbor12345",
	}
)

type testingRequest struct {
	method      string
	url         string
	header      http.Header
	queryStruct interface{}
	bodyJSON    interface{}
	credential  *usrInfo
}

type codeCheckingCase struct {
	request  *testingRequest
	code     int
	postFunc func(*httptest.ResponseRecorder) error
}

func newRequest(r *testingRequest) (*http.Request, error) {
	if r == nil {
		return nil, nil
	}

	reqBuilder := sling.New()
	switch strings.ToUpper(r.method) {
	case "", http.MethodGet:
		reqBuilder = reqBuilder.Get(r.url)
	case http.MethodPost:
		reqBuilder = reqBuilder.Post(r.url)
	case http.MethodPut:
		reqBuilder = reqBuilder.Put(r.url)
	case http.MethodDelete:
		reqBuilder = reqBuilder.Delete(r.url)
	case http.MethodHead:
		reqBuilder = reqBuilder.Head(r.url)
	case http.MethodPatch:
		reqBuilder = reqBuilder.Patch(r.url)
	default:
		return nil, fmt.Errorf("unsupported method %s", r.method)
	}

	for key, values := range r.header {
		for _, value := range values {
			reqBuilder = reqBuilder.Add(key, value)
		}
	}

	if r.queryStruct != nil {
		reqBuilder = reqBuilder.QueryStruct(r.queryStruct)
	}

	if r.bodyJSON != nil {
		reqBuilder = reqBuilder.BodyJSON(r.bodyJSON)
	}

	if r.credential != nil {
		reqBuilder = reqBuilder.SetBasicAuth(r.credential.Name, r.credential.Passwd)
	}

	return reqBuilder.Request()
}

func handle(r *testingRequest) (*httptest.ResponseRecorder, error) {
	req, err := newRequest(r)
	if err != nil {
		return nil, err
	}

	resp := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(resp, req)
	return resp, nil
}

func handleAndParse(r *testingRequest, v interface{}) error {
	resp, err := handle(r)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.Code >= 200 && resp.Code <= 299 {
		return json.Unmarshal(data, v)
	}

	return &common_http.Error{
		Code:    resp.Code,
		Message: string(data),
	}
}

func runCodeCheckingCases(t *testing.T, cases ...*codeCheckingCase) {
	for i, c := range cases {
		t.Logf("running case %d ...", i)
		resp, err := handle(c.request)
		require.Nil(t, err)
		equal := assert.Equal(t, c.code, resp.Code)
		if !equal {
			if resp.Body.Len() > 0 {
				t.Log(resp.Body.String())
			}
			t.FailNow()
		}

		if c.postFunc != nil {
			if err := c.postFunc(resp); err != nil {
				t.Logf("error in running post function: %v", err)
				t.Error(err)
			}
		}
	}
}

func parseResourceID(resp *httptest.ResponseRecorder) (int64, error) {
	location := resp.Header().Get(http.CanonicalHeaderKey("location"))
	if len(location) == 0 {
		return 0, fmt.Errorf("empty location header")
	}
	index := strings.LastIndex(location, "/")
	if index == -1 {
		return 0, fmt.Errorf("location header %s contains no /", location)
	}

	id := strings.TrimPrefix(location, location[:index+1])
	if len(id) == 0 {
		return 0, fmt.Errorf("location header %s contains no resource ID", location)
	}

	return strconv.ParseInt(id, 10, 64)
}

func TestMain(m *testing.M) {
	if err := prepare(); err != nil {
		panic(err)
	}
	dao.ExecuteBatchSQL([]string{
		"insert into user_group (group_name, group_type, ldap_group_dn) values ('test_group_01_api', 1, 'cn=harbor_users,ou=sample,ou=vmware,dc=harbor,dc=com')",
		"insert into user_group (group_name, group_type, ldap_group_dn) values ('vsphere.local\\administrators', 2, '')",
	})

	defer dao.ExecuteBatchSQL([]string{
		"delete from harbor_label",
		"delete from robot",
		"delete from user_group",
		"delete from project_member",
	})

	ret := m.Run()
	clean()
	os.Exit(ret)
}

func prepare() error {
	// register nonSysAdmin
	var err error
	nonSysAdminID, err = dao.Register(models.User{
		Username: nonSysAdmin.Name,
		Password: nonSysAdmin.Passwd,
		Email:    nonSysAdmin.Name + "@test.com",
	})
	if err != nil {
		return err
	}

	// register projAdmin and assign project admin role
	projAdminID, err = dao.Register(models.User{
		Username: projAdmin.Name,
		Password: projAdmin.Passwd,
		Email:    projAdmin.Name + "@test.com",
	})
	if err != nil {
		return err
	}

	if projAdminPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  1,
		Role:       models.PROJECTADMIN,
		EntityID:   int(projAdminID),
		EntityType: common.UserMember,
	}); err != nil {
		return err
	}

	// register projAdminRobots and assign project admin role
	projAdminRobotID, err = dao.Register(models.User{
		Username: projAdmin4Robot.Name,
		Password: projAdmin4Robot.Passwd,
		Email:    projAdmin4Robot.Name + "@test.com",
	})
	if err != nil {
		return err
	}

	if projAdminRobotPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  1,
		Role:       models.PROJECTADMIN,
		EntityID:   int(projAdminRobotID),
		EntityType: common.UserMember,
	}); err != nil {
		return err
	}

	// register projDeveloper and assign project developer role
	projDeveloperID, err = dao.Register(models.User{
		Username: projDeveloper.Name,
		Password: projDeveloper.Passwd,
		Email:    projDeveloper.Name + "@test.com",
	})
	if err != nil {
		return err
	}

	if projDeveloperPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  1,
		Role:       models.DEVELOPER,
		EntityID:   int(projDeveloperID),
		EntityType: common.UserMember,
	}); err != nil {
		return err
	}

	// register projGuest and assign project guest role
	projGuestID, err = dao.Register(models.User{
		Username: projGuest.Name,
		Password: projGuest.Passwd,
		Email:    projGuest.Name + "@test.com",
	})
	if err != nil {
		return err
	}

	if projGuestPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  1,
		Role:       models.GUEST,
		EntityID:   int(projGuestID),
		EntityType: common.UserMember,
	}); err != nil {
		return err
	}
	return err
}

func clean() {
	pmids := []int{projAdminPMID, projDeveloperPMID, projGuestPMID}

	for _, id := range pmids {
		if err := project.DeleteProjectMemberByID(id); err != nil {
			fmt.Printf("failed to clean up member %d from project library: %v", id, err)
		}
	}
	userids := []int64{nonSysAdminID, projAdminID, projDeveloperID, projGuestID}
	for _, id := range userids {
		if err := dao.DeleteUser(int(id)); err != nil {
			fmt.Printf("failed to clean up user %d: %v \n", id, err)
		}
	}
}

// Provides a mock chart controller for deletable test cases
func mockChartController() (*httptest.Server, *chartserver.Controller, error) {
	mockServer := httptest.NewServer(htesting.MockChartRepoHandler)

	var oldController, newController *chartserver.Controller
	url, err := url.Parse(mockServer.URL)
	if err == nil {
		newController, err = chartserver.NewController(url)
	}

	if err != nil {
		mockServer.Close()
		return nil, nil, err
	}

	// Override current controller and keep the old one for restoring
	oldController = chartController
	chartController = newController

	return mockServer, oldController, nil
}
