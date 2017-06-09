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
	"regexp"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"

	"strconv"
	"time"
)

// ProjectAPI handles request to /api/projects/{} /api/projects/{}/logs
type ProjectAPI struct {
	BaseController
	project *models.Project
}

type projectReq struct {
	ProjectName string `json:"project_name"`
	Public      int    `json:"public"`
}

const projectNameMaxLen int = 30
const projectNameMinLen int = 2
const restrictedNameChars = `[a-z0-9]+(?:[._-][a-z0-9]+)*`
const dupProjectPattern = `Duplicate entry '\w+' for key 'name'`

// Prepare validates the URL and the user
func (p *ProjectAPI) Prepare() {
	p.BaseController.Prepare()
	if len(p.GetStringFromPath(":id")) != 0 {
		id, err := p.GetInt64FromPath(":id")
		if err != nil || id <= 0 {
			text := "invalid project ID: "
			if err != nil {
				text += err.Error()
			} else {
				text += fmt.Sprintf("%d", id)
			}
			p.HandleBadRequest(text)
			return
		}

		project, err := p.ProjectMgr.Get(id)
		if err != nil {
			p.HandleInternalServerError(fmt.Sprintf("failed to get project %d: %v",
				id, err))
			return
		}

		if project == nil {
			p.HandleNotFound(fmt.Sprintf("project %d not found", id))
			return
		}

		p.project = project
	}
}

// Post ...
func (p *ProjectAPI) Post() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.HandleUnauthorized()
		return
	}

	onlyAdmin, err := config.OnlyAdminCreateProject()
	if err != nil {
		log.Errorf("failed to determine whether only admin can create projects: %v", err)
		p.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	if onlyAdmin && !p.SecurityCtx.IsSysAdmin() {
		log.Errorf("Only sys admin can create project")
		p.RenderError(http.StatusForbidden, "Only system admin can create project")
		return
	}
	var pro projectReq
	p.DecodeJSONReq(&pro)
	err = validateProjectReq(pro)
	if err != nil {
		log.Errorf("Invalid project request, error: %v", err)
		p.RenderError(http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	exist, err := p.ProjectMgr.Exist(pro.ProjectName)
	if err != nil {
		p.HandleInternalServerError(fmt.Sprintf("failed to check the existence of project %s: %v",
			pro.ProjectName, err))
		return
	}
	if exist {
		p.RenderError(http.StatusConflict, "")
		return
	}

	projectID, err := p.ProjectMgr.Create(&models.Project{
		Name:      pro.ProjectName,
		Public:    pro.Public,
		OwnerName: p.SecurityCtx.GetUsername(),
	})
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

	go func() {
		if err = dao.AddAccessLog(
			models.AccessLog{
				Username:  p.SecurityCtx.GetUsername(),
				ProjectID: projectID,
				RepoName:  pro.ProjectName + "/",
				RepoTag:   "N/A",
				Operation: "create",
				OpTime:    time.Now(),
			}); err != nil {
			log.Errorf("failed to add access log: %v", err)
		}
	}()

	p.Redirect(http.StatusCreated, strconv.FormatInt(projectID, 10))
}

// Head ...
func (p *ProjectAPI) Head() {
	name := p.GetString("project_name")
	if len(name) == 0 {
		p.HandleBadRequest("project_name is needed")
		return
	}

	project, err := p.ProjectMgr.Get(name)
	if err != nil {
		p.HandleInternalServerError(fmt.Sprintf("failed to get project %s: %v",
			name, err))
		return
	}

	if project == nil {
		p.HandleNotFound(fmt.Sprintf("project %s not found", name))
		return
	}
}

// Get ...
func (p *ProjectAPI) Get() {
	if p.project.Public == 0 {
		if !p.SecurityCtx.IsAuthenticated() {
			p.HandleUnauthorized()
			return
		}

		if !p.SecurityCtx.HasReadPerm(p.project.ProjectID) {
			p.HandleForbidden(p.SecurityCtx.GetUsername())
			return
		}
	}

	p.Data["json"] = p.project
	p.ServeJSON()
}

// Delete ...
func (p *ProjectAPI) Delete() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.HandleUnauthorized()
		return
	}

	if !p.SecurityCtx.HasAllPerm(p.project.ProjectID) {
		p.HandleForbidden(p.SecurityCtx.GetUsername())
		return
	}

	contains, err := projectContainsRepo(p.project.Name)
	if err != nil {
		log.Errorf("failed to check whether project %s contains any repository: %v", p.project.Name, err)
		p.CustomAbort(http.StatusInternalServerError, "")
	}
	if contains {
		p.CustomAbort(http.StatusPreconditionFailed, "project contains repositores, can not be deleted")
	}

	contains, err = projectContainsPolicy(p.project.ProjectID)
	if err != nil {
		log.Errorf("failed to check whether project %s contains any policy: %v", p.project.Name, err)
		p.CustomAbort(http.StatusInternalServerError, "")
	}
	if contains {
		p.CustomAbort(http.StatusPreconditionFailed, "project contains policies, can not be deleted")
	}

	if err = p.ProjectMgr.Delete(p.project.ProjectID); err != nil {
		p.HandleInternalServerError(
			fmt.Sprintf("failed to delete project %d: %v", p.project.ProjectID, err))
		return
	}

	go func() {
		if err := dao.AddAccessLog(models.AccessLog{
			Username:  p.SecurityCtx.GetUsername(),
			ProjectID: p.project.ProjectID,
			RepoName:  p.project.Name + "/",
			RepoTag:   "N/A",
			Operation: "delete",
			OpTime:    time.Now(),
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
	// query strings
	page, size := p.GetPaginationParams()
	query := &models.ProjectQueryParam{
		Name:  p.GetString("name"),
		Owner: p.GetString("owner"),
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}

	public := p.GetString("public")
	if len(public) > 0 {
		pub, err := strconv.ParseBool(public)
		if err != nil {
			p.HandleBadRequest(fmt.Sprintf("invalid public: %s", public))
			return
		}
		query.Public = &pub
	}

	member := p.GetString("member")
	if len(member) > 0 {
		query.Member = &models.Member{
			Name: member,
		}

		role := p.GetString("role")
		if len(role) > 0 {
			r, err := strconv.Atoi(role)
			if err != nil {
				if err != nil {
					p.HandleBadRequest(fmt.Sprintf("invalid role: %s", role))
					return
				}
			}
			query.Member.Role = r
		}
	}

	// base project collection from which filter is done
	base := &models.BaseProjectCollection{}
	if !p.SecurityCtx.IsAuthenticated() {
		if query.Member != nil && len(query.Member.Name) > 0 {
			// must login if query member
			p.HandleUnauthorized()
			return
		}
		base.Public = true
	} else {
		if !p.SecurityCtx.IsSysAdmin() {
			base.Member = p.SecurityCtx.GetUsername()
			if query.Member != nil && len(query.Member.Name) > 0 {
				base.Public = false
			} else {
				base.Public = true
			}
		}
	}

	total, err := p.ProjectMgr.GetTotal(query, base)
	if err != nil {
		p.HandleInternalServerError(fmt.Sprintf("failed to get total of projects: %v", err))
		return
	}

	projects, err := p.ProjectMgr.GetAll(query, base)
	if err != nil {
		p.HandleInternalServerError(fmt.Sprintf("failed to get projects: %v", err))
		return
	}

	for _, project := range projects {
		if p.SecurityCtx.IsAuthenticated() {
			roles, err := p.ProjectMgr.GetRoles(p.SecurityCtx.GetUsername(), project.ProjectID)
			if err != nil {
				p.HandleInternalServerError(fmt.Sprintf("failed to get roles of user %s to project %d: %v",
					p.SecurityCtx.GetUsername(), project.ProjectID, err))
				return
			}

			if len(roles) != 0 {
				project.Role = roles[0]
			}

			if project.Role == common.RoleProjectAdmin ||
				p.SecurityCtx.IsSysAdmin() {
				project.Togglable = true
			}
		}

		repos, err := dao.GetRepositoryByProjectName(project.Name)
		if err != nil {
			log.Errorf("failed to get repositories of project %s: %v", project.Name, err)
			p.CustomAbort(http.StatusInternalServerError, "")
		}

		project.RepoCount = len(repos)
	}

	p.SetPaginationHeader(total, page, size)
	p.Data["json"] = projects
	p.ServeJSON()
}

// ToggleProjectPublic ...
func (p *ProjectAPI) ToggleProjectPublic() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.HandleUnauthorized()
		return
	}

	if !p.SecurityCtx.HasAllPerm(p.project.ProjectID) {
		p.HandleForbidden(p.SecurityCtx.GetUsername())
		return
	}

	var req projectReq
	p.DecodeJSONReq(&req)
	if req.Public != 0 && req.Public != 1 {
		p.HandleBadRequest("public should be 0 or 1")
		return
	}

	if err := p.ProjectMgr.Update(p.project.ProjectID,
		&models.Project{
			Public: req.Public,
		}); err != nil {
		p.HandleInternalServerError(fmt.Sprintf("failed to update project %d: %v",
			p.project.ProjectID, err))
		return
	}
}

// Logs ...
func (p *ProjectAPI) Logs() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.HandleUnauthorized()
		return
	}

	if !p.SecurityCtx.HasReadPerm(p.project.ProjectID) {
		p.HandleForbidden(p.SecurityCtx.GetUsername())
		return
	}

	page, size := p.GetPaginationParams()
	query := &models.LogQueryParam{
		ProjectIDs: []int64{p.project.ProjectID},
		Username:   p.GetString("username"),
		Repository: p.GetString("repository"),
		Tag:        p.GetString("tag"),
		Operations: p.GetStrings("operation"),
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}

	timestamp := p.GetString("begin_timestamp")
	if len(timestamp) > 0 {
		t, err := utils.ParseTimeStamp(timestamp)
		if err != nil {
			p.HandleBadRequest(fmt.Sprintf("invalid begin_timestamp: %s", timestamp))
			return
		}
		query.BeginTime = t
	}

	timestamp = p.GetString("end_timestamp")
	if len(timestamp) > 0 {
		t, err := utils.ParseTimeStamp(timestamp)
		if err != nil {
			p.HandleBadRequest(fmt.Sprintf("invalid end_timestamp: %s", timestamp))
			return
		}
		query.EndTime = t
	}

	total, err := dao.GetTotalOfAccessLogs(query)
	if err != nil {
		p.HandleInternalServerError(fmt.Sprintf(
			"failed to get total of access log: %v", err))
		return
	}

	logs, err := dao.GetAccessLogs(query)
	if err != nil {
		p.HandleInternalServerError(fmt.Sprintf(
			"failed to get access log: %v", err))
		return
	}

	p.SetPaginationHeader(total, page, size)
	p.Data["json"] = logs
	p.ServeJSON()
}

// TODO move this to package models
func validateProjectReq(req projectReq) error {
	pn := req.ProjectName
	if isIllegalLength(req.ProjectName, projectNameMinLen, projectNameMaxLen) {
		return fmt.Errorf("Project name is illegal in length. (greater than 2 or less than 30)")
	}
	validProjectName := regexp.MustCompile(`^` + restrictedNameChars + `$`)
	legal := validProjectName.MatchString(pn)
	if !legal {
		return fmt.Errorf("project name is not in lower case or contains illegal characters")
	}
	return nil
}
