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
	"fmt"
	"net/http"
	"regexp"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	errutil "github.com/goharbor/harbor/src/common/utils/error"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"

	"errors"
	"strconv"
	"time"
)

type deletableResp struct {
	Deletable bool   `json:"deletable"`
	Message   string `json:"message"`
}

// ProjectAPI handles request to /api/projects/{} /api/projects/{}/logs
type ProjectAPI struct {
	BaseController
	project *models.Project
}

const projectNameMaxLen int = 255
const projectNameMinLen int = 2
const restrictedNameChars = `[a-z0-9]+(?:[._-][a-z0-9]+)*`

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
			p.SendBadRequestError(errors.New(text))
			return
		}

		project, err := p.ProjectMgr.Get(id)
		if err != nil {
			p.ParseAndHandleError(fmt.Sprintf("failed to get project %d", id), err)
			return
		}

		if project == nil {
			p.SendNotFoundError(fmt.Errorf("project %d not found", id))
			return
		}

		p.project = project
	}
}

func (p *ProjectAPI) requireAccess(action rbac.Action, subresource ...rbac.Resource) bool {
	if len(subresource) == 0 {
		subresource = append(subresource, rbac.ResourceSelf)
	}
	resource := rbac.NewProjectNamespace(p.project.ProjectID).Resource(subresource...)

	if !p.SecurityCtx.Can(action, resource) {
		if !p.SecurityCtx.IsAuthenticated() {
			p.SendUnAuthorizedError(errors.New("Unauthorized"))

		} else {
			p.SendForbiddenError(errors.New(p.SecurityCtx.GetUsername()))
		}

		return false
	}

	return true
}

// Post ...
func (p *ProjectAPI) Post() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}
	var onlyAdmin bool
	var err error
	if config.WithAdmiral() {
		onlyAdmin = true
	} else {
		onlyAdmin, err = config.OnlyAdminCreateProject()
		if err != nil {
			log.Errorf("failed to determine whether only admin can create projects: %v", err)
			p.SendInternalServerError(fmt.Errorf("failed to determine whether only admin can create projects: %v", err))
			return
		}
	}

	if onlyAdmin && !p.SecurityCtx.IsSysAdmin() {
		log.Errorf("Only sys admin can create project")
		p.SendForbiddenError(errors.New("Only system admin can create project"))
		return
	}
	var pro *models.ProjectRequest
	if err := p.DecodeJSONReq(&pro); err != nil {
		p.SendBadRequestError(err)
		return
	}
	err = validateProjectReq(pro)
	if err != nil {
		log.Errorf("Invalid project request, error: %v", err)
		p.SendBadRequestError(fmt.Errorf("invalid request: %v", err))
		return
	}

	exist, err := p.ProjectMgr.Exists(pro.Name)
	if err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %s",
			pro.Name), err)
		return
	}
	if exist {
		p.SendConflictError(errors.New("conflict project"))
		return
	}

	if pro.Metadata == nil {
		pro.Metadata = map[string]string{}
	}
	// accept the "public" property to make replication work well with old versions(<=1.2.0)
	if pro.Public != nil && len(pro.Metadata[models.ProMetaPublic]) == 0 {
		pro.Metadata[models.ProMetaPublic] = strconv.FormatBool(*pro.Public == 1)
	}

	// populate public metadata as false if it isn't set
	if _, ok := pro.Metadata[models.ProMetaPublic]; !ok {
		pro.Metadata[models.ProMetaPublic] = strconv.FormatBool(false)
	}

	projectID, err := p.ProjectMgr.Create(&models.Project{
		Name:      pro.Name,
		OwnerName: p.SecurityCtx.GetUsername(),
		Metadata:  pro.Metadata,
	})
	if err != nil {
		if err == errutil.ErrDupProject {
			log.Debugf("conflict %s", pro.Name)
			p.SendConflictError(fmt.Errorf("conflict %s", pro.Name))
		} else {
			p.ParseAndHandleError("failed to add project", err)
		}
		return
	}

	go func() {
		if err = dao.AddAccessLog(
			models.AccessLog{
				Username:  p.SecurityCtx.GetUsername(),
				ProjectID: projectID,
				RepoName:  pro.Name + "/",
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
		p.SendBadRequestError(errors.New("project_name is needed"))
		return
	}

	project, err := p.ProjectMgr.Get(name)
	if err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to get project %s", name), err)
		return
	}

	if project == nil {
		p.SendNotFoundError(fmt.Errorf("project %s not found", name))
		return
	}
}

// Get ...
func (p *ProjectAPI) Get() {
	if !p.requireAccess(rbac.ActionRead) {
		return
	}

	p.populateProperties(p.project)

	p.Data["json"] = p.project
	p.ServeJSON()
}

// Delete ...
func (p *ProjectAPI) Delete() {
	if !p.requireAccess(rbac.ActionDelete) {
		return
	}

	result, err := p.deletable(p.project.ProjectID)
	if err != nil {
		p.SendInternalServerError(fmt.Errorf(
			"failed to check the deletable of project %d: %v", p.project.ProjectID, err))
		return
	}
	if !result.Deletable {
		p.SendPreconditionFailedError(errors.New(result.Message))
		return
	}

	if err = p.ProjectMgr.Delete(p.project.ProjectID); err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to delete project %d", p.project.ProjectID), err)
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

// Deletable ...
func (p *ProjectAPI) Deletable() {
	if !p.requireAccess(rbac.ActionDelete) {
		return
	}

	result, err := p.deletable(p.project.ProjectID)
	if err != nil {
		p.SendInternalServerError(fmt.Errorf(
			"failed to check the deletable of project %d: %v", p.project.ProjectID, err))
		return
	}

	p.Data["json"] = result
	p.ServeJSON()
}

func (p *ProjectAPI) deletable(projectID int64) (*deletableResp, error) {
	count, err := dao.GetTotalOfRepositories(&models.RepositoryQuery{
		ProjectIDs: []int64{projectID},
	})
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return &deletableResp{
			Deletable: false,
			Message:   "the project contains repositories, can not be deleted",
		}, nil
	}

	policies, err := dao.GetRepPolicyByProject(projectID)
	if err != nil {
		return nil, err
	}

	if len(policies) > 0 {
		return &deletableResp{
			Deletable: false,
			Message:   "the project contains replication rules, can not be deleted",
		}, nil
	}

	// Check helm charts number
	if config.WithChartMuseum() {
		charts, err := chartController.ListCharts(p.project.Name)
		if err != nil {
			return nil, err
		}

		if len(charts) > 0 {
			return &deletableResp{
				Deletable: false,
				Message:   "the project contains helm charts, can not be deleted",
			}, nil
		}
	}

	return &deletableResp{
		Deletable: true,
	}, nil
}

// List ...
func (p *ProjectAPI) List() {
	// query strings
	page, size, err := p.GetPaginationParams()
	if err != nil {
		p.SendBadRequestError(err)
		return
	}
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
			p.SendBadRequestError(fmt.Errorf("invalid public: %s", public))
			return
		}
		query.Public = &pub
	}

	// standalone, filter projects according to the privilleges of the user first
	if !config.WithAdmiral() {
		var projects []*models.Project
		if !p.SecurityCtx.IsAuthenticated() {
			// not login, only get public projects
			pros, err := p.ProjectMgr.GetPublic()
			if err != nil {
				p.SendInternalServerError(fmt.Errorf("failed to get public projects: %v", err))
				return
			}
			projects = []*models.Project{}
			projects = append(projects, pros...)
		} else {
			if !(p.SecurityCtx.IsSysAdmin() || p.SecurityCtx.IsSolutionUser()) {
				projects = []*models.Project{}
				// login, but not system admin or solution user, get public projects and
				// projects that the user is member of
				pros, err := p.ProjectMgr.GetPublic()
				if err != nil {
					p.SendInternalServerError(fmt.Errorf("failed to get public projects: %v", err))
					return
				}
				projects = append(projects, pros...)
				mps, err := p.SecurityCtx.GetMyProjects()
				if err != nil {
					p.SendInternalServerError(fmt.Errorf("failed to list projects: %v", err))
					return
				}
				projects = append(projects, mps...)
			}
		}
		// Query projects by user group

		if projects != nil {
			projectIDs := []int64{}
			for _, project := range projects {
				projectIDs = append(projectIDs, project.ProjectID)
			}
			query.ProjectIDs = projectIDs
		}
	}

	result, err := p.ProjectMgr.List(query)
	if err != nil {
		p.ParseAndHandleError("failed to list projects", err)
		return
	}

	for _, project := range result.Projects {
		p.populateProperties(project)
	}

	p.SetPaginationHeader(result.Total, page, size)
	p.Data["json"] = result.Projects
	p.ServeJSON()
}

func (p *ProjectAPI) populateProperties(project *models.Project) {
	if p.SecurityCtx.IsAuthenticated() {
		roles := p.SecurityCtx.GetProjectRoles(project.ProjectID)
		if len(roles) != 0 {
			project.Role = roles[0]
		}

		if project.Role == common.RoleProjectAdmin ||
			p.SecurityCtx.IsSysAdmin() {
			project.Togglable = true
		}
	}

	total, err := dao.GetTotalOfRepositories(&models.RepositoryQuery{
		ProjectIDs: []int64{project.ProjectID},
	})
	if err != nil {
		log.Errorf("failed to get total of repositories of project %d: %v", project.ProjectID, err)
		p.SendInternalServerError(errors.New(""))
		return
	}

	project.RepoCount = total

	// Populate chart count property
	if config.WithChartMuseum() {
		count, err := chartController.GetCountOfCharts([]string{project.Name})
		if err != nil {
			log.Errorf("Failed to get total of charts under project %s: %v", project.Name, err)
			p.SendInternalServerError(errors.New(""))
			return
		}

		project.ChartCount = count
	}
}

// Put ...
func (p *ProjectAPI) Put() {
	if !p.requireAccess(rbac.ActionUpdate) {
		return
	}

	var req *models.ProjectRequest
	if err := p.DecodeJSONReq(&req); err != nil {
		p.SendBadRequestError(err)
		return
	}

	if err := p.ProjectMgr.Update(p.project.ProjectID,
		&models.Project{
			Metadata: req.Metadata,
		}); err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to update project %d",
			p.project.ProjectID), err)
		return
	}
}

// Logs ...
func (p *ProjectAPI) Logs() {
	if !p.requireAccess(rbac.ActionList, rbac.ResourceLog) {
		return
	}

	page, size, err := p.GetPaginationParams()
	if err != nil {
		p.SendBadRequestError(err)
		return
	}
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
			p.SendBadRequestError(fmt.Errorf("invalid begin_timestamp: %s", timestamp))
			return
		}
		query.BeginTime = t
	}

	timestamp = p.GetString("end_timestamp")
	if len(timestamp) > 0 {
		t, err := utils.ParseTimeStamp(timestamp)
		if err != nil {
			p.SendBadRequestError(fmt.Errorf("invalid end_timestamp: %s", timestamp))
			return
		}
		query.EndTime = t
	}

	total, err := dao.GetTotalOfAccessLogs(query)
	if err != nil {
		p.SendInternalServerError(fmt.Errorf(
			"failed to get total of access log: %v", err))
		return
	}

	logs, err := dao.GetAccessLogs(query)
	if err != nil {
		p.SendInternalServerError(fmt.Errorf(
			"failed to get access log: %v", err))
		return
	}

	p.SetPaginationHeader(total, page, size)
	p.Data["json"] = logs
	p.ServeJSON()
}

// TODO move this to pa ckage models
func validateProjectReq(req *models.ProjectRequest) error {
	pn := req.Name
	if utils.IsIllegalLength(req.Name, projectNameMinLen, projectNameMaxLen) {
		return fmt.Errorf("Project name is illegal in length. (greater than %d or less than %d)", projectNameMaxLen, projectNameMinLen)
	}
	validProjectName := regexp.MustCompile(`^` + restrictedNameChars + `$`)
	legal := validProjectName.MatchString(pn)
	if !legal {
		return fmt.Errorf("project name is not in lower case or contains illegal characters")
	}

	metas, err := validateProjectMetadata(req.Metadata)
	if err != nil {
		return err
	}

	req.Metadata = metas
	return nil
}
