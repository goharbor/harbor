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

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/promgr"
)

// SecurityContext implements security.Context interface based on database
type SecurityContext struct {
	robot     *models.Robot
	pm        promgr.ProjectManager
	policy    []*rbac.Policy
	evaluator rbac.Evaluator
	once      sync.Once
}

// NewSecurityContext ...
func NewSecurityContext(robot *models.Robot, pm promgr.ProjectManager, policy []*rbac.Policy) *SecurityContext {
	return &SecurityContext{
		robot:  robot,
		pm:     pm,
		policy: policy,
	}
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

// GetMyProjects no implementation
func (s *SecurityContext) GetMyProjects() ([]*models.Project, error) {
	return nil, nil
}

// GetProjectRoles no implementation
func (s *SecurityContext) GetProjectRoles(projectIDOrName interface{}) []int {
	return nil
}

// Can returns whether the robot can do action on resource
func (s *SecurityContext) Can(action rbac.Action, resource rbac.Resource) bool {
	s.once.Do(func() {
		s.evaluator = rbac.NewNamespaceEvaluator("project", func(ns rbac.Namespace) rbac.Evaluator {
			projectID := ns.Identity().(int64)
			proj, err := s.pm.Get(projectID)
			if err != nil {
				log.Errorf("failed to get project %d, error: %v", projectID, err)
				return nil
			}
			if proj == nil {
				return nil
			}

			robot := NewRobot(s.GetUsername(), rbac.NewProjectNamespace(projectID, proj.IsPublic()), s.policy)
			return rbac.NewUserEvaluator(robot)
		})
	})

	return s.evaluator != nil && s.evaluator.HasPermission(resource, action)
}
