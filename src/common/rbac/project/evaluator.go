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

package project

import (
	"context"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/namespace"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
)

// RBACUserBuilder builder to make types.RBACUser for the project
type RBACUserBuilder func(context.Context, *proModels.Project) types.RBACUser

// NewBuilderForUser create a builder for the local user
func NewBuilderForUser(user *models.User, ctl project.Controller) RBACUserBuilder {
	return func(ctx context.Context, p *proModels.Project) types.RBACUser {
		if user == nil {
			// anonymous access
			return &rbacUser{
				project:  p,
				username: "anonymous",
			}
		}

		roles, err := ctl.ListRoles(ctx, p.ProjectID, user)
		if err != nil {
			log.Errorf("failed to list roles: %v", err)
			return nil
		}

		return &rbacUser{
			project:      p,
			username:     user.Username,
			projectRoles: roles,
		}
	}
}

// NewBuilderForPolicies create a builder for the policies
func NewBuilderForPolicies(username string, policies []*types.Policy,
	filters ...func(*proModels.Project, []*types.Policy) []*types.Policy) RBACUserBuilder {

	return func(ctx context.Context, p *proModels.Project) types.RBACUser {
		for _, filter := range filters {
			policies = filter(p, policies)
		}

		return &rbacUser{
			project:  p,
			username: username,
			policies: policies,
		}
	}
}

// NewEvaluator create evaluator for the project by builders
func NewEvaluator(ctl project.Controller, builders ...RBACUserBuilder) evaluator.Evaluator {
	return namespace.New(NamespaceKind, func(ctx context.Context, ns types.Namespace) evaluator.Evaluator {
		p, err := ctl.Get(ctx, ns.Identity().(int64), project.Metadata(true))
		if err != nil {
			if err != nil {
				log.Warningf("Failed to get info of project %d for permission evaluator, error: %v", ns.Identity(), err)
			}
			return nil
		}

		var rbacUsers []types.RBACUser
		for _, builder := range builders {
			if rbacUser := builder(ctx, p); rbacUser != nil {
				rbacUsers = append(rbacUsers, rbacUser)
			}
		}

		switch len(rbacUsers) {
		case 0:
			return nil
		case 1:
			return rbac.New(rbacUsers[0])
		default:
			var evaluators evaluator.Evaluators
			for _, rbacUser := range rbacUsers {
				evaluators = evaluators.Add(rbac.New(rbacUser))
			}

			return evaluators
		}
	})
}
