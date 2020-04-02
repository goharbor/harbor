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

package local

import (
	"sync"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/admin"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// SecurityContext implements security.Context interface based on database
type SecurityContext struct {
	user      *models.User
	pm        promgr.ProjectManager
	evaluator evaluator.Evaluator
	once      sync.Once
}

// NewSecurityContext ...
func NewSecurityContext(user *models.User, pm promgr.ProjectManager) *SecurityContext {
	return &SecurityContext{
		user: user,
		pm:   pm,
	}
}

// Name returns the name of the security context
func (s *SecurityContext) Name() string {
	return "local"
}

// IsAuthenticated returns true if the user has been authenticated
func (s *SecurityContext) IsAuthenticated() bool {
	return s.user != nil
}

// GetUsername returns the username of the authenticated user
// It returns null if the user has not been authenticated
func (s *SecurityContext) GetUsername() string {
	if !s.IsAuthenticated() {
		return ""
	}
	return s.user.Username
}

// User get the current user
func (s *SecurityContext) User() *models.User {
	return s.user
}

// IsSysAdmin returns whether the authenticated user is system admin
// It returns false if the user has not been authenticated
func (s *SecurityContext) IsSysAdmin() bool {
	if !s.IsAuthenticated() {
		return false
	}
	return s.user.SysAdminFlag || s.user.AdminRoleInAuth
}

// IsSolutionUser ...
func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

// Can returns whether the user can do action on resource
func (s *SecurityContext) Can(action types.Action, resource types.Resource) bool {
	s.once.Do(func() {
		var evaluators evaluator.Evaluators
		if s.IsSysAdmin() {
			evaluators = evaluators.Add(admin.New(s.GetUsername()))
		}
		evaluators = evaluators.Add(rbac.NewProjectRBACEvaluator(s, s.pm))

		s.evaluator = evaluators
	})

	return s.evaluator != nil && s.evaluator.HasPermission(resource, action)
}

// GetProjectRoles ...
func (s *SecurityContext) GetProjectRoles(projectIDOrName interface{}) []int {
	if !s.IsAuthenticated() || projectIDOrName == nil {
		return []int{}
	}

	roles := []int{}
	user, err := dao.GetUser(models.User{
		Username: s.GetUsername(),
	})
	if err != nil {
		log.Errorf("failed to get user %s: %v", s.GetUsername(), err)
		return roles
	}
	if user == nil {
		log.Debugf("user %s not found", s.GetUsername())
		return roles
	}
	project, err := s.pm.Get(projectIDOrName)
	if err != nil {
		log.Errorf("failed to get project %v: %v", projectIDOrName, err)
		return roles
	}
	if project == nil {
		log.Errorf("project %v not found", projectIDOrName)
		return roles
	}
	roleList, err := dao.GetUserProjectRoles(user.UserID, project.ProjectID, common.UserMember)
	if err != nil {
		log.Errorf("failed to get roles of user %d to project %d: %v", user.UserID, project.ProjectID, err)
		return roles
	}
	for _, role := range roleList {
		switch role.RoleCode {
		case "MDRWS":
			roles = append(roles, common.RoleProjectAdmin)
		case "DRWS":
			roles = append(roles, common.RoleMaster)
		case "RWS":
			roles = append(roles, common.RoleDeveloper)
		case "RS":
			roles = append(roles, common.RoleGuest)
		case "LRS":
			roles = append(roles, common.RoleLimitedGuest)
		}
	}
	return mergeRoles(roles, s.GetRolesByGroup(projectIDOrName))
}

func mergeRoles(rolesA, rolesB []int) []int {
	type void struct{}
	var roles []int
	var placeHolder void
	roleSet := make(map[int]void)
	for _, r := range rolesA {
		roleSet[r] = placeHolder
	}
	for _, r := range rolesB {
		roleSet[r] = placeHolder
	}
	for r := range roleSet {
		roles = append(roles, r)
	}
	return roles
}

// GetRolesByGroup - Get the group role of current user to the project
func (s *SecurityContext) GetRolesByGroup(projectIDOrName interface{}) []int {
	var roles []int
	user := s.user
	project, err := s.pm.Get(projectIDOrName)
	// No user, group or project info
	if err != nil || project == nil || user == nil || len(user.GroupIDs) == 0 {
		return roles
	}
	// Get role by Group ID
	roles, err = dao.GetRolesByGroupID(project.ProjectID, user.GroupIDs)
	if err != nil {
		return nil
	}
	return roles
}

// GetMyProjects ...
func (s *SecurityContext) GetMyProjects() ([]*models.Project, error) {
	result, err := s.pm.List(
		&models.ProjectQueryParam{
			Member: &models.MemberQuery{
				Name:     s.GetUsername(),
				GroupIDs: s.user.GroupIDs,
			},
		})
	if err != nil {
		return nil, err
	}

	return result.Projects, nil
}
