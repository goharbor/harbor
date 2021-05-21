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
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/assert"
	"testing"
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

	projectAdminUser = &models.User{
		Username: "projectAdminUser",
		Email:    "projectAdminUser@vmware.com",
	}
	developerUser = &models.User{
		Username: "developerUser",
		Email:    "developerUser@vmware.com",
	}
	guestUser = &models.User{
		Username: "guestUser",
		Email:    "guestUser@vmware.com",
	}
)

func TestIsAuthenticated(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil)
	assert.False(t, ctx.IsAuthenticated())

	// authenticated
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	})
	assert.True(t, ctx.IsAuthenticated())
}

func TestGetUsername(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil)
	assert.Equal(t, "", ctx.GetUsername())

	// authenticated
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	})
	assert.Equal(t, "test", ctx.GetUsername())
}

func TestIsSysAdmin(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil)
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, non admin
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	})
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, admin
	ctx = NewSecurityContext(&models.User{
		Username:     "test",
		SysAdminFlag: true,
	})
	assert.True(t, ctx.IsSysAdmin())
}

func TestIsSolutionUser(t *testing.T) {
	ctx := NewSecurityContext(nil)
	assert.False(t, ctx.IsSolutionUser())
}

func TestHasPullPerm(t *testing.T) {
	{
		// public project
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(public, nil)

		ctx := NewSecurityContext(nil)
		ctx.ctl = ctl
		resource := rbac_project.NewNamespace(1).Resource(rbac.ResourceRepository)
		assert.True(t, ctx.Can(context.TODO(), rbac.ActionPull, resource))
	}

	{
		// private project, unauthenticated
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(private, nil)

		ctx := NewSecurityContext(nil)
		ctx.ctl = ctl
		resource := rbac_project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
		assert.False(t, ctx.Can(context.TODO(), rbac.ActionPull, resource))
	}

	{
		// private project, authenticated, has no perm
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(private, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{}, nil)

		ctx := NewSecurityContext(&models.User{Username: "test"})
		ctx.ctl = ctl
		resource := rbac_project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
		assert.False(t, ctx.Can(context.TODO(), rbac.ActionPull, resource))
	}

	{
		// private project, authenticated, has read perm
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(private, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleGuest}, nil)

		ctx := NewSecurityContext(guestUser)
		ctx.ctl = ctl
		resource := rbac_project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
		assert.True(t, ctx.Can(context.TODO(), rbac.ActionPull, resource))
	}

	{
		// private project, authenticated, system admin
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(private, nil)

		ctx := NewSecurityContext(&models.User{
			Username:     "admin",
			SysAdminFlag: true,
		})
		ctx.ctl = ctl
		resource := rbac_project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
		assert.True(t, ctx.Can(context.TODO(), rbac.ActionPull, resource))
	}
}

func TestHasPushPerm(t *testing.T) {
	resource := rbac_project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)

	{
		// unauthenticated
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(private, nil)

		ctx := NewSecurityContext(nil)
		ctx.ctl = ctl
		assert.False(t, ctx.Can(context.TODO(), rbac.ActionPush, resource))
	}

	{
		// authenticated, has read perm
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(private, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleGuest}, nil)

		ctx := NewSecurityContext(guestUser)
		ctx.ctl = ctl
		assert.False(t, ctx.Can(context.TODO(), rbac.ActionPush, resource))
	}

	{
		// authenticated, has write perm
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(private, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleDeveloper}, nil)

		ctx := NewSecurityContext(developerUser)
		ctx.ctl = ctl
		assert.True(t, ctx.Can(context.TODO(), rbac.ActionPush, resource))
	}

	{
		// authenticated, system admin
		ctl := &projecttesting.Controller{}
		ctx := NewSecurityContext(&models.User{
			Username:     "admin",
			SysAdminFlag: true,
		})
		ctx.ctl = ctl
		assert.True(t, ctx.Can(context.TODO(), rbac.ActionPush, resource))
	}
}

func TestHasPushPullPerm(t *testing.T) {
	resource := rbac_project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)

	{
		// unauthenticated
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(private, nil)

		ctx := NewSecurityContext(nil)
		ctx.ctl = ctl
		assert.False(t, ctx.Can(context.TODO(), rbac.ActionPush, resource) && ctx.Can(context.TODO(), rbac.ActionPull, resource))
	}

	{
		// authenticated, has all perms
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(private, nil)
		mock.OnAnything(ctl, "ListRoles").Return([]int{common.RoleProjectAdmin}, nil)

		ctx := NewSecurityContext(projectAdminUser)
		ctx.ctl = ctl
		assert.True(t, ctx.Can(context.TODO(), rbac.ActionPush, resource) && ctx.Can(context.TODO(), rbac.ActionPull, resource))
	}

	{
		// authenticated, system admin
		ctl := &projecttesting.Controller{}
		mock.OnAnything(ctl, "Get").Return(private, nil)

		ctx := NewSecurityContext(&models.User{
			Username:     "admin",
			SysAdminFlag: true,
		})
		ctx.ctl = ctl
		assert.True(t, ctx.Can(context.TODO(), rbac.ActionPush, resource) && ctx.Can(context.TODO(), rbac.ActionPull, resource))
	}
}

func TestSysadminPerms(t *testing.T) {
	// authenticated, system admin
	ctl := &projecttesting.Controller{}
	mock.OnAnything(ctl, "Get").Return(private, nil)
	mock.OnAnything(ctl, "ListRoles").Return([]int{}, nil)

	ctx := NewSecurityContext(&models.User{
		Username:     "admin",
		SysAdminFlag: true,
	})
	ctx.ctl = ctl
	resource := rbac_project.NewNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(context.TODO(), rbac.ActionPush, resource) && ctx.Can(context.TODO(), rbac.ActionPull, resource))
	assert.False(t, ctx.Can(context.TODO(), rbac.ActionScannerPull, resource))

}
