// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"

	"github.com/beego/beego/v2/server/web"
	"github.com/dghubble/sling"

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
	web.BConfig.WebConfig.Session.SessionOn = true
	web.TestBeegoInit(apppath)

	web.Router("/api/internal/syncquota", &InternalAPI{}, "post:SyncQuota")

	// Init user Info
	admin = &usrInfo{adminName, adminPwd}
	unknownUsr = &usrInfo{"unknown", "unknown"}
	testUser = &usrInfo{TestUserName, TestUserPwd}

	// Init mock jobservice
	mockServer := test.NewJobServiceServer()
	defer mockServer.Close()

	chain := middleware.Chain(orm.Middleware(), security.Middleware(), security.UnauthorizedMiddleware())
	handler = chain(web.BeeApp.Handlers)
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

	body, err := io.ReadAll(w.Body)
	return w.Code, w.Header(), body, err
}

func request(_sling *sling.Sling, acceptHeader string, authInfo ...usrInfo) (int, []byte, error) {
	code, _, body, err := request0(_sling, acceptHeader, authInfo...)
	return code, body, err
}
