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
	"github.com/goharbor/harbor/src/common/ram"
)

// visitorContext the context interface for the project visitor
type visitorContext interface {
	IsAuthenticated() bool
	// GetUsername returns the username of user related to the context
	GetUsername() string
	// IsSysAdmin returns whether the user is system admin
	IsSysAdmin() bool
}

// visitor implement the ram.User interface for project visitor
type visitor struct {
	ctx          visitorContext
	namespace    ram.Namespace
	projectRoles []int
}

// GetUserName returns username of the visitor
func (v *visitor) GetUserName() string {
	// anonymous username for unauthenticated Visitor
	if !v.ctx.IsAuthenticated() {
		return "anonymous"
	}

	return v.ctx.GetUsername()
}

// GetPolicies returns policies of the visitor
func (v *visitor) GetPolicies() []*ram.Policy {
	if v.ctx.IsSysAdmin() {
		return policiesForSystemAdmin(v.namespace)
	}

	if v.namespace.IsPublic() {
		return policiesForPublicProject(v.namespace)
	}

	return nil
}

// GetRoles returns roles of the visitor
func (v *visitor) GetRoles() []ram.Role {
	if !v.ctx.IsAuthenticated() {
		return nil
	}

	roles := []ram.Role{}

	for _, roleID := range v.projectRoles {
		roles = append(roles, &visitorRole{roleID: roleID, namespace: v.namespace})
	}

	return roles
}

// NewUser returns ram.User interface for the project visitor
func NewUser(ctx visitorContext, namespace ram.Namespace, projectRoles ...int) ram.User {
	return &visitor{
		ctx:          ctx,
		namespace:    namespace,
		projectRoles: projectRoles,
	}
}
