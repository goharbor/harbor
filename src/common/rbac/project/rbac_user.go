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

package project

import (
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/project/models"
)

type rbacUser struct {
	project      *models.Project
	username     string
	projectRoles []int
	policies     []*types.Policy
}

// GetUserName returns username of the visitor
func (pru *rbacUser) GetUserName() string {
	return pru.username
}

// GetPolicies returns policies of the visitor
func (pru *rbacUser) GetPolicies() []*types.Policy {
	policies := pru.policies

	// Visibility that comes with membership rather than with a role: anyone
	// holding at least one role in a project can see the project exists. From
	// #23804, where it stopped a custom role being unable to view the project
	// it is scoped to.
	if len(pru.projectRoles) > 0 {
		policies = append(policies, &types.Policy{
			Resource: NewNamespace(pru.project.ProjectID).Resource(rbac.ResourceSelf),
			Action:   rbac.ActionRead,
		})
	}

	if pru.project.IsPublic() {
		policies = append(policies, getPoliciesForPublicProject(pru.project.ProjectID)...)
	}

	return policies
}

// GetRoles returns roles of the visitor
func (pru *rbacUser) GetRoles() []types.RBACRole {
	roles := []types.RBACRole{}
	for _, roleID := range pru.projectRoles {
		roles = append(roles, &projectRBACRole{projectID: pru.project.ProjectID, roleID: int64(roleID)})
	}

	return roles
}
