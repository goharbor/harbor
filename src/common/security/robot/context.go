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
	"sync"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

// SecurityContext implements security.Context interface based on database
type SecurityContext struct {
	robot     *model.Robot
	pm        promgr.ProjectManager
	policy    []*types.Policy
	evaluator evaluator.Evaluator
	once      sync.Once
}

// NewSecurityContext ...
func NewSecurityContext(robot *model.Robot, pm promgr.ProjectManager, policy []*types.Policy) *SecurityContext {
	return &SecurityContext{
		robot:  robot,
		pm:     pm,
		policy: policy,
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

// IsSysAdmin robot cannot be a system admin
func (s *SecurityContext) IsSysAdmin() bool {
	return false
}

// IsSolutionUser robot cannot be a system admin
func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

// Can returns whether the robot can do action on resource
func (s *SecurityContext) Can(action types.Action, resource types.Resource) bool {
	s.once.Do(func() {
		robotFactory := func(ns types.Namespace) types.RBACUser {
			return NewRobot(s.GetUsername(), ns, s.policy)
		}

		s.evaluator = rbac.NewProjectRobotEvaluator(s, s.pm, robotFactory)
	})

	return s.evaluator != nil && s.evaluator.HasPermission(resource, action)
}
