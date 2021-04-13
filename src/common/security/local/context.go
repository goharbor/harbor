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
	"context"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"sync"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/admin"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// ContextName the name of the security context.
const ContextName = "local"

// SecurityContext implements security.Context interface based on database
type SecurityContext struct {
	user      *models.User
	ctl       project.Controller
	evaluator evaluator.Evaluator
	once      sync.Once
}

// NewSecurityContext ...
func NewSecurityContext(user *models.User) *SecurityContext {
	return &SecurityContext{
		user: user,
		ctl:  project.Ctl,
	}
}

// Name returns the name of the security context
func (s *SecurityContext) Name() string {
	return ContextName
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
func (s *SecurityContext) Can(ctx context.Context, action types.Action, resource types.Resource) bool {
	s.once.Do(func() {
		var evaluators evaluator.Evaluators
		if s.IsSysAdmin() {
			evaluators = evaluators.Add(admin.New(s.GetUsername()))
		}

		evaluators = evaluators.Add(rbac_project.NewEvaluator(s.ctl, rbac_project.NewBuilderForUser(s.user, s.ctl)))

		s.evaluator = evaluators
	})

	return s.evaluator != nil && s.evaluator.HasPermission(ctx, resource, action)
}
