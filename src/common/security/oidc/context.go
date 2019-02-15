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

package oidc

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/token"
	"github.com/goharbor/harbor/src/core/promgr"
)

// SecurityContext implements security.Context interface based on OIDC claims
type SecurityContext struct {
	claims *token.UserClaims
	user   *models.User
	pm     promgr.ProjectManager
}

// NewSecurityContext ...
func NewSecurityContext(user *models.User, pm promgr.ProjectManager) *SecurityContext {
	return &SecurityContext{
		user: user,
		pm:   pm,
	}
}

// IsAuthenticated returns true if the user has been authenticated
func (s *SecurityContext) IsAuthenticated() bool {
	return s.user != nil
}

func (s *SecurityContext) GetUsername() string {
	return s.user.Username
}

func (s *SecurityContext) IsSysAdmin() bool {
	return false
}

func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

func (s *SecurityContext) HasReadPerm(projectIDOrName interface{}) bool {
	return true
}

func (s *SecurityContext) HasWritePerm(projectIDOrName interface{}) bool {
	return true
}

func (s *SecurityContext) HasAllPerm(projectIDOrName interface{}) bool {
	return true
}

func (s *SecurityContext) GetMyProjects() ([]*models.Project, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *SecurityContext) GetProjectRoles(projectIDOrName interface{}) []int {
	return []int{0}
}

func (s *SecurityContext) Can(action rbac.Action, resource rbac.Resource) bool {
	return true
}
