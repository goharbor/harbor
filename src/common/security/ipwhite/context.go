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

package ipwhite

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"strings"
)

// SecurityContext implements security.Context interface based on IP White
type SecurityContext struct {
	UserIP string
}

// IsAuthenticated returns true if the user has been authenticated
func (s *SecurityContext) IsAuthenticated() bool {
	return true
}

// GetUsername returns the username of the authenticated user
// It returns null if the user has not been authenticated
func (s *SecurityContext) GetUsername() string {
	if !s.IsAuthenticated() {
		return ""
	}
	// todo to config
	return "robot-" + strings.Replace(s.UserIP, ".", "-", -1)
}

// IsSysAdmin returns whether the authenticated user is system admin
// It returns false if the user has not been authenticated
func (s *SecurityContext) IsSysAdmin() bool {
	return false
}

// IsSolutionUser ...
func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

// Can returns whether the user can do action on resource
func (s *SecurityContext) Can(action rbac.Action, resource rbac.Resource) bool {
	if action == "pull" || action == "push" {
		return true
	}
	return false
}

// GetProjectRoles ...
func (s *SecurityContext) GetProjectRoles(projectIDOrName interface{}) []int {
	return []int{}
}

// GetMyProjects ...
func (s *SecurityContext) GetMyProjects() ([]*models.Project, error) {
	return nil, nil
}
