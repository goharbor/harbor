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

type ProjectMemberAPI struct {
	BaseAPI
	memberID      int
	currentUserID int
	project       *models.Project
}

type memberReq struct {
	Username string `json:"user_name"`
	UserID   int    `json:"user_id"`
	Roles    []int  `json:"roles"`
}

func (pma *ProjectMemberAPI) Prepare() {
	pid, err := strconv.ParseInt(pma.Ctx.Input.Param(":pid"), 10, 64)
	if err != nil {
		beego.Error("Error parsing project id:", pid, ", error:", err)
		pma.CustomAbort(http.StatusBadRequest, "invalid project Id")
		return
	}
	p, err := dao.GetProjectById(pid)
	if err != nil {
		beego.Error("Error occurred in GetProjectById:", err)
		pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if p == nil {
		beego.Warning("Project with id:", pid, "does not exist.")
		pma.CustomAbort(http.StatusNotFound, "Project does not exist")
	}
	pma.project = p
	pma.currentUserID = pma.ValidateUser()
	mid := pma.Ctx.Input.Param(":mid")
	if mid == "current" {
		pma.memberID = pma.currentUserID
	} else if len(mid) == 0 {
		pma.memberID = 0
	} else if len(mid) > 0 {
		memberID, err := strconv.Atoi(mid)
		if err != nil {
			beego.Error("Invalid member Id, error:", err)
			pma.CustomAbort(http.StatusBadRequest, "Invalid member id")
		}
		pma.memberID = memberID
	}
}

func (pma *ProjectMemberAPI) Get() {
	pid := pma.project.ProjectID
	if !CheckProjectPermission(pma.currentUserID, pid) {
		beego.Warning("Current user, user id :", pma.currentUserID, "does not have permission for project, id:", pid)
		pma.RenderError(http.StatusForbidden, "")
		return
	}
	if pma.memberID == 0 { //member id not set return list of the members
		username := pma.GetString("username")
		queryUser := models.User{Username: "%" + username + "%"}
		userList, err := dao.GetUserByProject(pid, queryUser)
		if err != nil {
			beego.Error("Failed to query database for member list, error:", err)
			pma.RenderError(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		pma.Data["json"] = userList
	} else { //return detail of a  member
		roleList, err := dao.GetUserProjectRoles(models.User{UserID: pma.memberID}, pid)
		if err != nil {
			beego.Error("Error occurred in GetUserProjectRoles:", err)
			pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		//return empty role list to indicate if a user is not a member
		result := make(map[string]interface{})
		user, err := dao.GetUser(models.User{UserID: pma.memberID})
		if err != nil {
			beego.Error("Error occurred in GetUser:", err)
			pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		result["user_name"] = user.Username
		result["user_id"] = pma.memberID
		result["roles"] = roleList
		pma.Data["json"] = result
	}
	pma.ServeJSON()
}

func (pma *ProjectMemberAPI) Post() {
	pid := pma.project.ProjectID
	userQuery := models.User{UserID: pma.currentUserID, RoleID: models.PROJECTADMIN}
	rolelist, err := dao.GetUserProjectRoles(userQuery, pid)
	if err != nil {
		beego.Error("Error occurred in GetUserProjectRoles:", err)
		pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if len(rolelist) == 0 {
		beego.Warning("Current user, id:", pma.currentUserID, "does not have project admin role for project, id:", pid)
		pma.RenderError(http.StatusForbidden, "")
		return
	}
	var req memberReq
	pma.DecodeJSONReq(&req)
	username := req.Username
	userID := CheckUserExists(username)
	if userID <= 0 {
		beego.Warning("User does not exist, user name:", username)
		pma.RenderError(http.StatusNotFound, "User does not exist")
		return
	}
	rolelist, err = dao.GetUserProjectRoles(models.User{UserID: userID}, pid)
	if err != nil {
		beego.Error("Error occurred in GetUserProjectRoles:", err)
		pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if len(rolelist) > 0 {
		beego.Warning("user is already added to project, user id:", userID, ", project id:", pid)
		pma.RenderError(http.StatusConflict, "user is ready in project")
		return
	}

	for _, rid := range req.Roles {
		err = dao.AddUserProjectRole(userID, pid, int(rid))
		if err != nil {
			beego.Error("Failed to update DB to add project user role, project id:", pid, ", user id:", userID, ", role id:", rid)
			pma.RenderError(http.StatusInternalServerError, "Failed to update data in database")
			return
		}
	}
}

func (pma *ProjectMemberAPI) Put() {
	pid := pma.project.ProjectID
	mid := pma.memberID
	userQuery := models.User{UserID: pma.currentUserID, RoleID: models.PROJECTADMIN}
	rolelist, err := dao.GetUserProjectRoles(userQuery, pid)
	if err != nil {
		beego.Error("Error occurred in GetUserProjectRoles:", err)
		pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if len(rolelist) == 0 {
		beego.Warning("Current user, id:", pma.currentUserID, ", does not have project admin role for project, id:", pid)
		pma.RenderError(http.StatusForbidden, "")
		return
	}
	var req memberReq
	pma.DecodeJSONReq(&req)
	roleList, err := dao.GetUserProjectRoles(models.User{UserID: mid}, pid)
	if len(roleList) == 0 {
		beego.Warning("User is not in project, user id:", mid, ", project id:", pid)
		pma.RenderError(http.StatusNotFound, "user not exist in project")
		return
	}
	//TODO: delete and insert should in one transaction
	//delete user project role record for the given user
	err = dao.DeleteUserProjectRoles(mid, pid)
	if err != nil {
		beego.Error("Failed to delete project roles for user, user id:", mid, ", project id: ", pid, ", error: ", err)
		pma.RenderError(http.StatusInternalServerError, "Failed to update data in DB")
		return
	}
	//insert roles in request
	for _, rid := range req.Roles {
		err = dao.AddUserProjectRole(mid, pid, int(rid))
		if err != nil {
			beego.Error("Failed to update DB to add project user role, project id:", pid, ", user id:", mid, ", role id:", rid)
			pma.RenderError(http.StatusInternalServerError, "Failed to update data in database")
			return
		}
	}
}

func (pma *ProjectMemberAPI) Delete() {
	pid := pma.project.ProjectID
	mid := pma.memberID
	userQuery := models.User{UserID: pma.currentUserID, RoleID: models.PROJECTADMIN}
	rolelist, err := dao.GetUserProjectRoles(userQuery, pid)
	if len(rolelist) == 0 {
		beego.Warning("Current user, id:", pma.currentUserID, ", does not have project admin role for project, id:", pid)
		pma.RenderError(http.StatusForbidden, "")
		return
	}
	err = dao.DeleteUserProjectRoles(mid, pid)
	if err != nil {
		beego.Error("Failed to delete project roles for user, user id:", mid, ", project id:", pid, ", error:", err)
		pma.RenderError(http.StatusInternalServerError, "Failed to update data in DB")
		return
	}
}
