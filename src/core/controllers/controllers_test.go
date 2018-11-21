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
	"net/http"
	"net/http/httptest"
	// "net/url"
	"path/filepath"
	"runtime"
	"testing"

	"fmt"
	"os"
	"strings"

	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/config/client/db"
	"github.com/goharbor/harbor/src/common/models"
	_ "github.com/goharbor/harbor/src/core/auth/db"
	_ "github.com/goharbor/harbor/src/core/auth/ldap"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/proxy"
	"github.com/stretchr/testify/assert"
)

// const (
//	adminName = "admin"
//	adminPwd  = "Harbor12345"
// )

// type usrInfo struct {
//	Name   string
//	Passwd string
// }

// var admin *usrInfo

func init() {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	dir = filepath.Join(dir, "..")
	apppath, _ := filepath.Abs(dir)
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.TestBeegoInit(apppath)
	beego.AddTemplateExt("htm")

	beego.Router("/", &IndexController{})

	beego.Router("/c/login", &CommonController{}, "post:Login")
	beego.Router("/c/log_out", &CommonController{}, "get:LogOut")
	beego.Router("/c/reset", &CommonController{}, "post:ResetPassword")
	beego.Router("/c/userExists", &CommonController{}, "post:UserExists")
	beego.Router("/c/sendEmail", &CommonController{}, "get:SendResetEmail")
	beego.Router("/v2/*", &RegistryProxy{}, "*:Handle")
}

func TestMain(m *testing.M) {
	db.InitDatabaseAndConfigure()
	cfgManager := db.NewCoreConfigManager()
	cfgManager.Upload(common.TestServerDefaultConfig)
	cfg, err := cfgManager.Get()
	if err != nil {
		fmt.Printf("Error occurred when get config: %v", err)
	}
	fmt.Printf("config settings,cfg:%v\n", cfg)

	rc := m.Run()
	if rc != 0 {
		os.Exit(rc)
	}
	// Init user Info
	// admin = &usrInfo{adminName, adminPwd}
}

// TestUserResettable
func TestUserResettable(t *testing.T) {
	assert := assert.New(t)
	DBAuthConfig := map[string]interface{}{
		common.AUTHMode:        common.DBAuth,
		common.CfgExpiration:   5,
		common.TokenExpiration: 30,
	}
	cfgManager := db.NewCoreConfigManager()
	LDAPAuthConfig := map[string]interface{}{
		common.AUTHMode:        common.LDAPAuth,
		common.CfgExpiration:   5,
		common.TokenExpiration: 30,
	}
	cfgManager.Upload(LDAPAuthConfig)
	if err := config.Init(); err != nil {
		panic(err)
	}
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
	cfgManager.Upload(DBAuthConfig)
	assert.True(isUserResetable(u2))

}

// TestMain is a sample to run an endpoint test
func TestAll(t *testing.T) {
	cfgManager := db.NewCoreConfigManager()
	TestConfig := map[string]interface{}{
		common.RegistryURL: "http://" + os.Getenv("REGISTRY_URL"),
	}
	cfgManager.Upload(TestConfig)

	if err := config.Init(); err != nil {
		panic(err)
	}
	if err := proxy.Init(); err != nil {
		panic(err)
	}

	// database, err := config.Database()
	// if err != nil {
	// 	panic(err)
	// }
	// if err := dao.InitDatabase(database); err != nil {
	// 	panic(err)
	// }

	assert := assert.New(t)

	//	v := url.Values{}
	//	v.Set("principal", "admin")
	//	v.Add("password", "Harbor12345")

	r, _ := http.NewRequest("POST", "/c/login", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(401), w.Code, "'/c/login' httpStatusCode should be 401")

	r, _ = http.NewRequest("GET", "/c/log_out", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/c/log_out' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), ""), "http respond should be empty")

	r, _ = http.NewRequest("POST", "/c/reset", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(400), w.Code, "'/c/reset' httpStatusCode should be 400")

	r, _ = http.NewRequest("POST", "/c/userExists", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(500), w.Code, "'/c/userExists' httpStatusCode should be 500")

	r, _ = http.NewRequest("GET", "/c/sendEmail", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(400), w.Code, "'/c/sendEmail' httpStatusCode should be 400")

	r, _ = http.NewRequest("GET", "/v2/", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "ping v2 should get a 200 response")

	r, _ = http.NewRequest("GET", "/v2/noproject/manifests/1.0", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(400), w.Code, "GET v2/noproject/manifests/1.0 should get a 400 response")

	r, _ = http.NewRequest("GET", "/v2/project/notexist/manifests/1.0", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(404), w.Code, "GET v2/noproject/manifests/1.0 should get a 404 response")
}
