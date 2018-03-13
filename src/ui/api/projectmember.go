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
	"github.com/vmware/harbor/src/common/dao/project"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// ProjectMemberAPI handles request to /api/projects/{}/members/{}
type ProjectMemberAPI struct {
	BaseController
	id            int
	entityID      int
	entityType    string
	currentUserID int
	project       *models.Project
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

	pmid, err := pma.GetInt64FromPath(":pmid")
	if err != nil {
		log.Errorf("Failed to get pmid from path, error %v", err)
	}
	if pmid > 0 {
		pma.id = int(pmid)
	}
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
			log.Errorf("Failed to query database for member list, error: %v", err)
			pma.RenderError(http.StatusInternalServerError, "Internal Server Error")
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
			log.Errorf("Failed to query database for member list, error: %v", err)
			pma.RenderError(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		if len(memberList) > 0 {
			pma.Data["json"] = memberList[0]
		}
	}
	pma.ServeJSON()
}

// Post ... Add a project member
func (pma *ProjectMemberAPI) Post() {
	projectID := pma.project.ProjectID
	var request models.MemberReq
	pma.DecodeJSONReq(&request)
	request.ProjectID = projectID
	_, err := project.AddProjectMember(request)
	if err != nil {
		log.Errorf("Failed to add project member, error: %v", err)
		pma.RenderError(http.StatusInternalServerError, "Failed to add project member")
		return
	}
}

// Put ... Update an exist project member
func (pma *ProjectMemberAPI) Put() {
	pid := pma.project.ProjectID
	pmID := pma.id
	if pmID == 0 {
		log.Errorf("Failed to update DB to add project user role, project id: %d, pmid : %d", pid, pmID)
		pma.RenderError(http.StatusInternalServerError, "Failed to update data in DB")
		return
	}
	var req models.Member
	pma.DecodeJSONReq(&req)
	err := project.UpdateProjectMember(pmID, req.Role)
	if err != nil {
		log.Errorf("Failed to update DB to add project user role, project id: %d, pmid : %d, role id: %d", pid, pmID, req.Role)
		pma.RenderError(http.StatusInternalServerError, "Failed to update data in DB")
		return
	}
}

// Delete ...
func (pma *ProjectMemberAPI) Delete() {
	pmid := pma.id
	err := project.DeleteProjectMemberByID(pmid)
	if err != nil {
		log.Errorf("Failed to delete project roles for user, project member id: %d, error: %v", pmid, err)
		pma.RenderError(http.StatusInternalServerError, "Failed to update data in DB")
		return
	}
}
