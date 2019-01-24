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
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/core/promgr"
)

// SecurityContext implements security.Context interface based on database
type SecurityContext struct {
	robot  *models.Robot
	pm     promgr.ProjectManager
	policy []*rbac.Policy
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

// HasReadPerm returns whether the user has read permission to the project
func (s *SecurityContext) HasReadPerm(projectIDOrName interface{}) bool {
	isPublicProject, _ := s.pm.IsPublic(projectIDOrName)
	return s.Can(project.ActionPull, rbac.NewProjectNamespace(projectIDOrName, isPublicProject).Resource(project.ResourceImage))
}

// HasWritePerm returns whether the user has write permission to the project
func (s *SecurityContext) HasWritePerm(projectIDOrName interface{}) bool {
	isPublicProject, _ := s.pm.IsPublic(projectIDOrName)
	return s.Can(project.ActionPush, rbac.NewProjectNamespace(projectIDOrName, isPublicProject).Resource(project.ResourceImage))
}

// HasAllPerm returns whether the user has all permissions to the project
func (s *SecurityContext) HasAllPerm(projectIDOrName interface{}) bool {
	isPublicProject, _ := s.pm.IsPublic(projectIDOrName)
	return s.Can(project.ActionPushPull, rbac.NewProjectNamespace(projectIDOrName, isPublicProject).Resource(project.ResourceImage))
}

// GetMyProjects no implementation
func (s *SecurityContext) GetMyProjects() ([]*models.Project, error) {
	return nil, nil
}

// GetPolicies get access infor from the token and convert it to the rbac policy
func (s *SecurityContext) GetPolicies() []*rbac.Policy {
	return s.policy
}

// GetProjectRoles no implementation
func (s *SecurityContext) GetProjectRoles(projectIDOrName interface{}) []int {
	return nil
}

// Can returns whether the robot can do action on resource
func (s *SecurityContext) Can(action rbac.Action, resource rbac.Resource) bool {
	ns, err := resource.GetNamespace()
	if err == nil {
		switch ns.Kind() {
		case "project":
			projectIDOrName := ns.Identity()
			isPublicProject, _ := s.pm.IsPublic(projectIDOrName)
			projectNamespace := rbac.NewProjectNamespace(projectIDOrName, isPublicProject)
			robot := project.NewRobot(s, projectNamespace)
			return rbac.HasPermission(robot, resource, action)
		}
	}

	return false
}
