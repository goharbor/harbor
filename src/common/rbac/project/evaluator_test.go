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
	"github.com/goharbor/harbor/src/common/rbac"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/assert"
)

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
		mock.OnAnything(ctl, "Get").Return(public, nil)

		resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)

		evaluator := NewEvaluator(ctl, NewBuilderForUser(nil, ctl))
		assert.True(evaluator.HasPermission(context.TODO(), resource, rbac.ActionPull))
	}

	{
		// anonymous to access private project
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(private, nil)

		resource := NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)

		evaluator := NewEvaluator(ctl, NewBuilderForUser(nil, ctl))
		assert.False(evaluator.HasPermission(context.TODO(), resource, rbac.ActionPull))
	}
}

func TestProjectRoleAccess(t *testing.T) {
	assert := assert.New(t)

	{
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(public, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleProjectAdmin}, nil)

		user := &models.User{
			UserID:   1,
			Username: "username",
		}
		evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl))
		resorce := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
		assert.True(evaluator.HasPermission(context.TODO(), resorce, rbac.ActionPush))
	}

	{
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(public, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleGuest}, nil)

		user := &models.User{
			UserID:   1,
			Username: "username",
		}
		evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl))
		resorce := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
		assert.False(evaluator.HasPermission(context.TODO(), resorce, rbac.ActionPush))
	}
}

func BenchmarkProjectEvaluator(b *testing.B) {
	ctl := &projecttesting.Controller{}
	mock.OnAnything(ctl, "Get").Return(public, nil)
	mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleProjectAdmin}, nil)

	user := &models.User{
		UserID:   1,
		Username: "username",
	}
	evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl))
	resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.HasPermission(context.TODO(), resource, rbac.ActionPull)
	}
}

func BenchmarkProjectEvaluatorParallel(b *testing.B) {
	ctl := &projecttesting.Controller{}
	mock.OnAnything(ctl, "Get").Return(public, nil)
	mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleProjectAdmin}, nil)

	user := &models.User{
		UserID:   1,
		Username: "username",
	}
	evaluator := NewEvaluator(ctl, NewBuilderForUser(user, ctl))
	resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			evaluator.HasPermission(context.TODO(), resource, rbac.ActionPull)
		}
	})
}
