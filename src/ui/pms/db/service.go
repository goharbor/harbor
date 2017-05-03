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
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// PMS implements pms.PMS interface based on database
type PMS struct{}

// IsPublic returns whether the project is public or not
func (p *PMS) IsPublic(projectIDOrName interface{}) bool {
	var project *models.Project
	var err error
	switch projectIDOrName.(type) {
	case string:
		name := projectIDOrName.(string)
		project, err = dao.GetProjectByName(name)
		if err != nil {
			log.Errorf("failed to get project %s: %v", name, err)
		}
	case int64:
		id := projectIDOrName.(int64)
		project, err = dao.GetProjectByID(id)
		if err != nil {
			log.Errorf("failed to get project %d: %v", id, err)
		}
	default:
		log.Errorf("unsupported type of %v, must be string or int64", projectIDOrName)
	}

	if project == nil {
		return false
	}

	return project.Public == 1
}

// GetRoles return a role list which contains the user's roles to the project
func (p *PMS) GetRoles(username string, projectIDOrName interface{}) []int {
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

	var projectID int64
	switch projectIDOrName.(type) {
	case string:
		name := projectIDOrName.(string)
		project, err := dao.GetProjectByName(name)
		if err != nil {
			log.Errorf("failed to get project %s: %v", name, err)
			return roles
		}

		if project == nil {
			return roles
		}
		projectID = project.ProjectID
	case int64:
		projectID = projectIDOrName.(int64)
	default:
		log.Errorf("unsupported type of %v, must be string or int64", projectIDOrName)
		return roles
	}

	roleList, err := dao.GetUserProjectRoles(user.UserID, projectID)
	if err != nil {
		log.Errorf("failed to get roles for user %d to project %d: %v",
			user.UserID, projectID, err)
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
