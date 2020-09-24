// Copyright Project Harbor Authors
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

package rbac

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

type projectRBACUser struct {
	project      *models.Project
	username     string
	projectRoles []int
}

// GetUserName returns username of the visitor
func (user *projectRBACUser) GetUserName() string {
	return user.username
}

// GetPolicies returns policies of the visitor
func (user *projectRBACUser) GetPolicies() []*types.Policy {
	if user.project.IsPublic() {
		return getPoliciesForPublicProject(user.project.ProjectID)
	}

	return nil
}

// GetRoles returns roles of the visitor
func (user *projectRBACUser) GetRoles() []types.RBACRole {
	roles := []types.RBACRole{}
	for _, roleID := range user.projectRoles {
		roles = append(roles, &projectRBACRole{projectID: user.project.ProjectID, roleID: roleID})
	}

	return roles
}

// NewProjectRBACUser returns RBACUser for the project
func NewProjectRBACUser(project *models.Project, username string, projectRoles ...int) types.RBACUser {
	return &projectRBACUser{
		project:      project,
		username:     username,
		projectRoles: projectRoles,
	}
}
