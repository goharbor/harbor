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
	"strconv"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego"
)

type UserAPI struct {
	BaseAPI
	currentUid int
	userId     int
}

func (ua *UserAPI) Prepare() {

	ua.currentUid = ua.ValidateUser()
	id := ua.Ctx.Input.Param(":id")
	if id == "current" {
		ua.userId = ua.currentUid
	} else if len(id) > 0 {
		var err error
		ua.userId, err = strconv.Atoi(id)
		if err != nil {
			beego.Error("Invalid user id, error:", err)
			ua.CustomAbort(400, "Invalid user Id")
		}
		userQuery := models.User{UserId: ua.userId}
		u, err := dao.GetUser(userQuery)
		if err != nil {
			beego.Error("Error occurred in GetUser:", err)
			ua.CustomAbort(500, "Internal error.")
		}
		if u == nil {
			beego.Error("User with Id:", ua.userId, "does not exist")
			ua.CustomAbort(404, "")
		}
	}
}

func (ua *UserAPI) Get() {
	exist, err := dao.IsAdminRole(ua.currentUid)
	if err != nil {
		beego.Error("Error occurred in IsAdminRole:", err)
		ua.CustomAbort(500, "Internal error.")
	}

	if ua.userId == 0 { //list users
		if !exist {
			beego.Error("Current user, id:", ua.currentUid, ", does not have admin role, can not list users")
			ua.RenderError(403, "User does not have admin role")
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
			ua.RenderError(500, "Failed to query from database")
			return
		}
		ua.Data["json"] = userList

	} else if ua.userId == ua.currentUid || exist {
		userQuery := models.User{UserId: ua.userId}
		u, err := dao.GetUser(userQuery)
		if err != nil {
			beego.Error("Error occurred in GetUser:", err)
			ua.CustomAbort(500, "Internal error.")
		}
		ua.Data["json"] = u
	} else {
		beego.Error("Current user, id:", ua.currentUid, "does not have admin role, can not view other user's detail")
		ua.RenderError(403, "User does not have admin role")
		return
	}
	ua.ServeJSON()
}

func (ua *UserAPI) Put() { //currently only for toggle admin, so no request body
	exist, err := dao.IsAdminRole(ua.currentUid)
	if err != nil {
		beego.Error("Error occurred in IsAdminRole:", err)
		ua.CustomAbort(500, "Internal error.")
	}
	if !exist {
		beego.Warning("current user, id:", ua.currentUid, ", does not have admin role, can not update other user's role")
		ua.RenderError(403, "User does not have admin role")
		return
	}
	userQuery := models.User{UserId: ua.userId}
	dao.ToggleUserAdminRole(userQuery)
}

func (ua *UserAPI) Delete() {
	exist, err := dao.IsAdminRole(ua.currentUid)
	if err != nil {
		beego.Error("Error occurred in IsAdminRole:", err)
		ua.CustomAbort(500, "Internal error.")
	}
	if !exist {
		beego.Warning("current user, id:", ua.currentUid, ", does not have admin role, can not remove user")
		ua.RenderError(403, "User does not have admin role")
		return
	}
	err = dao.DeleteUser(ua.userId)
	if err != nil {
		beego.Error("Failed to delete data from database, error:", err)
		ua.RenderError(500, "Failed to delete User")
		return
	}
}
