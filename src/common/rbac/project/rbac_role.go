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
	"github.com/goharbor/harbor/src/controller/role"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// projectRBACRole implement the RBACRole interface
type projectRBACRole struct {
	projectID int64
	role      *role.Role
}

func (role *projectRBACRole) GetRoleName() string {
	log.Debug("*** get roleName", role.role.Name)

	return role.role.Name
}

// GetPolicies returns policies for the visitor role
func (role *projectRBACRole) GetPolicies() []*types.Policy {

	log.Debug("*** get policies for project:%n, role:%n", role.projectID, role.role.Name)
	policies := []*types.Policy{}
	namespace := NewNamespace(role.projectID)

	for _, permission := range role.role.Permissions {
		for _, policy := range permission.Access {
			policies = append(policies, &types.Policy{
				Resource: namespace.Resource(policy.Resource),
				Action:   policy.Action,
				Effect:   policy.Effect,
			})
		}
	}
	return policies
}
