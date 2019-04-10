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

package api

import (
	"fmt"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"net/http"
	"regexp"
	"strconv"
)

// UserAPI handles request to /api/users/{}
type UserAPI struct {
	BaseController
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

type userSearch struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

// Prepare validates the URL and parms
func (ua *UserAPI) Prepare() {
	ua.BaseController.Prepare()
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

	if !ua.SecurityCtx.IsAuthenticated() {
		if ua.Ctx.Input.IsPost() {
			return
		}
		ua.HandleUnauthorized()
		return
	}

	user, err := dao.GetUser(models.User{
		Username: ua.SecurityCtx.GetUsername(),
	})
	if err != nil {
		ua.HandleInternalServerError(fmt.Sprintf("failed to get user %s: %v",
			ua.SecurityCtx.GetUsername(), err))
		return
	}

	ua.currentUserID = user.UserID
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

	ua.IsAdmin = ua.SecurityCtx.IsSysAdmin()
}

// Get ...
func (ua *UserAPI) Get() {
	if ua.userID == ua.currentUserID || ua.IsAdmin {
		userQuery := models.User{UserID: ua.userID}
		u, err := dao.GetUser(userQuery)
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			ua.RenderFormatedError(http.StatusInternalServerError, err)
			return
		}
		u.Password = ""
		if ua.userID == ua.currentUserID {
			u.HasAdminRole = ua.SecurityCtx.IsSysAdmin()
		}
		if ua.AuthMode == common.OIDCAuth {
			o, err := ua.getOIDCUserInfo()
			if err != nil {
				ua.RenderFormatedError(http.StatusInternalServerError, err)
				return
			}
			u.OIDCUserMeta = o
		}
		ua.Data["json"] = u
		ua.ServeJSON()
		return
	}

	log.Errorf("Current user, id: %d does not have admin role, can not view other user's detail", ua.currentUserID)
	ua.RenderError(http.StatusForbidden, "User does not have admin role")
	return
}

// List ...
func (ua *UserAPI) List() {
	if !ua.IsAdmin {
		log.Errorf("Current user, id: %d does not have admin role, can not list users", ua.currentUserID)
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}

	page, size := ua.GetPaginationParams()
	query := &models.UserQuery{
		Username: ua.GetString("username"),
		Email:    ua.GetString("email"),
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}

	total, err := dao.GetTotalOfUsers(query)
	if err != nil {
		ua.HandleInternalServerError(fmt.Sprintf("failed to get total of users: %v", err))
		return
	}

	users, err := dao.ListUsers(query)
	if err != nil {
		ua.HandleInternalServerError(fmt.Sprintf("failed to get users: %v", err))
		return
	}

	ua.SetPaginationHeader(total, page, size)
	ua.Data["json"] = users
	ua.ServeJSON()
}

// Search ...
func (ua *UserAPI) Search() {
	page, size := ua.GetPaginationParams()
	query := &models.UserQuery{
		Username: ua.GetString("username"),
		Email:    ua.GetString("email"),
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}

	total, err := dao.GetTotalOfUsers(query)
	if err != nil {
		ua.HandleInternalServerError(fmt.Sprintf("failed to get total of users: %v", err))
		return
	}

	users, err := dao.ListUsers(query)
	if err != nil {
		ua.HandleInternalServerError(fmt.Sprintf("failed to get users: %v", err))
		return
	}

	var userSearches []userSearch
	for _, user := range users {
		userSearches = append(userSearches, userSearch{UserID: user.UserID, Username: user.Username})
	}

	ua.SetPaginationHeader(total, page, size)
	ua.Data["json"] = userSearches
	ua.ServeJSON()
}

// Put ...
func (ua *UserAPI) Put() {
	if !ua.modifiable() {
		ua.RenderError(http.StatusForbidden, fmt.Sprintf("User with ID %d cannot be modified", ua.userID))
		return
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

	if !(ua.AuthMode == common.DBAuth) {
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
	if !ua.IsAdmin || ua.AuthMode != common.DBAuth || ua.userID == 1 || ua.currentUserID == ua.userID {
		ua.RenderError(http.StatusForbidden, fmt.Sprintf("User with ID: %d cannot be removed, auth mode: %s, current user ID: %d", ua.userID, ua.AuthMode, ua.currentUserID))
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
	if !ua.modifiable() {
		ua.RenderError(http.StatusForbidden, fmt.Sprintf("User with ID: %d is not modifiable", ua.userID))
		return
	}

	changePwdOfOwn := ua.userID == ua.currentUserID

	var req passwordReq
	ua.DecodeJSONReq(&req)

	if changePwdOfOwn && len(req.OldPassword) == 0 {
		ua.HandleBadRequest("empty old_password")
		return
	}

	if len(req.NewPassword) == 0 {
		ua.HandleBadRequest("empty new_password")
		return
	}

	user, err := dao.GetUser(models.User{UserID: ua.userID})
	if err != nil {
		ua.HandleInternalServerError(fmt.Sprintf("failed to get user %d: %v", ua.userID, err))
		return
	}
	if user == nil {
		ua.HandleNotFound(fmt.Sprintf("user %d not found", ua.userID))
		return
	}
	if changePwdOfOwn {
		if user.Password != utils.Encrypt(req.OldPassword, user.Salt) {
			log.Info("incorrect old_password")
			ua.RenderError(http.StatusForbidden, "incorrect old_password")
			return
		}
	}
	if user.Password == utils.Encrypt(req.NewPassword, user.Salt) {
		ua.HandleBadRequest("the new password can not be same with the old one")
		return
	}

	updatedUser := models.User{
		UserID:   ua.userID,
		Password: req.NewPassword,
	}
	if err = dao.ChangeUserPassword(updatedUser); err != nil {
		ua.HandleInternalServerError(fmt.Sprintf("failed to change password of user %d: %v", ua.userID, err))
		return
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

// ListUserPermissions handles GET to /api/users/{}/permissions
func (ua *UserAPI) ListUserPermissions() {
	if ua.userID != ua.currentUserID {
		log.Warningf("Current user, id: %d can not view other user's permissions", ua.currentUserID)
		ua.RenderError(http.StatusForbidden, "User does not have permission")
		return
	}

	relative := ua.Ctx.Input.Query("relative") == "true"

	scope := rbac.Resource(ua.Ctx.Input.Query("scope"))
	policies := []*rbac.Policy{}

	namespace, err := scope.GetNamespace()
	if err == nil {
		switch namespace.Kind() {
		case "project":
			for _, policy := range project.GetAllPolicies(namespace) {
				if ua.SecurityCtx.Can(policy.Action, policy.Resource) {
					policies = append(policies, policy)
				}
			}
		}
	}

	results := []map[string]string{}
	for _, policy := range policies {
		var resource rbac.Resource

		// for resource `/project/1/repository` if `relative` is `true` then the resource in response will be `repository`
		if relative {
			relativeResource, err := policy.Resource.RelativeTo(scope)
			if err != nil {
				continue
			}
			resource = relativeResource
		} else {
			resource = policy.Resource
		}

		results = append(results, map[string]string{
			"resource": resource.String(),
			"action":   policy.Action.String(),
		})
	}

	ua.Data["json"] = results
	ua.ServeJSON()
	return
}

func (ua *UserAPI) getOIDCUserInfo() (*models.OIDCUser, error) {
	key, err := config.SecretKey()
	if err != nil {
		return nil, err
	}
	o, err := dao.GetOIDCUserByUserID(ua.userID)
	if err != nil || o == nil {
		return nil, err
	}
	if len(o.Secret) > 0 {
		p, err := utils.ReversibleDecrypt(o.Secret, key)
		if err != nil {
			return nil, err
		}
		o.PlainSecret = p
	}
	return o, nil
}

// modifiable returns whether the modify is allowed based on current auth mode and context
func (ua *UserAPI) modifiable() bool {
	if ua.AuthMode == common.DBAuth {
		// When the auth mode is local DB, admin can modify anyone, non-admin can modify himself.
		return ua.IsAdmin || ua.userID == ua.currentUserID
	}
	// When the auth mode is external IDM backend, only the super user can modify himself,
	// because he's the only one whose information is stored in local DB.
	return ua.userID == 1 && ua.userID == ua.currentUserID

}

// validate only validate when user register
func validate(user models.User) error {

	if utils.IsIllegalLength(user.Username, 1, 255) {
		return fmt.Errorf("username with illegal length")
	}
	if utils.IsContainIllegalChar(user.Username, []string{",", "~", "#", "$", "%"}) {
		return fmt.Errorf("username contains illegal characters")
	}
	if utils.IsIllegalLength(user.Password, 8, 20) {
		return fmt.Errorf("password with illegal length")
	}
	return commonValidate(user)
}

// commonValidate validates email, realname, comment information when user register or change their profile
func commonValidate(user models.User) error {

	if len(user.Email) > 0 {
		if m, _ := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, user.Email); !m {
			return fmt.Errorf("email with illegal format")
		}
	} else {
		return fmt.Errorf("Email can't be empty")
	}

	if utils.IsIllegalLength(user.Realname, 1, 255) {
		return fmt.Errorf("realname with illegal length")
	}

	if utils.IsContainIllegalChar(user.Realname, []string{",", "~", "#", "$", "%"}) {
		return fmt.Errorf("realname contains illegal characters")
	}
	if utils.IsIllegalLength(user.Comment, -1, 30) {
		return fmt.Errorf("comment with illegal length")
	}
	return nil

}
