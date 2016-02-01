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

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"strconv"
	"time"

	"github.com/astaxie/beego"
)

type ProjectAPI struct {
	BaseAPI
	userId    int
	projectId int64
}

type projectReq struct {
	ProjectName string `json:"project_name"`
	Public      bool   `json:"public"`
}

func (p *ProjectAPI) Prepare() {
	p.userId = p.ValidateUser()
	id_str := p.Ctx.Input.Param(":id")
	if len(id_str) > 0 {
		var err error
		p.projectId, err = strconv.ParseInt(id_str, 10, 64)
		if err != nil {
			log.Printf("Error parsing project id: %s, error: %v", id_str, err)
			p.CustomAbort(400, "invalid project id")
		}
		proj, err := dao.GetProjectById(models.Project{ProjectId: p.projectId})
		if err != nil {
			log.Printf("Error occurred in GetProjectById: %v", err)
			p.CustomAbort(500, "Internal error.")
		}
		if proj == nil {
			p.CustomAbort(404, fmt.Sprintf("project does not exist, id: %v", p.projectId))
		}
	}
}

func (p *ProjectAPI) Post() {
	var req projectReq
	var public int
	p.DecodeJsonReq(&req)
	if req.Public {
		public = 1
	}
	projectName := req.ProjectName
	exist, err := dao.ProjectExists(projectName)
	if err != nil {
		beego.Error("Error happened checking project existence in db:", err, ", project name:", projectName)
	}
	if exist {
		p.RenderError(409, "")
		return
	}
	project := models.Project{OwnerId: p.userId, Name: projectName, CreationTime: time.Now(), Public: public}
	dao.AddProject(project)
}

func (p *ProjectAPI) Head() {
	projectName := p.GetString("project_name")
	result, err := dao.ProjectExists(projectName)
	if err != nil {
		p.RenderError(500, "Error while communicating with DB")
		return
	}
	if !result {
		p.RenderError(404, "")
		return
	}
}

func (p *ProjectAPI) Get() {
	queryProject := models.Project{UserId: p.userId}
	projectName := p.GetString("project_name")
	if len(projectName) > 0 {
		queryProject.Name = "%" + projectName + "%"
	}
	public, _ := p.GetInt("is_public")
	queryProject.Public = public

	projectList, err := dao.QueryProject(queryProject)
	if err != nil {
		beego.Error("Error occurred in QueryProject:", err)
		p.CustomAbort(500, "Internal error.")
	}
	for i := 0; i < len(projectList); i++ {
		if isProjectAdmin(p.userId, projectList[i].ProjectId) {
			projectList[i].Togglable = true
		}
	}
	p.Data["json"] = projectList
	p.ServeJSON()
}

func (p *ProjectAPI) Put() {
	var req projectReq
	var public int

	projectId, err := strconv.ParseInt(p.Ctx.Input.Param(":id"), 10, 64)
	if err != nil {
		beego.Error("Error parsing project id:", projectId, ", error: ", err)
		p.RenderError(400, "invalid project id")
		return
	}

	p.DecodeJsonReq(&req)
	if req.Public {
		public = 1
	}

	proj, err := dao.GetProjectById(models.Project{ProjectId: projectId})
	if err != nil {
		beego.Error("Error occurred in GetProjectById:", err)
		p.CustomAbort(500, "Internal error.")
	}
	if proj == nil {
		p.RenderError(404, "")
		return
	}
	if !isProjectAdmin(p.userId, projectId) {
		beego.Warning("Current user, id:", p.userId, ", does not have project admin role for project, id:", projectId)
		p.RenderError(403, "")
		return
	}
	err = dao.ToggleProjectPublicity(p.projectId, public)
	if err != nil {
		beego.Error("Error while updating project, project id:", projectId, ", error:", err)
		p.RenderError(500, "Failed to update project")
	}
}

func (p *ProjectAPI) ListAccessLog() {

	query := models.AccessLog{ProjectId: p.projectId}
	accessLogList, err := dao.GetAccessLogs(query)
	if err != nil {
		log.Printf("Error occurred in GetAccessLogs: %v", err)
		p.CustomAbort(500, "Internal error.")
	}
	p.Data["json"] = accessLogList
	p.ServeJSON()
}

func (p *ProjectAPI) FilterAccessLog() {

	var filter models.AccessLog
	p.DecodeJsonReq(&filter)

	username := filter.Username
	keywords := filter.Keywords
	beginTime := filter.BeginTime
	endTime := filter.EndTime

	query := models.AccessLog{ProjectId: p.projectId, Username: "%" + username + "%", Keywords: keywords, BeginTime: beginTime, EndTime: endTime}
	accessLogList, err := dao.GetAccessLogs(query)
	if err != nil {
		log.Printf("Error occurred in GetAccessLogs: %v", err)
		p.CustomAbort(500, "Internal error.")
	}
	p.Data["json"] = accessLogList
	p.ServeJSON()
}

func isProjectAdmin(userId int, pid int64) bool {
	userQuery := models.User{UserId: userId, RoleId: models.PROJECTADMIN}
	rolelist, err := dao.GetUserProjectRoles(userQuery, pid)
	if err != nil {
		beego.Error("Error occurred in GetUserProjectRoles:", err, ", returning false")
		return false
	}
	return len(rolelist) > 0
}
