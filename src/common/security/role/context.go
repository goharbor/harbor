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

//TODO I think this is not used ... only the local context ... it is defined statically in rbac_role.go

package role

import (
	"context"
	"sync"

	"github.com/goharbor/harbor/src/common/rbac"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/log"

	"github.com/goharbor/harbor/src/controller/role"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/project/models"
)

// SecurityContext implements security.Context interface based on database
type SecurityContext struct {
	role      *role.Role
	ctl       project.Controller
	evaluator evaluator.Evaluator
	once      sync.Once
}

// NewSecurityContext ...
func NewSecurityContext(r *role.Role) *SecurityContext {
	return &SecurityContext{
		ctl:  project.Ctl,
		role: r,
	}
}

// Name returns the name of the security context
func (s *SecurityContext) Name() string {
	return "role"
}

// IsAuthenticated returns true if the user has been authenticated
func (s *SecurityContext) IsAuthenticated() bool {
	return s.role != nil
}

// GetUsername returns the username of the authenticated user
// It returns null if the user has not been authenticated
func (s *SecurityContext) GetUsername() string {
	if !s.IsAuthenticated() {
		return ""
	}
	return s.role.Name
}

// User get the current user
func (s *SecurityContext) User() *role.Role {
	return s.role
}

// IsSysAdmin role cannot be a system admin
func (s *SecurityContext) IsSysAdmin() bool {
	return false
}

// IsSolutionUser role cannot be a system admin
func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

// TODO MGS DELETE this --- Can returns whether the role can do action on resource
func (s *SecurityContext) Can(ctx context.Context, action types.Action, resource types.Resource) bool {
	log.Debug("*** roles Can do everything")
	return true

}

func filterRolePolicies(p *models.Project, policies []*types.Policy) []*types.Policy {
	namespace := rbac_project.NewNamespace(p.ProjectID)

	var results []*types.Policy
	for _, policy := range policies {
		if types.ResourceAllowedInNamespace(policy.Resource, namespace) {
			results = append(results, policy)
			// give the PUSH action a pull access
			if policy.Action == rbac.ActionPush {
				results = append(results, &types.Policy{Resource: policy.Resource, Action: rbac.ActionPull})
			}
		}
	}
	return results
}

// getPolicyResource to determine permissions for the project resource, the path should be /project instead of /project/project.
/*
func getPolicyResource(perm *role.Permission, pol *types.Policy) string {
	if strings.HasPrefix(perm.Scope, role.SCOPEPROJECT) && pol.Resource == rbac.ResourceProject {
		return perm.Scope
	}
	return fmt.Sprintf("%s/%s", perm.Scope, pol.Resource)
}
*/
