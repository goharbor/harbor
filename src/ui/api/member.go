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
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// ProjectMemberAPI handles request to /api/projects/{}/members/{}
type ProjectMemberAPI struct {
	BaseController
	memberID      int
	currentUserID int
	project       *models.Project
}

type memberReq struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	Roles    []int  `json:"roles"`
}

// Prepare validates the URL and parms
func (pma *ProjectMemberAPI) Prepare() {
	pma.BaseController.Prepare()

	if !pma.SecurityCtx.IsAuthenticated() {
		pma.HandleUnauthorized()
		return
	}
	user, err := dao.GetUser(models.User{
		Username: pma.SecurityCtx.GetUsername(),
	})
	if err != nil {
		pma.HandleInternalServerError(
			fmt.Sprintf("failed to get user %s: %v",
				pma.SecurityCtx.GetUsername(), err))
		return
	}
	pma.currentUserID = user.UserID

	pid, err := pma.GetInt64FromPath(":pid")
	if err != nil || pid <= 0 {
		text := "invalid project ID: "
		if err != nil {
			text += err.Error()
		} else {
			text += fmt.Sprintf("%d", pid)
		}
		pma.HandleBadRequest(text)
		return
	}
	project, err := pma.ProjectMgr.Get(pid)
	if err != nil {
		pma.ParseAndHandleError(fmt.Sprintf("failed to get project %d", pid), err)
		return
	}
	if project == nil {
		pma.HandleNotFound(fmt.Sprintf("project %d not found", pid))
		return
	}
	pma.project = project

	if !(pma.Ctx.Input.IsGet() && pma.SecurityCtx.HasReadPerm(pid) ||
		pma.SecurityCtx.HasAllPerm(pid)) {
		pma.HandleForbidden(pma.SecurityCtx.GetUsername())
		return
	}

	if len(pma.GetStringFromPath(":mid")) != 0 {
		mid, err := pma.GetInt64FromPath(":mid")
		if err != nil || mid <= 0 {
			text := "invalid member ID: "
			if err != nil {
				text += err.Error()
			} else {
				text += fmt.Sprintf("%d", mid)
			}
			pma.HandleBadRequest(text)
			return
		}

		member, err := dao.GetUser(models.User{
			UserID: int(mid),
		})
		if err != nil {
			pma.HandleInternalServerError(fmt.Sprintf("failed to get user %d: %v", mid, err))
			return
		}
		if member == nil {
			pma.HandleNotFound(fmt.Sprintf("member %d not found", mid))
			return
		}

		pma.memberID = member.UserID
	}
}

// Get ...
func (pma *ProjectMemberAPI) Get() {
	pid := pma.project.ProjectID
	if pma.memberID == 0 { //member id not set return list of the members
		username := pma.GetString("username")
		queryUser := models.User{Username: username}
		userList, err := dao.GetUserByProject(pid, queryUser)
		if err != nil {
			log.Errorf("Failed to query database for member list, error: %v", err)
			pma.RenderError(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		pma.Data["json"] = userList
	} else { //return detail of a  member
		roleList, err := listRoles(pma.memberID, pid)
		if err != nil {
			log.Errorf("Error occurred in GetUserProjectRoles, error: %v", err)
			pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}

		if len(roleList) == 0 {
			pma.CustomAbort(http.StatusNotFound, fmt.Sprintf("user %d is not a member of the project", pma.memberID))
		}

		//return empty role list to indicate if a user is not a member
		result := make(map[string]interface{})
		user, err := dao.GetUser(models.User{UserID: pma.memberID})
		if err != nil {
			log.Errorf("Error occurred in GetUser, error: %v", err)
			pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		result["username"] = user.Username
		result["user_id"] = pma.memberID
		result["roles"] = roleList
		pma.Data["json"] = result
	}
	pma.ServeJSON()
}

// Post ...
func (pma *ProjectMemberAPI) Post() {
	projectID := pma.project.ProjectID

	var req memberReq
	pma.DecodeJSONReq(&req)
	username := req.Username
	userID := checkUserExists(username)
	if userID <= 0 {
		log.Warningf("User does not exist, user name: %s", username)
		pma.RenderError(http.StatusNotFound, "User does not exist")
		return
	}
	rolelist, err := dao.GetUserProjectRoles(userID, projectID)
	if err != nil {
		log.Errorf("Error occurred in GetUserProjectRoles, error: %v", err)
		pma.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if len(rolelist) > 0 {
		log.Warningf("user is already added to project, user id: %d, project id: %d", userID, projectID)
		pma.RenderError(http.StatusConflict, "user is ready in project")
		return
	}

	if len(req.Roles) <= 0 || len(req.Roles) > 1 {
		pma.CustomAbort(http.StatusBadRequest, "only one role is supported")
	}

	rid := req.Roles[0]
	if !(rid == models.PROJECTADMIN ||
		rid == models.DEVELOPER ||
		rid == models.GUEST) {
		pma.CustomAbort(http.StatusBadRequest, "invalid role")
	}

	err = dao.AddProjectMember(projectID, userID, rid)
	if err != nil {
		log.Errorf("Failed to update DB to add project user role, project id: %d, user id: %d, role id: %d", projectID, userID, rid)
		pma.RenderError(http.StatusInternalServerError, "Failed to update data in database")
		return
	}
}

// Put ...
func (pma *ProjectMemberAPI) Put() {
	pid := pma.project.ProjectID
	mid := pma.memberID

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

	err := dao.DeleteProjectMember(pid, mid)
	if err != nil {
		log.Errorf("Failed to delete project roles for user, user id: %d, project id: %d, error: %v", mid, pid, err)
		pma.RenderError(http.StatusInternalServerError, "Failed to update data in DB")
		return
	}
}
