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
	"github.com/vmware/harbor/utils/log"
)

// ProjectMemberAPI handles request to /api/projects/{}/members/{}
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

// Prepare validates the URL and parms
func (pma *ProjectMemberAPI) Prepare() {
	pid, err := strconv.ParseInt(pma.Ctx.Input.Param(":pid"), 10, 64)
	if err != nil {
		log.Errorf("Error parsing project id: %d, error: %v", pid, err)
		pma.CustomAbort(http.StatusBadRequest, "invalid project Id")
		return
	}
	p, err := dao.GetProjectByID(pid)
	if err != nil {
		log.Errorf("Error occurred in GetProjectById, error: %v", err)
		pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if p == nil {
		log.Warningf("Project with id: %d does not exist.", pid)
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
			log.Errorf("Invalid member Id, error: %v", err)
			pma.CustomAbort(http.StatusBadRequest, "Invalid member id")
		}
		pma.memberID = memberID
	}
}

// Get ...
func (pma *ProjectMemberAPI) Get() {
	pid := pma.project.ProjectID
	if !checkProjectPermission(pma.currentUserID, pid) {
		log.Warningf("Current user, user id: %d does not have permission for project, id: %d", pma.currentUserID, pid)
		pma.RenderError(http.StatusForbidden, "")
		return
	}
	if pma.memberID == 0 { //member id not set return list of the members
		username := pma.GetString("username")
		queryUser := models.User{Username: "%" + username + "%"}
		userList, err := dao.GetUserByProject(pid, queryUser)
		if err != nil {
			log.Errorf("Failed to query database for member list, error: %v", err)
			pma.RenderError(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		pma.Data["json"] = userList
	} else { //return detail of a  member
		roleList, err := dao.GetUserProjectRoles(pma.memberID, pid)
		if err != nil {
			log.Errorf("Error occurred in GetUserProjectRoles, error: %v", err)
			pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		//return empty role list to indicate if a user is not a member
		result := make(map[string]interface{})
		user, err := dao.GetUser(models.User{UserID: pma.memberID})
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		result["user_name"] = user.Username
		result["user_id"] = pma.memberID
		result["roles"] = roleList
		pma.Data["json"] = result
	}
	pma.ServeJSON()
}

// Post ...
func (pma *ProjectMemberAPI) Post() {
	pid := pma.project.ProjectID

	//userQuery := models.User{UserID: pma.currentUserID, RoleID: models.PROJECTADMIN}
	rolelist, err := dao.GetUserProjectRoles(pma.currentUserID, pid)
	if err != nil {
		log.Errorf("Error occurred in GetUserProjectRoles, error: %v", err)
		pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	hasProjectAdminRole := false
	for _, role := range rolelist {
		if role.RoleID == models.PROJECTADMIN {
			hasProjectAdminRole = true
			break
		}
	}
	if !hasProjectAdminRole {
		log.Warningf("Current user, id: %d does not have project admin role for project, id:", pma.currentUserID, pid)
		pma.RenderError(http.StatusForbidden, "")
		return
	}

	var req memberReq
	pma.DecodeJSONReq(&req)
	username := req.Username
	userID := checkUserExists(username)
	if userID <= 0 {
		log.Warningf("User does not exist, user name: %s", username)
		pma.RenderError(http.StatusNotFound, "User does not exist")
		return
	}
	rolelist, err = dao.GetUserProjectRoles(userID, pid)
	if err != nil {
		log.Errorf("Error occurred in GetUserProjectRoles, error: %v", err)
		pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if len(rolelist) > 0 {
		log.Warningf("user is already added to project, user id: %d, project id: %d", userID, pid)
		pma.RenderError(http.StatusConflict, "user is ready in project")
		return
	}

	for _, rid := range req.Roles {
		err = dao.AddProjectMember(pid, userID, int(rid))
		if err != nil {
			log.Errorf("Failed to update DB to add project user role, project id: %d, user id: %d, role id: %d", pid, userID, rid)
			pma.RenderError(http.StatusInternalServerError, "Failed to update data in database")
			return
		}
	}
}

// Put ...
func (pma *ProjectMemberAPI) Put() {
	pid := pma.project.ProjectID
	mid := pma.memberID

	rolelist, err := dao.GetUserProjectRoles(pma.currentUserID, pid)
	if err != nil {
		log.Errorf("Error occurred in GetUserProjectRoles, error: %v", err)
		pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	hasProjectAdminRole := false
	for _, role := range rolelist {
		if role.RoleID == models.PROJECTADMIN {
			hasProjectAdminRole = true
			break
		}
	}

	if !hasProjectAdminRole {
		log.Warningf("Current user, id: %d does not have project admin role for project, id: %d", pma.currentUserID, pid)
		pma.RenderError(http.StatusForbidden, "")
		return
	}
	var req memberReq
	pma.DecodeJSONReq(&req)
	roleList, err := dao.GetUserProjectRoles(mid, pid)
	if len(roleList) == 0 {
		log.Warningf("User is not in project, user id: %d, project id: %d", mid, pid)
		pma.RenderError(http.StatusNotFound, "user not exist in project")
		return
	}
	//TODO: delete and insert should in one transaction
	//delete user project role record for the given user
	err = dao.DeleteProjectMember(pid, mid)
	if err != nil {
		log.Errorf("Failed to delete project roles for user, user id: %d, project id: %d, error: %v", mid, pid, err)
		pma.RenderError(http.StatusInternalServerError, "Failed to update data in DB")
		return
	}
	//insert roles in request
	for _, rid := range req.Roles {
		err = dao.AddProjectMember(pid, mid, int(rid))
		if err != nil {
			log.Errorf("Failed to update DB to add project user role, project id: %d, user id: %d, role id: %d", pid, mid, rid)
			pma.RenderError(http.StatusInternalServerError, "Failed to update data in database")
			return
		}
	}
}

// Delete ...
func (pma *ProjectMemberAPI) Delete() {
	pid := pma.project.ProjectID
	mid := pma.memberID

	rolelist, err := dao.GetUserProjectRoles(pma.currentUserID, pid)
	hasProjectAdminRole := false
	for _, role := range rolelist {
		if role.RoleID == models.PROJECTADMIN {
			hasProjectAdminRole = true
			break
		}
	}

	if !hasProjectAdminRole {
		log.Warningf("Current user, id: %d does not have project admin role for project, id: %d", pma.currentUserID, pid)
		pma.RenderError(http.StatusForbidden, "")
		return
	}
	err = dao.DeleteProjectMember(pid, mid)
	if err != nil {
		log.Errorf("Failed to delete project roles for user, user id: %d, project id: %d, error: %v", mid, pid, err)
		pma.RenderError(http.StatusInternalServerError, "Failed to update data in DB")
		return
	}
}
