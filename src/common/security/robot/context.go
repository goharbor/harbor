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

package robot

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/common/rbac"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/rbac/system"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/project/models"
)

// SecurityContext implements security.Context interface based on database
type SecurityContext struct {
	robot     *robot.Robot
	ctl       project.Controller
	evaluator evaluator.Evaluator
	once      sync.Once
}

// NewSecurityContext ...
func NewSecurityContext(r *robot.Robot) *SecurityContext {
	return &SecurityContext{
		ctl:   project.Ctl,
		robot: r,
	}
}

// Name returns the name of the security context
func (s *SecurityContext) Name() string {
	return "robot"
}

// IsAuthenticated returns true if the user has been authenticated
func (s *SecurityContext) IsAuthenticated() bool {
	return s.robot != nil
}

// GetUsername returns the username of the authenticated user
// It returns null if the user has not been authenticated
func (s *SecurityContext) GetUsername() string {
	if !s.IsAuthenticated() {
		return ""
	}
	return s.robot.Name
}

// User get the current user
func (s *SecurityContext) User() *robot.Robot {
	return s.robot
}

// IsSysAdmin robot cannot be a system admin
func (s *SecurityContext) IsSysAdmin() bool {
	return false
}

// IsSolutionUser robot cannot be a system admin
func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

// Can returns whether the robot can do action on resource
func (s *SecurityContext) Can(ctx context.Context, action types.Action, resource types.Resource) bool {
	if s.robot == nil {
		return false
	}

	s.once.Do(func() {
		var accesses []*types.Policy
		for _, p := range s.robot.Permissions {
			for _, a := range p.Access {
				policyResource := getPolicyResource(p, a)
				accesses = append(accesses, &types.Policy{
					Action:   a.Action,
					Effect:   a.Effect,
					Resource: types.Resource(policyResource),
				})
			}
		}

		if s.robot.Level == robot.LEVELSYSTEM {
			var proPolicies []*types.Policy
			var sysPolicies []*types.Policy
			var evaluators evaluator.Evaluators
			for _, p := range accesses {
				if strings.HasPrefix(p.Resource.String(), robot.SCOPESYSTEM) {
					sysPolicies = append(sysPolicies, p)
				} else if strings.HasPrefix(p.Resource.String(), robot.SCOPEPROJECT) {
					proPolicies = append(proPolicies, p)
				}
			}
			if len(sysPolicies) != 0 {
				evaluators = evaluators.Add(system.NewEvaluator(s.GetUsername(), sysPolicies))
			}
			if len(proPolicies) != 0 {
				evaluators = evaluators.Add(rbac_project.NewEvaluator(s.ctl, rbac_project.NewBuilderForPolicies(s.GetUsername(), proPolicies, filterRobotPolicies)))
			}
			s.evaluator = evaluators
		} else {
			s.evaluator = rbac_project.NewEvaluator(s.ctl, rbac_project.NewBuilderForPolicies(s.GetUsername(), accesses, filterRobotPolicies))
		}
	})

	return s.evaluator != nil && s.evaluator.HasPermission(ctx, resource, action)
}

func filterRobotPolicies(p *models.Project, policies []*types.Policy) []*types.Policy {
	if p == nil {
		return policies
	}
	namespace := rbac_project.NewNamespace(p.ProjectID)

	var results []*types.Policy
	for _, policy := range policies {
		// Check if this is a wildcard permission by checking if resource starts with "/project/*"
		// Wildcard permissions from namespace="*" create resources like "/project/*/repository"
		if strings.HasPrefix(policy.Resource.String(), robot.SCOPEALLPROJECT) {
			// For wildcard permissions, allow access to any project
			results = append(results, policy)
			// give the PUSH action a pull access
			if policy.Action == rbac.ActionPush {
				results = append(results, &types.Policy{Resource: policy.Resource, Action: rbac.ActionPull, Effect: policy.Effect})
			}
		} else if types.ResourceAllowedInNamespace(policy.Resource, namespace) {
			// For specific project permissions, only allow if project matches
			results = append(results, policy)
			// give the PUSH action a pull access
			if policy.Action == rbac.ActionPush {
				results = append(results, &types.Policy{Resource: policy.Resource, Action: rbac.ActionPull, Effect: policy.Effect})
			}
		}
	}
	return results
}

// getPolicyResource to determine permissions for the project resource, the path should be /project instead of /project/project.
func getPolicyResource(perm *robot.Permission, pol *types.Policy) string {
	if strings.HasPrefix(perm.Scope, robot.SCOPEPROJECT) && pol.Resource == rbac.ResourceProject {
		return perm.Scope
	}
	return fmt.Sprintf("%s/%s", perm.Scope, pol.Resource)
}
