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
package controllers

import (
	"net/http"
	"net/http/httptest"
	//"net/url"
	"path/filepath"
	"runtime"
	"testing"

	"fmt"
	"strings"

	"github.com/astaxie/beego"
	//"github.com/dghubble/sling"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

//const (
//	adminName = "admin"
//	adminPwd  = "Harbor12345"
//)

//type usrInfo struct {
//	Name   string
//	Passwd string
//}

//var admin *usrInfo

func init() {
	if err := config.Init(); err != nil {
		log.Fatalf("failed to initialize configurations: %v", err)
	}

	_, file, _, _ := runtime.Caller(1)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".."+string(filepath.Separator))))
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.TestBeegoInit(apppath)
	beego.AddTemplateExt("htm")

	beego.Router("/", &IndexController{})

	beego.Router("/login", &CommonController{}, "post:Login")
	beego.Router("/log_out", &CommonController{}, "get:LogOut")
	beego.Router("/reset", &CommonController{}, "post:ResetPassword")
	beego.Router("/userExists", &CommonController{}, "post:UserExists")
	beego.Router("/sendEmail", &CommonController{}, "get:SendEmail")

	//Init user Info
	//admin = &usrInfo{adminName, adminPwd}
}

// TestMain is a sample to run an endpoint test
func TestMain(t *testing.T) {
	assert := assert.New(t)

	//	v := url.Values{}
	//	v.Set("principal", "admin")
	//	v.Add("password", "Harbor12345")

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), "<title>Harbor</title>"), "http respond should have '<title>Harbor</title>'")

	r, _ = http.NewRequest("POST", "/login", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(401), w.Code, "'/login' httpStatusCode should be 401")

	r, _ = http.NewRequest("GET", "/log_out", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(200), w.Code, "'/log_out' httpStatusCode should be 200")
	assert.Equal(true, strings.Contains(fmt.Sprintf("%s", w.Body), ""), "http respond should be empty")

	r, _ = http.NewRequest("POST", "/reset", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(400), w.Code, "'/reset' httpStatusCode should be 400")

	r, _ = http.NewRequest("POST", "/userExists", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(500), w.Code, "'/userExists' httpStatusCode should be 500")

	r, _ = http.NewRequest("GET", "/sendEmail", nil)
	w = httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	assert.Equal(int(400), w.Code, "'/sendEmail' httpStatusCode should be 400")

}
