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
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/project"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/auth"
)

// ProjectMemberAPI handles request to /api/projects/{}/members/{}
type ProjectMemberAPI struct {
	BaseController
	id         int
	entityID   int
	entityType string
	project    *models.Project
}

// ErrDuplicateProjectMember ...
var ErrDuplicateProjectMember = errors.New("The project member specified already exist")

// ErrInvalidRole ...
var ErrInvalidRole = errors.New("Failed to update project member, role is not in 1,2,3")

// Prepare validates the URL and parms
func (pma *ProjectMemberAPI) Prepare() {
	pma.BaseController.Prepare()

	if !pma.SecurityCtx.IsAuthenticated() {
		pma.SendUnAuthorizedError(errors.New("Unauthorized"))
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
		pma.SendBadRequestError(errors.New(text))
		return
	}
	project, err := pma.ProjectMgr.Get(pid)
	if err != nil {
		pma.ParseAndHandleError(fmt.Sprintf("failed to get project %d", pid), err)
		return
	}
	if project == nil {
		pma.SendNotFoundError(fmt.Errorf("project %d not found", pid))
		return
	}
	pma.project = project

	pmid, err := pma.GetInt64FromPath(":pmid")
	if err != nil {
		log.Warningf("Failed to get pmid from path, error %v", err)
	}
	if pmid <= 0 && (pma.Ctx.Input.IsPut() || pma.Ctx.Input.IsDelete()) {
		pma.SendBadRequestError(fmt.Errorf("The project member id is invalid, pmid:%s", pma.GetStringFromPath(":pmid")))
		return
	}
	pma.id = int(pmid)
}

func (pma *ProjectMemberAPI) requireAccess(action rbac.Action) bool {
	resource := rbac.NewProjectNamespace(pma.project.ProjectID).Resource(rbac.ResourceMember)

	if !pma.SecurityCtx.Can(action, resource) {
		if !pma.SecurityCtx.IsAuthenticated() {
			pma.SendUnAuthorizedError(errors.New("Unauthorized"))
		} else {
			pma.SendForbiddenError(errors.New(pma.SecurityCtx.GetUsername()))
		}

		return false
	}

	return true
}

// Get ...
func (pma *ProjectMemberAPI) Get() {
	projectID := pma.project.ProjectID
	queryMember := models.Member{}
	queryMember.ProjectID = projectID
	pma.Data["json"] = make([]models.Member, 0)
	if pma.id == 0 {
		if !pma.requireAccess(rbac.ActionList) {
			return
		}
		entityname := pma.GetString("entityname")
		memberList, err := project.SearchMemberByName(projectID, entityname)
		if err != nil {
			pma.SendInternalServerError(fmt.Errorf("Failed to query database for member list, error: %v", err))
			return
		}
		if len(memberList) > 0 {
			pma.Data["json"] = memberList
		}

	} else {
		// return a specific member
		queryMember.ID = pma.id
		memberList, err := project.GetProjectMember(queryMember)
		if err != nil {
			pma.SendInternalServerError(fmt.Errorf("Failed to query database for member list, error: %v", err))
			return
		}
		if len(memberList) == 0 {
			pma.SendNotFoundError(fmt.Errorf("The project member does not exit, pmid:%v", pma.id))
			return
		}

		if !pma.requireAccess(rbac.ActionRead) {
			return
		}
		pma.Data["json"] = memberList[0]
	}
	pma.ServeJSON()
}

// Post ... Add a project member
func (pma *ProjectMemberAPI) Post() {
	if !pma.requireAccess(rbac.ActionCreate) {
		return
	}
	projectID := pma.project.ProjectID
	var request models.MemberReq
	if err := pma.DecodeJSONReq(&request); err != nil {
		pma.SendBadRequestError(err)
		return
	}
	request.MemberGroup.LdapGroupDN = strings.TrimSpace(request.MemberGroup.LdapGroupDN)

	pmid, err := AddProjectMember(projectID, request)
	if err == auth.ErrorGroupNotExist || err == auth.ErrorUserNotExist {
		pma.SendBadRequestError(fmt.Errorf("Failed to add project member, error: %v", err))
		return
	} else if err == auth.ErrDuplicateLDAPGroup {
		pma.SendConflictError(fmt.Errorf("Failed to add project member, already exist LDAP group or project member, groupDN:%v", request.MemberGroup.LdapGroupDN))
		return
	} else if err == ErrDuplicateProjectMember {
		pma.SendConflictError(fmt.Errorf("Failed to add project member, already exist LDAP group or project member, groupMemberID:%v", request.MemberGroup.ID))
		return
	} else if err == ErrInvalidRole {
		pma.SendBadRequestError(fmt.Errorf("Invalid role ID, role ID %v", request.Role))
		return
	} else if err == auth.ErrInvalidLDAPGroupDN {
		pma.SendBadRequestError(fmt.Errorf("Invalid LDAP DN: %v", request.MemberGroup.LdapGroupDN))
		return
	} else if err != nil {
		pma.SendInternalServerError(fmt.Errorf("Failed to add project member, error: %v", err))
		return
	}
	pma.Redirect(http.StatusCreated, strconv.FormatInt(int64(pmid), 10))
}

// Put ... Update an exist project member
func (pma *ProjectMemberAPI) Put() {
	if !pma.requireAccess(rbac.ActionUpdate) {
		return
	}
	pid := pma.project.ProjectID
	pmID := pma.id
	var req models.Member
	if err := pma.DecodeJSONReq(&req); err != nil {
		pma.SendBadRequestError(err)
		return
	}
	if req.Role < 1 || req.Role > 4 {
		pma.SendBadRequestError(fmt.Errorf("Invalid role id %v", req.Role))
		return
	}
	err := project.UpdateProjectMemberRole(pmID, req.Role)
	if err != nil {
		pma.SendInternalServerError(fmt.Errorf("Failed to update DB to add project user role, project id: %d, pmid : %d, role id: %d", pid, pmID, req.Role))
		return
	}
}

// Delete ...
func (pma *ProjectMemberAPI) Delete() {
	if !pma.requireAccess(rbac.ActionDelete) {
		return
	}
	pmid := pma.id
	err := project.DeleteProjectMemberByID(pmid)
	if err != nil {
		pma.SendInternalServerError(fmt.Errorf("Failed to delete project roles for user, project member id: %d, error: %v", pmid, err))
		return
	}
}

// AddProjectMember ...
func AddProjectMember(projectID int64, request models.MemberReq) (int, error) {
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
		var userID int
		member.EntityType = common.UserMember
		u, err := dao.GetUser(models.User{Username: request.MemberUser.Username})
		if err != nil {
			return 0, err
		}
		if u != nil {
			userID = u.UserID
		} else {
			userID, err = auth.SearchAndOnBoardUser(request.MemberUser.Username)
			if err != nil {
				return 0, err
			}
		}
		member.EntityID = userID
	} else if len(request.MemberGroup.LdapGroupDN) > 0 {

		// If groupname provided, use the provided groupname to name this group
		groupID, err := auth.SearchAndOnBoardGroup(request.MemberGroup.LdapGroupDN, request.MemberGroup.GroupName)
		if err != nil {
			return 0, err
		}
		member.EntityID = groupID
		member.EntityType = common.GroupMember
	}
	if member.EntityID <= 0 {
		return 0, fmt.Errorf("Can not get valid member entity, request: %+v", request)
	}

	// Check if member already exist in current project
	memberList, err := project.GetProjectMember(models.Member{
		ProjectID:  member.ProjectID,
		EntityID:   member.EntityID,
		EntityType: member.EntityType,
	})
	if err != nil {
		return 0, err
	}
	if len(memberList) > 0 {
		return 0, ErrDuplicateProjectMember
	}

	if member.Role < 1 || member.Role > 4 {
		// Return invalid role error
		return 0, ErrInvalidRole
	}
	return project.AddProjectMember(member)
}
