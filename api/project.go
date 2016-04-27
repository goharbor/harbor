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
	"fmt"
	"net/http"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"

	"strconv"
	"time"
)

// ProjectAPI handles request to /api/projects/{} /api/projects/{}/logs
type ProjectAPI struct {
	BaseAPI
	userID    int
	projectID int64
}

type projectReq struct {
	ProjectName string `json:"project_name"`
	Public      bool   `json:"public"`
}

const projectNameMaxLen int = 30

// Prepare validates the URL and the user
func (p *ProjectAPI) Prepare() {
	p.userID = p.ValidateUser()
	idStr := p.Ctx.Input.Param(":id")
	if len(idStr) > 0 {
		var err error
		p.projectID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Errorf("Error parsing project id: %s, error: %v", idStr, err)
			p.CustomAbort(http.StatusBadRequest, "invalid project id")
		}
		exist, err := dao.ProjectExists(p.projectID)
		if err != nil {
			log.Errorf("Error occurred in ProjectExists, error: %v", err)
			p.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if !exist {
			p.CustomAbort(http.StatusNotFound, fmt.Sprintf("project does not exist, id: %v", p.projectID))
		}
	}
}

// Post ...
func (p *ProjectAPI) Post() {
	var req projectReq
	var public int
	p.DecodeJSONReq(&req)
	if req.Public {
		public = 1
	}
	err := validateProjectReq(req)
	if err != nil {
		log.Errorf("Invalid project request, error: %v", err)
		p.RenderError(http.StatusBadRequest, "Invalid request for creating project")
		return
	}
	projectName := req.ProjectName
	exist, err := dao.ProjectExists(projectName)
	if err != nil {
		log.Errorf("Error happened checking project existence in db, error: %v, project name: %s", err, projectName)
	}
	if exist {
		p.RenderError(http.StatusConflict, "")
		return
	}
	project := models.Project{OwnerID: p.userID, Name: projectName, CreationTime: time.Now(), Public: public}
	projectID, err := dao.AddProject(project)
	if err != nil {
		log.Errorf("Failed to add project, error: %v", err)
		p.RenderError(http.StatusInternalServerError, "Failed to add project")
	}

	p.Redirect(http.StatusCreated, strconv.FormatInt(projectID, 10))
}

// Head ...
func (p *ProjectAPI) Head() {
	projectName := p.GetString("project_name")
	result, err := dao.ProjectExists(projectName)
	if err != nil {
		log.Errorf("Error while communicating with DB, error: %v", err)
		p.RenderError(http.StatusInternalServerError, "Error while communicating with DB")
		return
	}
	if !result {
		p.RenderError(http.StatusNotFound, "")
		return
	}
}

// Get ...
func (p *ProjectAPI) Get() {
	queryProject := models.Project{UserID: p.userID}
	projectName := p.GetString("project_name")
	if len(projectName) > 0 {
		queryProject.Name = "%" + projectName + "%"
	}
	public, _ := p.GetInt("is_public")
	queryProject.Public = public

	projectList, err := dao.QueryProject(queryProject)
	if err != nil {
		log.Errorf("Error occurred in QueryProject, error: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	for i := 0; i < len(projectList); i++ {
		if isProjectAdmin(p.userID, projectList[i].ProjectID) {
			projectList[i].Togglable = true
		}
	}
	p.Data["json"] = projectList
	p.ServeJSON()
}

// Put ...
func (p *ProjectAPI) Put() {
	var req projectReq
	var public int

	projectID, err := strconv.ParseInt(p.Ctx.Input.Param(":id"), 10, 64)
	if err != nil {
		log.Errorf("Error parsing project id: %d, error: %v", projectID, err)
		p.RenderError(http.StatusBadRequest, "invalid project id")
		return
	}

	p.DecodeJSONReq(&req)
	if req.Public {
		public = 1
	}
	if !isProjectAdmin(p.userID, projectID) {
		log.Warningf("Current user, id: %d does not have project admin role for project, id: %d", p.userID, projectID)
		p.RenderError(http.StatusForbidden, "")
		return
	}
	err = dao.ToggleProjectPublicity(p.projectID, public)
	if err != nil {
		log.Errorf("Error while updating project, project id: %d, error: %v", projectID, err)
		p.RenderError(http.StatusInternalServerError, "Failed to update project")
	}
}

// FilterAccessLog handles GET to /api/projects/{}/logs
func (p *ProjectAPI) FilterAccessLog() {

	var filter models.AccessLog
	p.DecodeJSONReq(&filter)

	username := filter.Username
	keywords := filter.Keywords

	beginTime := time.Unix(filter.BeginTimestamp, 0)
	endTime := time.Unix(filter.EndTimestamp, 0)

	query := models.AccessLog{ProjectID: p.projectID, Username: "%" + username + "%", Keywords: keywords, BeginTime: beginTime, BeginTimestamp: filter.BeginTimestamp, EndTime: endTime, EndTimestamp: filter.EndTimestamp}

	log.Infof("Query AccessLog: begin: %v, end: %v, keywords: %s", query.BeginTime, query.EndTime, query.Keywords)

	accessLogList, err := dao.GetAccessLogs(query)
	if err != nil {
		log.Errorf("Error occurred in GetAccessLogs, error: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	p.Data["json"] = accessLogList
	p.ServeJSON()
}

func isProjectAdmin(userID int, pid int64) bool {
	rolelist, err := dao.GetUserProjectRoles(userID, pid)
	if err != nil {
		log.Errorf("Error occurred in GetUserProjectRoles, returning false, error: %v", err)
		return false
	}

	hasProjectAdminRole := false
	for _, role := range rolelist {
		if role.RoleID == models.PROJECTADMIN {
			hasProjectAdminRole = true
			break
		}
	}

	return hasProjectAdminRole
}

func validateProjectReq(req projectReq) error {
	pn := req.ProjectName
	if len(pn) == 0 {
		return fmt.Errorf("Project name can not be empty")
	}
	if len(pn) > projectNameMaxLen {
		return fmt.Errorf("Project name is too long")
	}
	return nil
}
