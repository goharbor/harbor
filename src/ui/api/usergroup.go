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

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/dao/group"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// UserGroupAPI ...
type UserGroupAPI struct {
	BaseController
	id            int
	currentUserID int
}

// Prepare validates the URL and parms
func (uga *UserGroupAPI) Prepare() {
	uga.BaseController.Prepare()
	if !uga.SecurityCtx.IsAuthenticated() {
		uga.HandleUnauthorized()
		return
	}
	user, err := dao.GetUser(models.User{
		Username: uga.SecurityCtx.GetUsername(),
	})
	if err != nil {
		uga.HandleInternalServerError(
			fmt.Sprintf("failed to get user %s: %v",
				uga.SecurityCtx.GetUsername(), err))
		return
	}
	uga.currentUserID = user.UserID

	ugid, err := uga.GetInt64FromPath(":ugid")
	if err != nil {
		log.Errorf("failed to parse user group id, error: %v", err)
	} else {
		uga.id = int(ugid)
	}

}

// Get ...
func (uga *UserGroupAPI) Get() {
	ID := uga.id
	uga.Data["json"] = make([]models.UserGroup, 0)
	if ID == 0 {
		//user group id not set, return all user group
		query := models.UserGroup{GroupType: 1} //Current query LDAP group only
		userGroupList, err := group.QueryUserGroup(query)
		if err != nil {
			log.Errorf("Failed to query database for user group list, error: %v", err)
			uga.RenderError(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		if len(userGroupList) > 0 {
			uga.Data["json"] = userGroupList
		}
	} else {
		//return a specific user group
		userGroup, err := group.GetUserGroup(ID)
		if err != nil {
			log.Errorf("Failed to query database for user group list, error: %v", err)
			uga.RenderError(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		if userGroup == nil {
			log.Errorf("Failed to query database for user group list, user group not found, id %v", ID)
			uga.RenderError(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		uga.Data["json"] = userGroup
	}
	uga.ServeJSON()
}

// Post ... Create User Group
func (uga *UserGroupAPI) Post() {
	userGroup := models.UserGroup{}
	uga.DecodeJSONReq(&userGroup)
	_, err := group.AddUserGroup(userGroup)
	if err != nil {
		log.Errorf("Error occurred in add user group, error: %v", err)
		uga.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	return
}

// Put ... Only support update name
func (uga *UserGroupAPI) Put() {
	userGroup := models.UserGroup{}
	uga.DecodeJSONReq(&userGroup)
	ID := uga.id
	log.Debugf("Updated user group %v", userGroup)
	if ID > 0 {
		err := group.UpdateUserGroup(ID, userGroup.GroupName)
		if err != nil {
			log.Errorf("Error occurred in update user group, error: %v", err)
			uga.CustomAbort(http.StatusInternalServerError, "Internal error.")
			return
		}
	}
	return
}

// Delete ...
func (uga *UserGroupAPI) Delete() {
	if uga.id > 0 {
		err := group.DeleteUserGroup(uga.id)
		if err != nil {
			log.Errorf("Error occurred in update user group, error: %v", err)
			uga.CustomAbort(http.StatusInternalServerError, "Internal error.")
			return
		}
	}
	return
}
