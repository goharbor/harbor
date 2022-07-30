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
package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/core/middlewares"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/orm"

	"github.com/beego/beego"
	"github.com/goharbor/harbor/src/common"
	utilstest "github.com/goharbor/harbor/src/common/utils/test"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/stretchr/testify/assert"
)

func init() {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	dir = filepath.Join(dir, "..")
	apppath, _ := filepath.Abs(dir)
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.TestBeegoInit(apppath)
	beego.AddTemplateExt("htm")

	beego.Router("/c/login", &CommonController{}, "post:Login")
	beego.Router("/c/log_out", &CommonController{}, "get:LogOut")
	beego.Router("/c/userExists", &CommonController{}, "post:UserExists")
}

func TestMain(m *testing.M) {
	utilstest.InitDatabaseFromEnv()
	rc := m.Run()
	if rc != 0 {
		os.Exit(rc)
	}
}

func TestRedirectForOIDC(t *testing.T) {
	ctx := lib.WithAuthMode(orm.Context(), common.DBAuth)
	assert.False(t, redirectForOIDC(ctx, "nonexist"))
	ctx = lib.WithAuthMode(orm.Context(), common.OIDCAuth)
	assert.True(t, redirectForOIDC(ctx, "nonexist"))
	assert.False(t, redirectForOIDC(ctx, "admin"))

}

// TestMain is a sample to run an endpoint test
func TestAll(t *testing.T) {
	config.InitWithSettings(utilstest.GetUnitTestConfig())
	assert := assert.New(t)
	handler := http.Handler(beego.BeeApp.Handlers)
	mws := middlewares.MiddleWares()
	for i := len(mws) - 1; i >= 0; i-- {
		if mws[i] == nil {
			continue
		}
		handler = mws[i](handler)
	}

	r, _ := http.NewRequest("POST", "/c/login", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(http.StatusForbidden, w.Code, "'/c/login' httpStatusCode should be 403")

	r, _ = http.NewRequest("GET", "/c/log_out", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/c/log_out' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), ""), "http respond should be empty")

	r, _ = http.NewRequest("POST", "/c/userExists", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(http.StatusForbidden, w.Code, "'/c/userExists' httpStatusCode should be 403")

}
