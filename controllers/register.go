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

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"github.com/vmware/harbor/utils/log"
)

// RegisterController handles request to /register
type RegisterController struct {
	BaseController
}

// Get renders the Sign In page, it only works if the auth mode is set to db_auth
func (rc *RegisterController) Get() {

	if !rc.SelfRegistration {
		log.Warning("Registration is disabled when self-registration is off.")
		rc.Redirect("/signIn", http.StatusFound)
	}

	if rc.AuthMode == "db_auth" {
		rc.ForwardTo("page_title_registration", "register")
	} else {
		rc.Redirect("/signIn", http.StatusFound)
	}
}

// AddUserController handles request for adding user with an admin role user
type AddUserController struct {
	BaseController
}

// Get renders the Sign In page, it only works if the auth mode is set to db_auth
func (ac *AddUserController) Get() {

	if !ac.IsAdmin {
		log.Warning("Add user can only be used by admin role user.")
		ac.Redirect("/signIn", http.StatusFound)
	}

	if ac.AuthMode == "db_auth" {
		ac.ForwardTo("page_title_add_user", "register")
	} else {
		ac.Redirect("/signIn", http.StatusFound)
	}
}

// UserExists checks if user exists when user input value in sign in form.
func (cc *CommonController) UserExists() {
	target := cc.GetString("target")
	value := cc.GetString("value")

	user := models.User{}
	switch target {
	case "username":
		user.Username = value
	case "email":
		user.Email = value
	}

	exist, err := dao.UserExists(user, target)
	if err != nil {
		log.Errorf("Error occurred in UserExists: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	cc.Data["json"] = exist
	cc.ServeJSON()
}
