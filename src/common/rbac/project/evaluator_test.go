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
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
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

// customRole returns a *roleCtl.Role carrying the given project-scoped accesses,
// as the DB-backed role controller would return for a custom role.
func customRole(name string, access ...*types.Policy) *roleCtl.Role {
	r := &roleCtl.Role{
		Permissions: []*roleCtl.Permission{{
			Kind:      roleCtl.LEVELROLE,
			Namespace: "*",
			Access:    access,
		}},
	}
	r.Name = name
	return r
}

var (
	public = &proModels.Project{
		ProjectID: 1,
		Name:      "public_project",
		OwnerID:   1,
		Metadata: map[string]string{
			"public": "true",
		},
	}

	private = &proModels.Project{
		ProjectID: 2,
		Name:      "private_project",
		OwnerID:   1,
		Metadata: map[string]string{
			"public": "false",
		},
	}
)

func TestAnonymousAccess(t *testing.T) {
	assert := assert.New(t)

	{
		// anonymous to access public project
		ctl := &projecttesting.Controller{}
		ctl_r := &stubRoleCtl{}
		mock.OnAnything(ctl, "Get").Return(public, nil)

		resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)

		evaluator := NewEvaluator(ctl, NewBuilderForUser(nil, ctl, ctl_r))
		assert.True(evaluator.HasPermission(context.TODO(), resource, rbac.ActionPull))
	}

	{
		// anonymous to access private project
		ctl := &projecttesting.Controller{}
		ctl_r := &stubRoleCtl{}
		mock.OnAnything(ctl, "Get").Return(private, nil)

		resource := NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)

		evaluator := NewEvaluator(ctl, NewBuilderForUser(nil, ctl, ctl_r))
		assert.False(evaluator.HasPermission(context.TODO(), resource, rbac.ActionPull))
	}
}

func TestProjectRoleAccess(t *testing.T) {
	assert := assert.New(t)

	{
		ctl := &projecttesting.Controller{}
		ctl_r := &stubRoleCtl{}
		mock.OnAnything(ctl, "Get").Return(public, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleProjectAdmin}, nil)

		user := &models.User{
			UserID:   1,
			Username: "username",
		}
		evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl, ctl_r))
		resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
		assert.True(evaluator.HasPermission(context.TODO(), resource, rbac.ActionPush))
		// built-in roles resolve from the compile-time map — no DB lookup
		ctl_r.AssertNotCalled(t, "Get")
	}

	{
		ctl := &projecttesting.Controller{}
		ctl_r := &stubRoleCtl{}
		mock.OnAnything(ctl, "Get").Return(public, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleGuest}, nil)

		user := &models.User{
			UserID:   1,
			Username: "username",
		}
		evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl, ctl_r))
		resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
		assert.False(evaluator.HasPermission(context.TODO(), resource, rbac.ActionPush))
		ctl_r.AssertNotCalled(t, "Get")
	}
}

func TestCustomProjectRoleAccess(t *testing.T) {
	assert := assert.New(t)
	user := &models.User{
		UserID:   1,
		Username: "username",
	}
	const customRoleID = 100 // any ID outside the built-in range (1-5)

	{
		// a custom role granting repository:push is loaded from the DB controller
		ctl := &projecttesting.Controller{}
		ctl_r := &stubRoleCtl{}
		mock.OnAnything(ctl, "Get").Return(public, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{customRoleID}, nil)
		ctl_r.On("Get", testifymock.Anything, int64(customRoleID), testifymock.Anything).
			Return(customRole("pusher", &types.Policy{Resource: rbac.ResourceRepository, Action: rbac.ActionPush}), nil)

		evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl, ctl_r))
		resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
		assert.True(evaluator.HasPermission(context.TODO(), resource, rbac.ActionPush))
		ctl_r.AssertCalled(t, "Get", testifymock.Anything, int64(customRoleID), testifymock.Anything)
	}

	{
		// a custom role without repository:push cannot push
		ctl := &projecttesting.Controller{}
		ctl_r := &stubRoleCtl{}
		mock.OnAnything(ctl, "Get").Return(public, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{customRoleID}, nil)
		ctl_r.On("Get", testifymock.Anything, int64(customRoleID), testifymock.Anything).
			Return(customRole("puller", &types.Policy{Resource: rbac.ResourceRepository, Action: rbac.ActionPull}), nil)

		evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl, ctl_r))
		resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
		assert.False(evaluator.HasPermission(context.TODO(), resource, rbac.ActionPush))
	}
}

func BenchmarkProjectEvaluator(b *testing.B) {
	ctl := &projecttesting.Controller{}
	ctl_r := &stubRoleCtl{}
	mock.OnAnything(ctl, "Get").Return(public, nil)
	mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleProjectAdmin}, nil)

	user := &models.User{
		UserID:   1,
		Username: "username",
	}
	evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl, ctl_r))
	resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)

	b.ResetTimer()
	for b.Loop() {
		evaluator.HasPermission(context.TODO(), resource, rbac.ActionPull)
	}
}

func BenchmarkProjectEvaluatorParallel(b *testing.B) {
	ctl := &projecttesting.Controller{}
	ctl_r := &stubRoleCtl{}
	mock.OnAnything(ctl, "Get").Return(public, nil)
	mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleProjectAdmin}, nil)

	user := &models.User{
		UserID:   1,
		Username: "username",
	}
	evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl, ctl_r))
	resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			evaluator.HasPermission(context.TODO(), resource, rbac.ActionPull)
		}
	})
}
