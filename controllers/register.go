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

package controllers

import (
	"net/http"
	"os"
	"strings"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego"
)

// RegisterController handles request to /register
type RegisterController struct {
	BaseController
}

// Get renders the Sign In page, it only works if the auth mode is set to db_auth
func (rc *RegisterController) Get() {
	authMode := os.Getenv("AUTH_MODE")
	if authMode == "" || authMode == "db_auth" {
		rc.ForwardTo("page_title_registration", "register")
	} else {
		rc.Redirect("/signIn", http.StatusNotFound)
	}
}

// SignUp insert data into DB based on data in form.
func (rc *CommonController) SignUp() {
	username := strings.TrimSpace(rc.GetString("username"))
	email := strings.TrimSpace(rc.GetString("email"))
	realname := strings.TrimSpace(rc.GetString("realname"))
	password := strings.TrimSpace(rc.GetString("password"))
	comment := strings.TrimSpace(rc.GetString("comment"))

	user := models.User{Username: username, Email: email, Realname: realname, Password: password, Comment: comment}

	_, err := dao.Register(user)
	if err != nil {
		beego.Error("Error occurred in Register:", err)
		rc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
}

// UserExists checks if user exists when user input value in sign in form.
func (rc *CommonController) UserExists() {
	target := rc.GetString("target")
	value := rc.GetString("value")

	user := models.User{}
	switch target {
	case "username":
		user.Username = value
	case "email":
		user.Email = value
	}

	exist, err := dao.UserExists(user, target)
	if err != nil {
		beego.Error("Error occurred in UserExists:", err)
		rc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	rc.Data["json"] = exist
	rc.ServeJSON()
}
