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
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/ram"
)

var (
	rolePoliciesMap = map[string][]*ram.Policy{
		"projectAdmin": {
			{Resource: ResourceImage, Action: ActionPushPull}, // compatible with security all perm of project
			{Resource: ResourceImage, Action: ActionPush},
			{Resource: ResourceImage, Action: ActionPull},
		},

		"developer": {
			{Resource: ResourceImage, Action: ActionPush},
			{Resource: ResourceImage, Action: ActionPull},
		},

		"guest": {
			{Resource: ResourceImage, Action: ActionPull},
		},
	}
)

// visitorRole implement the ram.Role interface
type visitorRole struct {
	namespace ram.Namespace
	roleID    int
}

// GetRoleName returns role name for the visitor role
func (role *visitorRole) GetRoleName() string {
	switch role.roleID {
	case common.RoleProjectAdmin:
		return "projectAdmin"
	case common.RoleDeveloper:
		return "developer"
	case common.RoleGuest:
		return "guest"
	default:
		return ""
	}
}

// GetPolicies returns policies for the visitor role
func (role *visitorRole) GetPolicies() []*ram.Policy {
	policies := []*ram.Policy{}

	roleName := role.GetRoleName()
	if roleName == "" {
		return policies
	}

	for _, policy := range rolePoliciesMap[roleName] {
		policies = append(policies, &ram.Policy{
			Resource: role.namespace.Resource(policy.Resource),
			Action:   policy.Action,
			Effect:   policy.Effect,
		})
	}

	return policies
}
