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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao/group"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/ldap"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/auth"
)

// UserGroupAPI ...
type UserGroupAPI struct {
	BaseController
	id int
}

const (
	userNameEmptyMsg = "User group name can not be empty!"
)

// Prepare validates the URL and parms
func (uga *UserGroupAPI) Prepare() {
	uga.BaseController.Prepare()
	if !uga.SecurityCtx.IsAuthenticated() {
		uga.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	ugid, err := uga.GetInt64FromPath(":ugid")
	if err != nil {
		log.Warningf("failed to parse user group id, error: %v", err)
	}
	if ugid <= 0 && (uga.Ctx.Input.IsPut() || uga.Ctx.Input.IsDelete()) {
		uga.SendBadRequestError(fmt.Errorf("invalid user group ID: %s", uga.GetStringFromPath(":ugid")))
		return
	}
	uga.id = int(ugid)
	// Common user can create/update, only harbor admin can delete user group.
	if uga.Ctx.Input.IsDelete() && !uga.SecurityCtx.IsSysAdmin() {
		uga.SendForbiddenError(errors.New(uga.SecurityCtx.GetUsername()))
		return
	}
}

// Get ...
func (uga *UserGroupAPI) Get() {
	ID := uga.id
	uga.Data["json"] = make([]models.UserGroup, 0)
	if ID == 0 {
		// user group id not set, return all user group
		query := models.UserGroup{GroupType: common.LdapGroupType} // Current query LDAP group only
		userGroupList, err := group.QueryUserGroup(query)
		if err != nil {
			uga.SendInternalServerError(fmt.Errorf("failed to query database for user group list, error: %v", err))
			return
		}
		if len(userGroupList) > 0 {
			uga.Data["json"] = userGroupList
		}
	} else {
		// return a specific user group
		userGroup, err := group.GetUserGroup(ID)
		if userGroup == nil {
			uga.SendNotFoundError(errors.New("the user group does not exist"))
			return
		}
		if err != nil {
			uga.SendInternalServerError(fmt.Errorf("failed to query database for user group list, error: %v", err))
			return
		}
		uga.Data["json"] = userGroup
	}
	uga.ServeJSON()
}

// Post ... Create User Group
func (uga *UserGroupAPI) Post() {
	userGroup := models.UserGroup{}
	if err := uga.DecodeJSONReq(&userGroup); err != nil {
		uga.SendBadRequestError(err)
		return
	}

	userGroup.ID = 0
	userGroup.GroupType = common.LdapGroupType
	userGroup.LdapGroupDN = strings.TrimSpace(userGroup.LdapGroupDN)
	userGroup.GroupName = strings.TrimSpace(userGroup.GroupName)
	if len(userGroup.GroupName) == 0 {
		uga.SendBadRequestError(errors.New(userNameEmptyMsg))
		return
	}
	query := models.UserGroup{GroupType: userGroup.GroupType, LdapGroupDN: userGroup.LdapGroupDN}
	result, err := group.QueryUserGroup(query)
	if err != nil {
		uga.SendInternalServerError(fmt.Errorf("error occurred in add user group, error: %v", err))
		return
	}
	if len(result) > 0 {
		uga.SendConflictError(errors.New("error occurred in add user group, duplicate user group exist"))
		return
	}
	// User can not add ldap group when the ldap server is offline
	ldapGroup, err := auth.SearchGroup(userGroup.LdapGroupDN)
	if err == ldap.ErrNotFound || ldapGroup == nil {
		uga.SendBadRequestError(fmt.Errorf("LDAP Group DN is not found: DN:%v", userGroup.LdapGroupDN))
		return
	}
	if err == ldap.ErrDNSyntax {
		uga.SendBadRequestError(fmt.Errorf("invalid DN syntax. DN: %v", userGroup.LdapGroupDN))
		return
	}
	if err != nil {
		uga.SendInternalServerError(fmt.Errorf("Error occurred in search user group. error: %v", err))
		return
	}

	groupID, err := group.AddUserGroup(userGroup)
	if err != nil {
		uga.SendInternalServerError(fmt.Errorf("Error occurred in add user group, error: %v", err))
		return
	}
	uga.Redirect(http.StatusCreated, strconv.FormatInt(int64(groupID), 10))
}

// Put ... Only support update name
func (uga *UserGroupAPI) Put() {
	userGroup := models.UserGroup{}
	if err := uga.DecodeJSONReq(&userGroup); err != nil {
		uga.SendBadRequestError(err)
		return
	}
	ID := uga.id
	userGroup.GroupName = strings.TrimSpace(userGroup.GroupName)
	if len(userGroup.GroupName) == 0 {
		uga.SendBadRequestError(errors.New(userNameEmptyMsg))
		return
	}
	userGroup.GroupType = common.LdapGroupType
	log.Debugf("Updated user group %v", userGroup)
	err := group.UpdateUserGroupName(ID, userGroup.GroupName)
	if err != nil {
		uga.SendInternalServerError(fmt.Errorf("Error occurred in update user group, error: %v", err))
		return
	}
	return
}

// Delete ...
func (uga *UserGroupAPI) Delete() {
	err := group.DeleteUserGroup(uga.id)
	if err != nil {
		uga.SendInternalServerError(fmt.Errorf("Error occurred in update user group, error: %v", err))
		return
	}
	return
}
