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

// These benchmarks measure this branch. The figures quoted for the base commit
// were taken by running the equivalent questions against 37dc02fd with a
// throwaway harness: this file cannot be compiled there as it stands, because
// it seeds through src/testing/pkg/permission/policy and builds an rbacUser
// with the projectRoles field, both of which arrive with this change.

package project

import (
	"context"
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	rbacevaluator "github.com/goharbor/harbor/src/pkg/permission/evaluator/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	policytesting "github.com/goharbor/harbor/src/testing/pkg/permission/policy"
)

var benchProject = &proModels.Project{
	ProjectID: 42,
	Name:      "bench",
	OwnerID:   1,
	Metadata:  map[string]string{"public": "false"},
}

func benchUser(roles ...int) types.RBACUser {
	return &rbacUser{project: benchProject, username: "alice", projectRoles: roles}
}

func BenchmarkHasPermission(b *testing.B) {
	policytesting.Seed(b)
	resource := NewNamespace(benchProject.ProjectID).Resource(rbac.ResourceRepository)
	ctx := context.TODO()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := rbacevaluator.New(benchUser(common.RoleProjectAdmin))
		if !e.HasPermission(ctx, resource, rbac.ActionPull) {
			b.Fatal("permission denied")
		}
	}
}

func BenchmarkHasPermissionDenied(b *testing.B) {
	policytesting.Seed(b)
	resource := NewNamespace(benchProject.ProjectID).Resource(rbac.ResourceRepository)
	ctx := context.TODO()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := rbacevaluator.New(benchUser(common.RoleGuest))
		if e.HasPermission(ctx, resource, rbac.ActionPush) {
			b.Fatal("permission granted")
		}
	}
}

func BenchmarkHasPermissionManyProjects(b *testing.B) {
	policytesting.Seed(b)
	const projects = 10000
	resources := make([]types.Resource, projects)
	for i := range resources {
		resources[i] = NewNamespace(int64(i + 1)).Resource(rbac.ResourceRepository)
	}
	users := make([]types.RBACUser, projects)
	for i := range users {
		users[i] = &rbacUser{
			project:      &proModels.Project{ProjectID: int64(i + 1), Name: fmt.Sprintf("p%d", i), Metadata: map[string]string{"public": "false"}},
			username:     "alice",
			projectRoles: []int{common.RoleProjectAdmin},
		}
	}
	ctx := context.TODO()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := i % projects
		e := rbacevaluator.New(users[n])
		if !e.HasPermission(ctx, resources[n], rbac.ActionPull) {
			b.Fatal("permission denied")
		}
	}
}

func BenchmarkHasPermissionPublicProject(b *testing.B) {
	policytesting.Seed(b)
	public := &proModels.Project{ProjectID: 43, Name: "public", Metadata: map[string]string{"public": "true"}}
	resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
	ctx := context.TODO()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := rbacevaluator.New(&rbacUser{project: public, username: "anonymous"})
		if !e.HasPermission(ctx, resource, rbac.ActionPull) {
			b.Fatal("permission denied")
		}
	}
}
