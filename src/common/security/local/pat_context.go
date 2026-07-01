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
	"encoding/json"
	"sync"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	patmodel "github.com/goharbor/harbor/src/pkg/pat/model"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

type PATSecurityContext struct {
	user      *models.User
	scope     []patmodel.ProjectScope
	evaluator evaluator.Evaluator
	once      sync.Once
}

func NewPATSecurityContext(user *models.User, scope string) *PATSecurityContext {
	var parsedScope []patmodel.ProjectScope
	if scope != "" {
		_ = json.Unmarshal([]byte(scope), &parsedScope)
	}
	return &PATSecurityContext{
		user:  user,
		scope: parsedScope,
	}
}

func (s *PATSecurityContext) Name() string {
	return ContextName
}

func (s *PATSecurityContext) IsAuthenticated() bool {
	return s.user != nil
}

func (s *PATSecurityContext) GetUsername() string {
	if !s.IsAuthenticated() {
		return ""
	}
	return s.user.Username
}

func (s *PATSecurityContext) User() *models.User {
	return s.user
}

func (s *PATSecurityContext) IsSysAdmin() bool {
	if !s.IsAuthenticated() {
		return false
	}
	return s.user.SysAdminFlag || s.user.AdminRoleInAuth
}

func (s *PATSecurityContext) IsSolutionUser() bool {
	return false
}

func (s *PATSecurityContext) Can(ctx context.Context, action types.Action, resource types.Resource) bool {
	s.once.Do(func() {
		s.evaluator = &patScopeEvaluator{
			scope: s.scope,
		}
	})

	return s.evaluator != nil && s.evaluator.HasPermission(ctx, resource, action)
}

type patScopeEvaluator struct {
	scope []patmodel.ProjectScope
}

func (e *patScopeEvaluator) HasPermission(_ context.Context, resource types.Resource, action types.Action) bool {
	resourceStr := resource.String()
	actionStr := action.String()

	for _, projectScope := range e.scope {
		for _, access := range projectScope.Access {
			if access.Resource == resourceStr {
				for _, a := range access.Actions {
					if a == "*" || a == actionStr {
						return true
					}
				}
			}
		}
	}

	return false
}

func init() {
	_ = rbac.ActionPull
	_ = rbac.ActionPush
	_ = rbac.ActionDelete
}
