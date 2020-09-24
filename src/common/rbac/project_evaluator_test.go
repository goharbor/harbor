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

package rbac

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/project"
	"github.com/goharbor/harbor/src/common/models"
	promgr "github.com/goharbor/harbor/src/core/promgr/mocks"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/testing/common/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	projectID = int64(1)

	projectAdminSecurity = makeMockSecurity("projectAdmin", common.RoleProjectAdmin)
	guestSecurity        = makeMockSecurity("guest", common.RoleGuest)
	anonymousSecurity    = makeMockSecurity("")

	publicProjectManager  = makeMockProjectManager(projectID, true)
	privateProjectManager = makeMockProjectManager(projectID, false)
)

func makeMockSecurity(username string, roles ...int) *security.Context {
	var isAuthenticated bool
	if username != "" {
		isAuthenticated = true
	}

	ctx := &security.Context{}
	ctx.On("IsAuthenticated").Return(isAuthenticated)
	ctx.On("GetUsername").Return(username)
	ctx.On("GetProjectRoles", mock.AnythingOfType("int64")).Return(roles)

	return ctx
}

func makeMockProjectManager(projectID int64, isPublic bool) *promgr.ProjectManager {
	pm := &promgr.ProjectManager{}

	project := &models.Project{ProjectID: projectID}
	if isPublic {
		project.SetMetadata(models.ProMetaPublic, "true")
	} else {
		project.SetMetadata(models.ProMetaPublic, "false")
	}

	pm.On("Get", projectID).Return(project, nil)

	return pm
}

func makeResource(subresource ...types.Resource) types.Resource {
	return NewProjectNamespace(projectID).Resource(subresource...)
}

func TestAnonymousAccess(t *testing.T) {
	assert := assert.New(t)
	evaluator1 := NewProjectUserEvaluator(nil, publicProjectManager)
	assert.True(evaluator1.HasPermission(makeResource(ResourceRepository), ActionPull))

	evaluator2 := NewProjectUserEvaluator(nil, privateProjectManager)
	assert.False(evaluator2.HasPermission(makeResource(ResourceRepository), ActionPull))

	evaluator3 := NewProjectRobotEvaluator(anonymousSecurity, publicProjectManager, func(ns types.Namespace) types.RBACUser { return nil })
	assert.True(evaluator3.HasPermission(makeResource(ResourceRepository), ActionPull))

	evaluator4 := NewProjectRobotEvaluator(anonymousSecurity, privateProjectManager, func(ns types.Namespace) types.RBACUser { return nil })
	assert.False(evaluator4.HasPermission(makeResource(ResourceRepository), ActionPull))
}

func TestProjectRoleAccess(t *testing.T) {
	assert := assert.New(t)
	dao.PrepareTestForPostgresSQL()

	projectID, err := dao.AddProject(models.Project{
		OwnerID: 1,
		Name:    "project_for_test_evaluator",
	})
	require.Nil(t, err)
	defer dao.DeleteProject(projectID)

	pm := makeMockProjectManager(projectID, true)

	memberID, err := project.AddProjectMember(models.Member{
		ProjectID:  projectID,
		Role:       common.RoleProjectAdmin,
		EntityID:   1,
		EntityType: "u",
	})
	require.Nil(t, err)
	defer project.DeleteProjectMemberByID(memberID)
	evaluator1 := NewProjectUserEvaluator(&models.User{
		UserID:   1,
		Username: "admin",
	}, pm)
	assert.True(evaluator1.HasPermission(NewProjectNamespace(projectID).Resource(ResourceRepository), ActionPush))

	project.UpdateProjectMemberRole(memberID, common.RoleGuest)
	evaluator2 := NewProjectUserEvaluator(&models.User{
		UserID:   1,
		Username: "admin",
	}, pm)
	assert.False(evaluator2.HasPermission(NewProjectNamespace(projectID).Resource(ResourceRepository), ActionPush))
}

func BenchmarkProjectRBACEvaluator(b *testing.B) {
	evaluator := NewProjectUserEvaluator(nil, publicProjectManager)
	resource := NewProjectNamespace(projectID).Resource(ResourceRepository)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.HasPermission(resource, ActionPull)
	}
}

func BenchmarkProjectRBACEvaluatorParallel(b *testing.B) {
	evaluator := NewProjectUserEvaluator(nil, publicProjectManager)
	resource := NewProjectNamespace(projectID).Resource(ResourceRepository)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			evaluator.HasPermission(resource, ActionPull)
		}
	})
}
