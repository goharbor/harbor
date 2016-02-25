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
	"strconv"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego"
)

type UserAPI struct {
	BaseAPI
	currentUserID int
	userID        int
}

func (ua *UserAPI) Prepare() {

	ua.currentUserID = ua.ValidateUser()
	id := ua.Ctx.Input.Param(":id")
	if id == "current" {
		ua.userID = ua.currentUserID
	} else if len(id) > 0 {
		var err error
		ua.userID, err = strconv.Atoi(id)
		if err != nil {
			beego.Error("Invalid user id, error:", err)
			ua.CustomAbort(http.StatusBadRequest, "Invalid user Id")
		}
		userQuery := models.User{UserId: ua.userID}
		u, err := dao.GetUser(userQuery)
		if err != nil {
			beego.Error("Error occurred in GetUser:", err)
			ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if u == nil {
			beego.Error("User with Id:", ua.userID, "does not exist")
			ua.CustomAbort(http.StatusNotFound, "")
		}
	}
}

func (ua *UserAPI) Get() {
	exist, err := dao.IsAdminRole(ua.currentUserID)
	if err != nil {
		beego.Error("Error occurred in IsAdminRole:", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if ua.userID == 0 { //list users
		if !exist {
			beego.Error("Current user, id:", ua.currentUserID, ", does not have admin role, can not list users")
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
			beego.Error("Failed to get data from database, error:", err)
			ua.RenderError(http.StatusInternalServerError, "Failed to query from database")
			return
		}
		ua.Data["json"] = userList

	} else if ua.userID == ua.currentUserID || exist {
		userQuery := models.User{UserId: ua.userID}
		u, err := dao.GetUser(userQuery)
		if err != nil {
			beego.Error("Error occurred in GetUser:", err)
			ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		ua.Data["json"] = u
	} else {
		beego.Error("Current user, id:", ua.currentUserID, "does not have admin role, can not view other user's detail")
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}
	ua.ServeJSON()
}

func (ua *UserAPI) Put() { //currently only for toggle admin, so no request body
	exist, err := dao.IsAdminRole(ua.currentUserID)
	if err != nil {
		beego.Error("Error occurred in IsAdminRole:", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if !exist {
		beego.Warning("current user, id:", ua.currentUserID, ", does not have admin role, can not update other user's role")
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}
	userQuery := models.User{UserId: ua.userID}
	dao.ToggleUserAdminRole(userQuery)
}

func (ua *UserAPI) Delete() {
	exist, err := dao.IsAdminRole(ua.currentUserID)
	if err != nil {
		beego.Error("Error occurred in IsAdminRole:", err)
		ua.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if !exist {
		beego.Warning("current user, id:", ua.currentUserID, ", does not have admin role, can not remove user")
		ua.RenderError(http.StatusForbidden, "User does not have admin role")
		return
	}
	err = dao.DeleteUser(ua.userID)
	if err != nil {
		beego.Error("Failed to delete data from database, error:", err)
		ua.RenderError(http.StatusInternalServerError, "Failed to delete User")
		return
	}
}
