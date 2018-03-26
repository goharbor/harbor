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
	"strconv"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao/project"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"
)

// ProjectMemberAPI handles request to /api/projects/{}/members/{}
type ProjectMemberAPI struct {
	BaseController
	id         int
	entityID   int
	entityType string
	project    *models.Project
}

// Prepare validates the URL and parms
func (pma *ProjectMemberAPI) Prepare() {
	pma.BaseController.Prepare()

	if !pma.SecurityCtx.IsAuthenticated() {
		pma.HandleUnauthorized()
		return
	}
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

	pmid, err := pma.GetInt64FromPath(":pmid")
	if err != nil {
		log.Errorf("Failed to get pmid from path, error %v", err)
	}
	if pmid <= 0 && (pma.Ctx.Input.IsPut() || pma.Ctx.Input.IsDelete()) {
		pma.HandleBadRequest(fmt.Sprintf("The project member id is invalid, pmid:%s", pma.GetStringFromPath(":pmid")))
		return
	}
	pma.id = int(pmid)
}

//Get ...
func (pma *ProjectMemberAPI) Get() {
	projectID := pma.project.ProjectID
	queryMember := models.Member{}
	queryMember.ProjectID = projectID
	pma.Data["json"] = make([]models.Member, 0)
	if pma.id == 0 {
		//member id not set, return all member of current project
		memberList, err := project.GetProjectMember(queryMember)
		if err != nil {
			pma.HandleInternalServerError(fmt.Sprintf("Failed to query database for member list, error: %v", err))
			return
		}
		if len(memberList) > 0 {
			pma.Data["json"] = memberList
		}
	} else {
		//return a specific member
		queryMember.ID = pma.id
		memberList, err := project.GetProjectMember(queryMember)
		if err != nil {
			pma.HandleInternalServerError(fmt.Sprintf("Failed to query database for member list, error: %v", err))
			return
		}
		if len(memberList) == 0 {
			pma.HandleNotFound(fmt.Sprintf("The project member does not exit, pmid:%v", pma.id))
			return
		}
		pma.Data["json"] = memberList[0]
	}
	pma.ServeJSON()
}

// Post ... Add a project member
func (pma *ProjectMemberAPI) Post() {
	projectID := pma.project.ProjectID
	var request models.MemberReq
	pma.DecodeJSONReq(&request)
	pmid, err := AddOrUpdateProjectMember(projectID, request)
	if err != nil {
		pma.HandleInternalServerError(fmt.Sprintf("Failed to add project member, error: %v", err))
		return
	}
	pma.Redirect(http.StatusCreated, strconv.FormatInt(int64(pmid), 10))
}

// Put ... Update an exist project member
func (pma *ProjectMemberAPI) Put() {
	pid := pma.project.ProjectID
	pmID := pma.id
	var req models.Member
	pma.DecodeJSONReq(&req)
	if req.Role < 1 || req.Role > 3 {
		pma.HandleBadRequest(fmt.Sprintf("Invalid role id %v", req.Role))
		return
	}
	err := project.UpdateProjectMemberRole(pmID, req.Role)
	if err != nil {
		pma.HandleInternalServerError(fmt.Sprintf("Failed to update DB to add project user role, project id: %d, pmid : %d, role id: %d", pid, pmID, req.Role))
		return
	}
}

// Delete ...
func (pma *ProjectMemberAPI) Delete() {
	pmid := pma.id
	err := project.DeleteProjectMemberByID(pmid)
	if err != nil {
		pma.HandleInternalServerError(fmt.Sprintf("Failed to delete project roles for user, project member id: %d, error: %v", pmid, err))
		return
	}
}

// AddOrUpdateProjectMember ... If the project member relationship does not exist, create it. if exist, update it
func AddOrUpdateProjectMember(projectID int64, request models.MemberReq) (int, error) {
	var member models.Member
	member.ProjectID = projectID
	member.Role = request.Role
	if request.MemberUser.UserID > 0 {
		member.EntityID = request.MemberUser.UserID
		member.EntityType = common.UserMember
	} else if request.MemberGroup.ID > 0 {
		member.EntityID = request.MemberGroup.ID
		member.EntityType = common.GroupMember
	} else if len(request.MemberUser.Username) > 0 {
		member.EntityType = common.UserMember
		userID, err := auth.SearchAndOnBoardUser(request.MemberUser.Username)
		if err != nil {
			return 0, err
		}
		member.EntityID = userID
	} else if len(request.MemberGroup.LdapGroupDN) > 0 {
		member.EntityType = common.GroupMember
		//If groupname provided, use the provided groupname
		//If ldap group already exist in harbor, use the previous group name
		groupID, err := auth.SearchAndOnBoardGroup(request.MemberGroup.LdapGroupDN, request.MemberGroup.GroupName)
		if err != nil {
			return 0, err
		}
		member.EntityID = groupID
	}
	if member.EntityID <= 0 {
		return 0, fmt.Errorf("Can not get valid member entity, request: %+v", request)
	}
	memberList, err := project.GetProjectMember(models.Member{
		ProjectID:  member.ProjectID,
		EntityID:   member.EntityID,
		EntityType: member.EntityType,
	})
	if err != nil {
		return 0, err
	}
	if len(memberList) > 0 {
		project.UpdateProjectMemberRole(memberList[0].ID, member.Role)
		return 0, nil
	}

	if member.Role < 1 || member.Role > 3 {
		return 0, fmt.Errorf("Failed to update project member, role is not in 1,2,3 role:%v", member.Role)
	}
	return project.AddProjectMember(member)
}
