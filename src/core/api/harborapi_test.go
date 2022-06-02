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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"

	"github.com/beego/beego"
	"github.com/dghubble/sling"
	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/job/test"
	testutils "github.com/goharbor/harbor/src/common/utils/test"
	_ "github.com/goharbor/harbor/src/core/auth/db"
	_ "github.com/goharbor/harbor/src/core/auth/ldap"
	"github.com/goharbor/harbor/src/lib/config"
	libOrm "github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/orm"
	"github.com/goharbor/harbor/src/server/middleware/security"
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

func (a testapi) PingEmail(authInfo usrInfo, settings []byte) (int, string, error) {
	_sling := sling.New().Base(a.basePath).Post("/api/email/ping").Body(bytes.NewReader(settings))

	code, body, err := request(_sling, jsonAcceptHeader, authInfo)

	return code, string(body), err
}
