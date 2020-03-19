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
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/core/middlewares"
	"github.com/goharbor/harbor/src/internal"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	utilstest "github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/config"
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
	beego.Router("/c/reset", &CommonController{}, "post:ResetPassword")
	beego.Router("/c/userExists", &CommonController{}, "post:UserExists")
	beego.Router("/c/sendEmail", &CommonController{}, "get:SendResetEmail")
}

func TestMain(m *testing.M) {
	utilstest.InitDatabaseFromEnv()
	rc := m.Run()
	if rc != 0 {
		os.Exit(rc)
	}
}

// TestUserResettable
func TestUserResettable(t *testing.T) {
	assert := assert.New(t)
	DBAuthConfig := map[string]interface{}{
		common.AUTHMode:        common.DBAuth,
		common.TokenExpiration: 30,
	}

	LDAPAuthConfig := map[string]interface{}{
		common.AUTHMode:        common.LDAPAuth,
		common.TokenExpiration: 30,
	}
	config.InitWithSettings(LDAPAuthConfig)
	u1 := &models.User{
		UserID:   3,
		Username: "daniel",
		Email:    "daniel@test.com",
	}
	u2 := &models.User{
		UserID:   1,
		Username: "jack",
		Email:    "jack@test.com",
	}
	assert.False(isUserResetable(u1))
	assert.True(isUserResetable(u2))
	config.InitWithSettings(DBAuthConfig)
	assert.True(isUserResetable(u1))
}

func TestRedirectForOIDC(t *testing.T) {
	ctx := internal.WithAuthMode(context.Background(), common.DBAuth)
	assert.False(t, redirectForOIDC(ctx, "nonexist"))
	ctx = internal.WithAuthMode(context.Background(), common.OIDCAuth)
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

	r, _ = http.NewRequest("POST", "/c/reset", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(http.StatusForbidden, w.Code, "'/c/reset' httpStatusCode should be 403")

	r, _ = http.NewRequest("POST", "/c/userExists", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(http.StatusForbidden, w.Code, "'/c/userExists' httpStatusCode should be 403")

	r, _ = http.NewRequest("GET", "/c/sendEmail", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	assert.Equal(int(400), w.Code, "'/c/sendEmail' httpStatusCode should be 400")
}
