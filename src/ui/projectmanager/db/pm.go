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

package db

import (
	"fmt"
	"time"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// ProjectManager implements pm.PM interface based on database
type ProjectManager struct{}

// Get ...
func (p *ProjectManager) Get(projectIDOrName interface{}) *models.Project {
	switch projectIDOrName.(type) {
	case string:
		name := projectIDOrName.(string)
		project, err := dao.GetProjectByName(name)
		if err != nil {
			log.Errorf("failed to get project %s: %v", name, err)
			return nil
		}
		return project
	case int64:
		id := projectIDOrName.(int64)
		project, err := dao.GetProjectByID(id)
		if err != nil {
			log.Errorf("failed to get project %d: %v", id, err)
			return nil
		}
		return project
	default:
		log.Errorf("unsupported type of %v, must be string or int64", projectIDOrName)
		return nil
	}
}

// Exist ...
func (p *ProjectManager) Exist(projectIDOrName interface{}) bool {
	return p.Get(projectIDOrName) != nil
}

// IsPublic returns whether the project is public or not
func (p *ProjectManager) IsPublic(projectIDOrName interface{}) bool {
	project := p.Get(projectIDOrName)
	if project == nil {
		return false
	}

	return project.Public == 1
}

// GetRoles return a role list which contains the user's roles to the project
func (p *ProjectManager) GetRoles(username string, projectIDOrName interface{}) []int {
	roles := []int{}

	user, err := dao.GetUser(models.User{
		Username: username,
	})
	if err != nil {
		log.Errorf("failed to get user %s: %v", username, err)
		return roles
	}
	if user == nil {
		return roles
	}

	project := p.Get(projectIDOrName)
	if project == nil {
		return roles
	}

	roleList, err := dao.GetUserProjectRoles(user.UserID, project.ProjectID)
	if err != nil {
		log.Errorf("failed to get roles for user %d to project %d: %v",
			user.UserID, project.ProjectID, err)
		return roles
	}

	for _, role := range roleList {
		switch role.RoleCode {
		case "MDRWS":
			roles = append(roles, common.RoleProjectAdmin)
		case "RWS":
			roles = append(roles, common.RoleDeveloper)
		case "RS":
			roles = append(roles, common.RoleGuest)
		}
	}

	return roles
}

// GetPublic returns all public projects
func (p *ProjectManager) GetPublic() []*models.Project {
	return filter("", "", "true", "", 0, 0, 0)
}

// GetByMember returns all projects which the user is a member of
func (p *ProjectManager) GetByMember(username string) []*models.Project {
	return filter("", "", "", username, 0, 0, 0)
}

// Create ...
func (p *ProjectManager) Create(project *models.Project) (int64, error) {
	if project == nil {
		return 0, fmt.Errorf("project is nil")
	}

	if len(project.Name) == 0 {
		return 0, fmt.Errorf("project name is nil")
	}

	if project.OwnerID == 0 {
		if len(project.OwnerName) == 0 {
			return 0, fmt.Errorf("owner ID and owner name are both nil")
		}

		user, err := dao.GetUser(models.User{
			Username: project.OwnerName,
		})
		if err != nil {
			return 0, err
		}
		if user == nil {
			return 0, fmt.Errorf("can not get owner whose name is %s", project.OwnerName)
		}
		project.OwnerID = user.UserID
	}

	t := time.Now()
	pro := &models.Project{
		Name:         project.Name,
		Public:       project.Public,
		OwnerID:      project.OwnerID,
		CreationTime: t,
		UpdateTime:   t,
	}

	return dao.AddProject(*pro)
}

// Delete ...
func (p *ProjectManager) Delete(projectIDOrName interface{}) error {
	id, ok := projectIDOrName.(int64)
	if !ok {
		project := p.Get(projectIDOrName)
		if project == nil {
			return fmt.Errorf(fmt.Sprintf("project %v not found", projectIDOrName))
		}
		id = project.ProjectID
	}

	return dao.DeleteProject(id)
}

// Update ...
func (p *ProjectManager) Update(projectIDOrName interface{},
	project *models.Project) error {
	id, ok := projectIDOrName.(int64)
	if !ok {
		pro := p.Get(projectIDOrName)
		if pro == nil {
			return fmt.Errorf(fmt.Sprintf("project %v not found", projectIDOrName))
		}
		id = pro.ProjectID
	}
	return dao.ToggleProjectPublicity(id, project.Public)
}

// GetAll ...
func (p *ProjectManager) GetAll(owner, name, public, member string,
	role int, page, size int64) ([]*models.Project, int64) {
	total, err := dao.GetTotalOfProjects(owner, name, public, member, role)
	if err != nil {
		log.Errorf("failed to get total of projects: %v", err)
		return []*models.Project{}, 0
	}

	return filter(owner, name, public, member, role, page, size), total
}

func filter(owner, name, public, member string,
	role int, page, size int64) []*models.Project {
	projects := []*models.Project{}

	list, err := dao.GetProjects(owner, name, public, member, role,
		page, size)
	if err != nil {
		log.Errorf("failed to get projects: %v", err)
		return projects
	}

	if len(list) != 0 {
		projects = append(projects, list...)
	}

	return projects
}
