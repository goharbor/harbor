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
	"regexp"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"

	"strconv"
	"time"
)

// ProjectAPI handles request to /api/projects/{} /api/projects/{}/logs
type ProjectAPI struct {
	BaseAPI
	userID      int
	projectID   int64
	projectName string
}

type projectReq struct {
	ProjectName string `json:"project_name"`
	Public      int    `json:"public"`
}

const projectNameMaxLen int = 30
const projectNameMinLen int = 4
const dupProjectPattern = `Duplicate entry '\w+' for key 'name'`

// Prepare validates the URL and the user
func (p *ProjectAPI) Prepare() {
	idStr := p.Ctx.Input.Param(":id")
	if len(idStr) > 0 {
		var err error
		p.projectID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Errorf("Error parsing project id: %s, error: %v", idStr, err)
			p.CustomAbort(http.StatusBadRequest, "invalid project id")
		}

		project, err := dao.GetProjectByID(p.projectID)
		if err != nil {
			log.Errorf("failed to get project %d: %v", p.projectID, err)
			p.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if project == nil {
			p.CustomAbort(http.StatusNotFound, fmt.Sprintf("project does not exist, id: %v", p.projectID))
		}
		p.projectName = project.Name
	}
}

// Post ...
func (p *ProjectAPI) Post() {
	p.userID = p.ValidateUser()

	var req projectReq
	p.DecodeJSONReq(&req)
	public := req.Public
	err := validateProjectReq(req)
	if err != nil {
		log.Errorf("Invalid project request, error: %v", err)
		p.RenderError(http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
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
		dup, _ := regexp.MatchString(dupProjectPattern, err.Error())
		if dup {
			p.RenderError(http.StatusConflict, "")
		} else {
			p.RenderError(http.StatusInternalServerError, "Failed to add project")
		}
		return
	}
	p.Redirect(http.StatusCreated, strconv.FormatInt(projectID, 10))
}

// Head ...
func (p *ProjectAPI) Head() {
	projectName := p.GetString("project_name")
	if len(projectName) == 0 {
		p.CustomAbort(http.StatusBadRequest, "project_name is needed")
	}

	project, err := dao.GetProjectByName(projectName)
	if err != nil {
		log.Errorf("error occurred in GetProjectByName: %v", err)
		p.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	// only public project can be Headed by user without login
	if project != nil && project.Public == 1 {
		return
	}

	userID := p.ValidateUser()
	if project == nil {
		p.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	if !checkProjectPermission(userID, project.ProjectID) {
		p.CustomAbort(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
}

// Get ...
func (p *ProjectAPI) Get() {
	project, err := dao.GetProjectByID(p.projectID)
	if err != nil {
		log.Errorf("failed to get project %d: %v", p.projectID, err)
		p.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if project.Public == 0 {
		userID := p.ValidateUser()
		if !checkProjectPermission(userID, p.projectID) {
			p.CustomAbort(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		}
	}

	p.Data["json"] = project
	p.ServeJSON()
}

// Delete ...
func (p *ProjectAPI) Delete() {
	if p.projectID == 0 {
		p.CustomAbort(http.StatusBadRequest, "project ID is required")
	}

	userID := p.ValidateUser()

	if !hasProjectAdminRole(userID, p.projectID) {
		p.CustomAbort(http.StatusForbidden, "")
	}

	contains, err := projectContainsRepo(p.projectName)
	if err != nil {
		log.Errorf("failed to check whether project %s contains any repository: %v", p.projectName, err)
		p.CustomAbort(http.StatusInternalServerError, "")
	}
	if contains {
		p.CustomAbort(http.StatusPreconditionFailed, "project contains repositores, can not be deleted")
	}

	contains, err = projectContainsPolicy(p.projectID)
	if err != nil {
		log.Errorf("failed to check whether project %s contains any policy: %v", p.projectName, err)
		p.CustomAbort(http.StatusInternalServerError, "")
	}
	if contains {
		p.CustomAbort(http.StatusPreconditionFailed, "project contains policies, can not be deleted")
	}

	if err = dao.DeleteProject(p.projectID); err != nil {
		log.Errorf("failed to delete project %d: %v", p.projectID, err)
		p.CustomAbort(http.StatusInternalServerError, "")
	}

	go func() {
		if err := dao.AddAccessLog(models.AccessLog{
			UserID:    userID,
			ProjectID: p.projectID,
			RepoName:  p.projectName,
			Operation: "delete",
		}); err != nil {
			log.Errorf("failed to add access log: %v", err)
		}
	}()
}

func projectContainsRepo(name string) (bool, error) {
	repositories, err := getReposByProject(name)
	if err != nil {
		return false, err
	}

	return len(repositories) > 0, nil
}

func projectContainsPolicy(id int64) (bool, error) {
	policies, err := dao.GetRepPolicyByProject(id)
	if err != nil {
		return false, err
	}

	return len(policies) > 0, nil
}

// List ...
func (p *ProjectAPI) List() {
	var total int64
	var public int
	var err error

	page, pageSize := p.getPaginationParams()

	var projectList []models.Project
	projectName := p.GetString("project_name")

	isPublic := p.GetString("is_public")
	if len(isPublic) > 0 {
		public, err = strconv.Atoi(isPublic)
		if err != nil {
			log.Errorf("Error parsing public property: %v, error: %v", isPublic, err)
			p.CustomAbort(http.StatusBadRequest, "invalid project Id")
		}
	}
	isAdmin := false
	if public == 1 {
		total, err = dao.GetTotalOfProjects(projectName, 1)
		if err != nil {
			log.Errorf("failed to get total of projects: %v", err)
			p.CustomAbort(http.StatusInternalServerError, "")
		}
		projectList, err = dao.GetProjects(projectName, 1, pageSize, pageSize*(page-1))
		if err != nil {
			log.Errorf("failed to get projects: %v", err)
			p.CustomAbort(http.StatusInternalServerError, "")
		}
	} else {
		//if the request is not for public projects, user must login or provide credential
		p.userID = p.ValidateUser()
		isAdmin, err = dao.IsAdminRole(p.userID)
		if err != nil {
			log.Errorf("Error occured in check admin, error: %v", err)
			p.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if isAdmin {
			total, err = dao.GetTotalOfProjects(projectName)
			if err != nil {
				log.Errorf("failed to get total of projects: %v", err)
				p.CustomAbort(http.StatusInternalServerError, "")
			}
			projectList, err = dao.GetProjects(projectName, pageSize, pageSize*(page-1))
			if err != nil {
				log.Errorf("failed to get projects: %v", err)
				p.CustomAbort(http.StatusInternalServerError, "")
			}
		} else {
			total, err = dao.GetTotalOfUserRelevantProjects(p.userID, projectName)
			if err != nil {
				log.Errorf("failed to get total of projects: %v", err)
				p.CustomAbort(http.StatusInternalServerError, "")
			}
			projectList, err = dao.GetUserRelevantProjects(p.userID, projectName, pageSize, pageSize*(page-1))
			if err != nil {
				log.Errorf("failed to get projects: %v", err)
				p.CustomAbort(http.StatusInternalServerError, "")
			}
		}
	}

	for i := 0; i < len(projectList); i++ {
		if public != 1 {
			if isAdmin {
				projectList[i].Role = models.PROJECTADMIN
			}
			if projectList[i].Role == models.PROJECTADMIN {
				projectList[i].Togglable = true
			}
		}

		repos, err := dao.GetRepositoryByProjectName(projectList[i].Name)
		if err != nil {
			log.Errorf("failed to get repositories of project %s: %v", projectList[i].Name, err)
			p.CustomAbort(http.StatusInternalServerError, "")
		}

		projectList[i].RepoCount = len(repos)
	}

	p.setPaginationHeader(total, page, pageSize)
	p.Data["json"] = projectList
	p.ServeJSON()
}

// ToggleProjectPublic ...
func (p *ProjectAPI) ToggleProjectPublic() {
	p.userID = p.ValidateUser()
	var req projectReq

	projectID, err := strconv.ParseInt(p.Ctx.Input.Param(":id"), 10, 64)
	if err != nil {
		log.Errorf("Error parsing project id: %d, error: %v", projectID, err)
		p.RenderError(http.StatusBadRequest, "invalid project id")
		return
	}

	p.DecodeJSONReq(&req)
	public := req.Public
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
	p.userID = p.ValidateUser()

	var query models.AccessLog
	p.DecodeJSONReq(&query)

	if !checkProjectPermission(p.userID, p.projectID) {
		log.Warningf("Current user, user id: %d does not have permission to read accesslog of project, id: %d", p.userID, p.projectID)
		p.RenderError(http.StatusForbidden, "")
		return
	}
	query.ProjectID = p.projectID
	query.BeginTime = time.Unix(query.BeginTimestamp, 0)
	query.EndTime = time.Unix(query.EndTimestamp, 0)

	page, pageSize := p.getPaginationParams()

	total, err := dao.GetTotalOfAccessLogs(query)
	if err != nil {
		log.Errorf("failed to get total of access log: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "")
	}

	logs, err := dao.GetAccessLogs(query, pageSize, pageSize*(page-1))
	if err != nil {
		log.Errorf("failed to get access log: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "")
	}

	p.setPaginationHeader(total, page, pageSize)

	p.Data["json"] = logs

	p.ServeJSON()
}

func isProjectAdmin(userID int, pid int64) bool {
	isSysAdmin, err := dao.IsAdminRole(userID)
	if err != nil {
		log.Errorf("Error occurred in IsAdminRole, returning false, error: %v", err)
		return false
	}

	if isSysAdmin {
		return true
	}

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
	if isIllegalLength(req.ProjectName, projectNameMinLen, projectNameMaxLen) {
		return fmt.Errorf("Project name is illegal in length. (greater than 4 or less than 30)")
	}
	validProjectName := regexp.MustCompile(`^[a-z0-9](?:-*[a-z0-9])*(?:[._][a-z0-9](?:-*[a-z0-9])*)*$`)
	legal := validProjectName.MatchString(pn)
	if !legal {
		return fmt.Errorf("Project name is not in lower case or contains illegal characters!")
	}
	return nil
}
