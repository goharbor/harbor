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

package secret

import (
	"fmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// SecurityContext implements security.Context interface based on secret store
type SecurityContext struct {
	secret string
	store  *secret.Store
}

// NewSecurityContext ...
func NewSecurityContext(secret string, store *secret.Store) *SecurityContext {
	return &SecurityContext{
		secret: secret,
		store:  store,
	}
}

// IsAuthenticated returns true if the secret is valid
func (s *SecurityContext) IsAuthenticated() bool {
	if s.store == nil {
		log.Debug("secret store is nil")
		return false
	}
	valid := s.store.IsValid(s.secret)
	if !valid {
		log.Debugf("invalid secret: %s", s.secret)
	}

	return valid
}

// GetUsername returns the corresponding username of the secret
// or null if the secret is invalid
func (s *SecurityContext) GetUsername() string {
	if s.store == nil {
		return ""
	}
	return s.store.GetUsername(s.secret)
}

// IsSysAdmin always returns false
func (s *SecurityContext) IsSysAdmin() bool {
	return false
}

// PolicyCheck ...
func (s *SecurityContext) PolicyCheck() bool {
	return true
}

// IsSolutionUser ...
func (s *SecurityContext) IsSolutionUser() bool {
	return s.IsAuthenticated()
}

// Can returns whether the user can do action on resource
// returns true if the corresponding user of the secret
// is jobservice or core service, otherwise returns false
func (s *SecurityContext) Can(action rbac.Action, resource rbac.Resource) bool {
	if s.store == nil {
		return false
	}
	return s.store.GetUsername(s.secret) == secret.JobserviceUser || s.store.GetUsername(s.secret) == secret.CoreUser
}

// GetMyProjects ...
func (s *SecurityContext) GetMyProjects() ([]*models.Project, error) {
	return nil, fmt.Errorf("GetMyProjects is unsupported")
}

// GetProjectRoles return guest role if has read permission, otherwise return nil
func (s *SecurityContext) GetProjectRoles(projectIDOrName interface{}) []int {
	roles := []int{}
	if s.store != nil &&
		(s.store.GetUsername(s.secret) == secret.JobserviceUser ||
			s.store.GetUsername(s.secret) == secret.CoreUser) {
		roles = append(roles, common.RoleGuest)
	}
	return roles
}
