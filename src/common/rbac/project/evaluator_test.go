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
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	roleCtl "github.com/goharbor/harbor/src/controller/role"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/testing/mock"
	mocksProject "github.com/goharbor/harbor/src/testing/controller/project"
)

// stubRoleCtl implements roleCtl.Controller for tests.
// Only Get() needs real behaviour; all other methods are no-ops.
type stubRoleCtl struct {
	testifymock.Mock
}

func (s *stubRoleCtl) Get(ctx context.Context, id int64, option *roleCtl.Option) (*roleCtl.Role, error) {
	args := s.Called(ctx, id, option)
	r, _ := args.Get(0).(*roleCtl.Role)
	return r, args.Error(1)
}
func (s *stubRoleCtl) Create(ctx context.Context, r *roleCtl.Role) (int64, error) { return 0, nil }
func (s *stubRoleCtl) Delete(ctx context.Context, id int64, opt ...*roleCtl.Option) error {
	return nil
}
func (s *stubRoleCtl) Update(ctx context.Context, r *roleCtl.Role, opt *roleCtl.Option) error {
	return nil
}
func (s *stubRoleCtl) List(ctx context.Context, query *q.Query, opt *roleCtl.Option) ([]*roleCtl.Role, error) {
	return nil, nil
}
func (s *stubRoleCtl) Count(ctx context.Context, query *q.Query) (int64, error) { return 0, nil }

var (
	public = &proModels.Project{
		ProjectID: 1,
		Name:      "public_project",
		OwnerID:   1,
		Metadata:  map[string]string{"public": "true"},
	}
	private = &proModels.Project{
		ProjectID: 2,
		Name:      "private_project",
		OwnerID:   1,
		Metadata:  map[string]string{"public": "false"},
	}
)

// builtinRole returns a *roleCtl.Role with the permissions for a built-in role ID,
// sourced from the same static map used by the upstream.
func builtinRole(roleID int) *roleCtl.Role {
	nameMap := map[int]string{
		common.RoleProjectAdmin: "projectAdmin",
		common.RoleMaintainer:   "maintainer",
		common.RoleDeveloper:    "developer",
		common.RoleGuest:        "guest",
		common.RoleLimitedGuest: "limitedGuest",
	}
	policies := rolePoliciesMap[nameMap[roleID]]
	access := make([]*types.Policy, len(policies))
	copy(access, policies)
	return &roleCtl.Role{
		Permissions: []*roleCtl.Permission{{
			Kind:      roleCtl.LEVELROLE,
			Namespace: "*",
			Access:    access,
		}},
	}
}

func TestAnonymousAccess(t *testing.T) {
	// anonymous can pull from a public project
	ctl := &mocksProject.Controller{}
	ctl_r := &stubRoleCtl{}
	mock.OnAnything(ctl, "Get").Return(public, nil)
	resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t,
		NewEvaluator(ctl, NewBuilderForUser(nil, ctl, ctl_r)).HasPermission(context.TODO(), resource, rbac.ActionPull))

	// anonymous cannot pull from a private project
	ctl2 := &mocksProject.Controller{}
	ctl_r2 := &stubRoleCtl{}
	mock.OnAnything(ctl2, "Get").Return(private, nil)
	resource2 := NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.False(t,
		NewEvaluator(ctl2, NewBuilderForUser(nil, ctl2, ctl_r2)).HasPermission(context.TODO(), resource2, rbac.ActionPull))
}

func TestProjectRoleAccess(t *testing.T) {
	user := &models.User{UserID: 1, Username: "username"}

	// projectAdmin can push
	{
		ctl := &mocksProject.Controller{}
		ctl_r := &stubRoleCtl{}
		mock.OnAnything(ctl, "Get").Return(public, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleProjectAdmin}, nil)
		ctl_r.On("Get", testifymock.Anything, int64(common.RoleProjectAdmin), testifymock.Anything).
			Return(builtinRole(common.RoleProjectAdmin), nil)

		evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl, ctl_r))
		resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
		assert.True(t, evaluator.HasPermission(context.TODO(), resource, rbac.ActionPush))
	}

	// guest cannot push
	{
		ctl := &mocksProject.Controller{}
		ctl_r := &stubRoleCtl{}
		mock.OnAnything(ctl, "Get").Return(public, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleGuest}, nil)
		ctl_r.On("Get", testifymock.Anything, int64(common.RoleGuest), testifymock.Anything).
			Return(builtinRole(common.RoleGuest), nil)

		evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl, ctl_r))
		resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
		assert.False(t, evaluator.HasPermission(context.TODO(), resource, rbac.ActionPush))
	}
}

func BenchmarkProjectEvaluator(b *testing.B) {
	user := &models.User{UserID: 1, Username: "username"}
	ctl := &mocksProject.Controller{}
	ctl_r := &stubRoleCtl{}
	mock.OnAnything(ctl, "Get").Return(public, nil)
	mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleProjectAdmin}, nil)
	ctl_r.On("Get", testifymock.Anything, int64(common.RoleProjectAdmin), testifymock.Anything).
		Return(builtinRole(common.RoleProjectAdmin), nil)

	evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl, ctl_r))
	resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)

	b.ResetTimer()
	for b.Loop() {
		evaluator.HasPermission(context.TODO(), resource, rbac.ActionPull)
	}
}

func BenchmarkProjectEvaluatorParallel(b *testing.B) {
	user := &models.User{UserID: 1, Username: "username"}
	ctl := &mocksProject.Controller{}
	ctl_r := &stubRoleCtl{}
	mock.OnAnything(ctl, "Get").Return(public, nil)
	mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleProjectAdmin}, nil)
	ctl_r.On("Get", testifymock.Anything, int64(common.RoleProjectAdmin), testifymock.Anything).
		Return(builtinRole(common.RoleProjectAdmin), nil)

	evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl, ctl_r))
	resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			evaluator.HasPermission(context.TODO(), resource, rbac.ActionPull)
		}
	})
}
