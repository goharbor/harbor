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
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	rbacevaluator "github.com/goharbor/harbor/src/pkg/permission/evaluator/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
)

var benchProject = &proModels.Project{
	ProjectID: 42,
	Name:      "bench",
	OwnerID:   1,
	Metadata:  map[string]string{"public": "false"},
}

func benchUser(project *proModels.Project, roleID int) types.RBACUser {
	return &rbacUser{
		project:      project,
		username:     "alice",
		projectRoles: []*projectRBACRole{{projectID: project.ProjectID, roleID: roleID}},
	}
}

func BenchmarkHasPermission(b *testing.B) {
	resource := NewNamespace(benchProject.ProjectID).Resource(rbac.ResourceRepository)
	ctx := context.TODO()

	b.ReportAllocs()
	for b.Loop() {
		e := rbacevaluator.New(benchUser(benchProject, common.RoleProjectAdmin))
		if !e.HasPermission(ctx, resource, rbac.ActionPull) {
			b.Fatal("permission denied")
		}
	}
}

func BenchmarkHasPermissionDenied(b *testing.B) {
	resource := NewNamespace(benchProject.ProjectID).Resource(rbac.ResourceRepository)
	ctx := context.TODO()

	b.ReportAllocs()
	for b.Loop() {
		e := rbacevaluator.New(benchUser(benchProject, common.RoleGuest))
		if e.HasPermission(ctx, resource, rbac.ActionPush) {
			b.Fatal("permission granted")
		}
	}
}

func BenchmarkHasPermissionManyProjects(b *testing.B) {
	const projects = 10000
	resources := make([]types.Resource, projects)
	users := make([]types.RBACUser, projects)
	for i := range projects {
		p := &proModels.Project{ProjectID: int64(i + 1), Name: fmt.Sprintf("p%d", i), Metadata: map[string]string{"public": "false"}}
		resources[i] = NewNamespace(p.ProjectID).Resource(rbac.ResourceRepository)
		users[i] = benchUser(p, common.RoleProjectAdmin)
	}
	ctx := context.TODO()

	b.ReportAllocs()
	i := 0
	for b.Loop() {
		n := i % projects
		i++
		e := rbacevaluator.New(users[n])
		if !e.HasPermission(ctx, resources[n], rbac.ActionPull) {
			b.Fatal("permission denied")
		}
	}
}

func BenchmarkHasPermissionPublicProject(b *testing.B) {
	public := &proModels.Project{ProjectID: 43, Name: "public", Metadata: map[string]string{"public": "true"}}
	resource := NewNamespace(public.ProjectID).Resource(rbac.ResourceRepository)
	ctx := context.TODO()

	b.ReportAllocs()
	for b.Loop() {
		e := rbacevaluator.New(&rbacUser{project: public, username: "anonymous"})
		if !e.HasPermission(ctx, resource, rbac.ActionPull) {
			b.Fatal("permission denied")
		}
	}
}
