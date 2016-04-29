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

package api

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// UserAPI handles request to /api/users/{}
type UserAPI struct {
	BaseAPI
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

	authMode := strings.ToLower(os.Getenv("AUTH_MODE"))
	if authMode == "" {
		authMode = "db_auth"
	}
	ua.AuthMode = authMode

	selfRegistration := strings.ToLower(os.Getenv("SELF_REGISTRATION"))
	if selfRegistration == "on" {
		ua.SelfRegistration = true
	}

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

	var err error
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
			userQuery.Username = "%" + username + "%"
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
func (ua *UserAPI) Put() { //currently only for toggle admin, so no request body
	if !ua.IsAdmin {
		log.Warningf("current user, id: %d does not have admin role, can not update other user's role", ua.currentUserID)
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}
	userQuery := models.User{UserID: ua.userID}
	dao.ToggleUserAdminRole(userQuery)
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

	userID, err := dao.Register(user)
	if err != nil {
		log.Errorf("Error occurred in Register: %v", err)
		ua.RenderError(http.StatusInternalServerError, "Internal error.")
		return
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

	if !(ua.AuthMode == "db_auth") {
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
