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

package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/vmware/harbor/src/common/api"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

// UserAPI handles request to /api/users/{}
type UserAPI struct {
	api.BaseAPI
	currentUserID    int
	userID           int
	SelfRegistration bool
	IsAdmin          bool
	AuthMode         string
}

type passwordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// Prepare validates the URL and parms
func (ua *UserAPI) Prepare() {
	mode, err := config.AuthMode()
	if err != nil {
		log.Errorf("failed to get auth mode: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	ua.AuthMode = mode

	self, err := config.SelfRegistration()
	if err != nil {
		log.Errorf("failed to get self registration: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	ua.SelfRegistration = self

	if ua.Ctx.Input.IsPost() {
		sessionUserID := ua.GetSession("userId")
		_, _, ok := ua.Ctx.Request.BasicAuth()
		if sessionUserID == nil && !ok {
			return
		}
	}

	ua.currentUserID = ua.ValidateUser()
	id := ua.Ctx.Input.Param(":id")
	if id == "current" {
		ua.userID = ua.currentUserID
	} else if len(id) > 0 {
		var err error
		ua.userID, err = strconv.Atoi(id)
		if err != nil {
			log.Errorf("Invalid user id, error: %v", err)
			ua.CustomAbort(http.StatusBadRequest, "Invalid user Id")
		}
		userQuery := models.User{UserID: ua.userID}
		u, err := dao.GetUser(userQuery)
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if u == nil {
			log.Errorf("User with Id: %d does not exist", ua.userID)
			ua.CustomAbort(http.StatusNotFound, "")
		}
	}

	ua.IsAdmin, err = dao.IsAdminRole(ua.currentUserID)
	if err != nil {
		log.Errorf("Error occurred in IsAdminRole:%v", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

}

// Get ...
func (ua *UserAPI) Get() {
	if ua.userID == 0 { //list users
		if !ua.IsAdmin {
			log.Errorf("Current user, id: %d does not have admin role, can not list users", ua.currentUserID)
			ua.RenderError(http.StatusForbidden, "User does not have admin role")
			return
		}
		username := ua.GetString("username")
		userQuery := models.User{}
		if len(username) > 0 {
			userQuery.Username = username
		}
		userList, err := dao.ListUsers(userQuery)
		if err != nil {
			log.Errorf("Failed to get data from database, error: %v", err)
			ua.RenderError(http.StatusInternalServerError, "Failed to query from database")
			return
		}
		ua.Data["json"] = userList

	} else if ua.userID == ua.currentUserID || ua.IsAdmin {
		userQuery := models.User{UserID: ua.userID}
		u, err := dao.GetUser(userQuery)
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		ua.Data["json"] = u
	} else {
		log.Errorf("Current user, id: %d does not have admin role, can not view other user's detail", ua.currentUserID)
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}
	ua.ServeJSON()
}

// Put ...
func (ua *UserAPI) Put() {
	ldapAdminUser := (ua.AuthMode == "ldap_auth" && ua.userID == 1 && ua.userID == ua.currentUserID)

	if !(ua.AuthMode == "db_auth" || ldapAdminUser) {
		ua.CustomAbort(http.StatusForbidden, "")
	}
	if !ua.IsAdmin {
		if ua.userID != ua.currentUserID {
			log.Warning("Guests can only change their own account.")
			ua.CustomAbort(http.StatusForbidden, "Guests can only change their own account.")
		}
	}
	user := models.User{UserID: ua.userID}
	ua.DecodeJSONReq(&user)
	err := commonValidate(user)
	if err != nil {
		log.Warningf("Bad request in change user profile: %v", err)
		ua.RenderError(http.StatusBadRequest, "change user profile error:"+err.Error())
		return
	}
	userQuery := models.User{UserID: ua.userID}
	u, err := dao.GetUser(userQuery)
	if err != nil {
		log.Errorf("Error occurred in GetUser, error: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if u == nil {
		log.Errorf("User with Id: %d does not exist", ua.userID)
		ua.CustomAbort(http.StatusNotFound, "")
	}
	if u.Email != user.Email {
		emailExist, err := dao.UserExists(user, "email")
		if err != nil {
			log.Errorf("Error occurred in change user profile: %v", err)
			ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if emailExist {
			log.Warning("email has already been used!")
			ua.RenderError(http.StatusConflict, "email has already been used!")
			return
		}
	}
	if err := dao.ChangeUserProfile(user); err != nil {
		log.Errorf("Failed to update user profile, error: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, err.Error())
	}
}

// Post ...
func (ua *UserAPI) Post() {

	if !(ua.AuthMode == "db_auth") {
		ua.CustomAbort(http.StatusForbidden, "")
	}

	if !(ua.SelfRegistration || ua.IsAdmin) {
		log.Warning("Registration can only be used by admin role user when self-registration is off.")
		ua.CustomAbort(http.StatusForbidden, "")
	}

	user := models.User{}
	ua.DecodeJSONReq(&user)
	err := validate(user)
	if err != nil {
		log.Warningf("Bad request in Register: %v", err)
		ua.RenderError(http.StatusBadRequest, "register error:"+err.Error())
		return
	}
	userExist, err := dao.UserExists(user, "username")
	if err != nil {
		log.Errorf("Error occurred in Register: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if userExist {
		log.Warning("username has already been used!")
		ua.RenderError(http.StatusConflict, "username has already been used!")
		return
	}
	emailExist, err := dao.UserExists(user, "email")
	if err != nil {
		log.Errorf("Error occurred in change user profile: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if emailExist {
		log.Warning("email has already been used!")
		ua.RenderError(http.StatusConflict, "email has already been used!")
		return
	}
	userID, err := dao.Register(user)
	if err != nil {
		log.Errorf("Error occurred in Register: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	ua.Redirect(http.StatusCreated, strconv.FormatInt(userID, 10))
}

// Delete ...
func (ua *UserAPI) Delete() {
	if !ua.IsAdmin {
		log.Warningf("current user, id: %d does not have admin role, can not remove user", ua.currentUserID)
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}

	if ua.AuthMode == "ldap_auth" {
		ua.CustomAbort(http.StatusForbidden, "user can not be deleted in LDAP authentication mode")
	}

	if ua.currentUserID == ua.userID {
		ua.CustomAbort(http.StatusForbidden, "can not delete yourself")
	}

	var err error
	err = dao.DeleteUser(ua.userID)
	if err != nil {
		log.Errorf("Failed to delete data from database, error: %v", err)
		ua.RenderError(http.StatusInternalServerError, "Failed to delete User")
		return
	}
}

// ChangePassword handles PUT to /api/users/{}/password
func (ua *UserAPI) ChangePassword() {
	ldapAdminUser := (ua.AuthMode == "ldap_auth" && ua.userID == 1 && ua.userID == ua.currentUserID)

	if !(ua.AuthMode == "db_auth" || ldapAdminUser) {
		ua.CustomAbort(http.StatusForbidden, "")
	}

	if !ua.IsAdmin {
		if ua.userID != ua.currentUserID {
			log.Error("Guests can only change their own account.")
			ua.CustomAbort(http.StatusForbidden, "Guests can only change their own account.")
		}
	}

	var req passwordReq
	ua.DecodeJSONReq(&req)
	if req.OldPassword == "" {
		log.Error("Old password is blank")
		ua.CustomAbort(http.StatusBadRequest, "Old password is blank")
	}

	queryUser := models.User{UserID: ua.userID, Password: req.OldPassword}
	user, err := dao.CheckUserPassword(queryUser)
	if err != nil {
		log.Errorf("Error occurred in CheckUserPassword: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if user == nil {
		log.Warning("Password input is not correct")
		ua.CustomAbort(http.StatusForbidden, "old_password_is_not_correct")
	}

	if req.NewPassword == "" {
		ua.CustomAbort(http.StatusBadRequest, "please_input_new_password")
	}
	updateUser := models.User{UserID: ua.userID, Password: req.NewPassword, Salt: user.Salt}
	err = dao.ChangeUserPassword(updateUser, req.OldPassword)
	if err != nil {
		log.Errorf("Error occurred in ChangeUserPassword: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
}

// ToggleUserAdminRole handles PUT api/users/{}/sysadmin
func (ua *UserAPI) ToggleUserAdminRole() {
	if !ua.IsAdmin {
		log.Warningf("current user, id: %d does not have admin role, can not update other user's role", ua.currentUserID)
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}
	userQuery := models.User{UserID: ua.userID}
	ua.DecodeJSONReq(&userQuery)
	if err := dao.ToggleUserAdminRole(userQuery.UserID, userQuery.HasAdminRole); err != nil {
		log.Errorf("Error occurred in ToggleUserAdminRole: %v", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
}

// validate only validate when user register
func validate(user models.User) error {

	if isIllegalLength(user.Username, 1, 20) {
		return fmt.Errorf("username with illegal length")
	}
	if isContainIllegalChar(user.Username, []string{",", "~", "#", "$", "%"}) {
		return fmt.Errorf("username contains illegal characters")
	}
	if isIllegalLength(user.Password, 8, 20) {
		return fmt.Errorf("password with illegal length")
	}
	if err := commonValidate(user); err != nil {
		return err
	}
	return nil
}

//commonValidate validates email, realname, comment information when user register or change their profile
func commonValidate(user models.User) error {

	if len(user.Email) > 0 {
		if m, _ := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, user.Email); !m {
			return fmt.Errorf("email with illegal format")
		}
	} else {
		return fmt.Errorf("Email can't be empty")
	}

	if isIllegalLength(user.Realname, 0, 20) {
		return fmt.Errorf("realname with illegal length")
	}

	if isContainIllegalChar(user.Realname, []string{",", "~", "#", "$", "%"}) {
		return fmt.Errorf("realname contains illegal characters")
	}
	if isIllegalLength(user.Comment, -1, 30) {
		return fmt.Errorf("comment with illegal length")
	}
	return nil

}

func isIllegalLength(s string, min int, max int) bool {
	if min == -1 {
		return (len(s) > max)
	}
	if max == -1 {
		return (len(s) <= min)
	}
	return (len(s) < min || len(s) > max)
}

func isContainIllegalChar(s string, illegalChar []string) bool {
	for _, c := range illegalChar {
		if strings.Index(s, c) >= 0 {
			return true
		}
	}
	return false
}
