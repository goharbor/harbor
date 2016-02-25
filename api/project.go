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
	"log"
	"net/http"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"strconv"
	"time"

	"github.com/astaxie/beego"
)

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

func (p *ProjectAPI) Prepare() {
	p.userID = p.ValidateUser()
	idStr := p.Ctx.Input.Param(":id")
	if len(idStr) > 0 {
		var err error
		p.projectID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing project id: %s, error: %v", idStr, err)
			p.CustomAbort(http.StatusBadRequest, "invalid project id")
		}
		exist, err := dao.ProjectExists(p.projectID)
		if err != nil {
			log.Printf("Error occurred in ProjectExists: %v", err)
			p.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if !exist {
			p.CustomAbort(http.StatusNotFound, fmt.Sprintf("project does not exist, id: %v", p.projectID))
		}
	}
}

func (p *ProjectAPI) Post() {
	var req projectReq
	var public int
	p.DecodeJSONReq(&req)
	if req.Public {
		public = 1
	}
	err := validateProjectReq(req)
	if err != nil {
		beego.Error("Invalid project request, error: ", err)
		p.RenderError(http.StatusBadRequest, "Invalid request for creating project")
		return
	}
	projectName := req.ProjectName
	exist, err := dao.ProjectExists(projectName)
	if err != nil {
		beego.Error("Error happened checking project existence in db:", err, ", project name:", projectName)
	}
	if exist {
		p.RenderError(http.StatusConflict, "")
		return
	}
	project := models.Project{OwnerId: p.userID, Name: projectName, CreationTime: time.Now(), Public: public}
	err = dao.AddProject(project)
	if err != nil {
		beego.Error("Failed to add project, error: %v", err)
		p.RenderError(http.StatusInternalServerError, "Failed to add project")
	}
}

func (p *ProjectAPI) Head() {
	projectName := p.GetString("project_name")
	result, err := dao.ProjectExists(projectName)
	if err != nil {
		beego.Error("Error while communicating with DB: ", err)
		p.RenderError(http.StatusInternalServerError, "Error while communicating with DB")
		return
	}
	if !result {
		p.RenderError(http.StatusNotFound, "")
		return
	}
}

func (p *ProjectAPI) Get() {
	queryProject := models.Project{UserId: p.userID}
	projectName := p.GetString("project_name")
	if len(projectName) > 0 {
		queryProject.Name = "%" + projectName + "%"
	}
	public, _ := p.GetInt("is_public")
	queryProject.Public = public

	projectList, err := dao.QueryProject(queryProject)
	if err != nil {
		beego.Error("Error occurred in QueryProject:", err)
		p.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	for i := 0; i < len(projectList); i++ {
		if isProjectAdmin(p.userID, projectList[i].ProjectId) {
			projectList[i].Togglable = true
		}
	}
	p.Data["json"] = projectList
	p.ServeJSON()
}

func (p *ProjectAPI) Put() {
	var req projectReq
	var public int

	projectID, err := strconv.ParseInt(p.Ctx.Input.Param(":id"), 10, 64)
	if err != nil {
		beego.Error("Error parsing project id:", projectID, ", error: ", err)
		p.RenderError(http.StatusBadRequest, "invalid project id")
		return
	}

	p.DecodeJSONReq(&req)
	if req.Public {
		public = 1
	}
	if !isProjectAdmin(p.userID, projectID) {
		beego.Warning("Current user, id:", p.userID, ", does not have project admin role for project, id:", projectID)
		p.RenderError(http.StatusForbidden, "")
		return
	}
	err = dao.ToggleProjectPublicity(p.projectID, public)
	if err != nil {
		beego.Error("Error while updating project, project id:", projectID, ", error:", err)
		p.RenderError(http.StatusInternalServerError, "Failed to update project")
	}
}

func (p *ProjectAPI) FilterAccessLog() {

	var filter models.AccessLog
	p.DecodeJSONReq(&filter)

	username := filter.Username
	keywords := filter.Keywords

	beginTime := time.Unix(filter.BeginTimestamp, 0)
	endTime := time.Unix(filter.EndTimestamp, 0)

	query := models.AccessLog{ProjectId: p.projectID, Username: "%" + username + "%", Keywords: keywords, BeginTime: beginTime, BeginTimestamp: filter.BeginTimestamp, EndTime: endTime, EndTimestamp: filter.EndTimestamp}

	log.Printf("Query AccessLog: begin: %v, end: %v, keywords: %s", query.BeginTime, query.EndTime, query.Keywords)

	accessLogList, err := dao.GetAccessLogs(query)
	if err != nil {
		log.Printf("Error occurred in GetAccessLogs: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	p.Data["json"] = accessLogList
	p.ServeJSON()
}

func isProjectAdmin(userID int, pid int64) bool {
	userQuery := models.User{UserId: userID, RoleId: models.PROJECTADMIN}
	rolelist, err := dao.GetUserProjectRoles(userQuery, pid)
	if err != nil {
		beego.Error("Error occurred in GetUserProjectRoles:", err, ", returning false")
		return false
	}
	return len(rolelist) > 0
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
