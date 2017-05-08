// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/models"
)

type fakePM struct {
	public string
	roles  map[string][]int
}

func (f *fakePM) IsPublic(projectIDOrName interface{}) bool {
	return f.public == projectIDOrName.(string)
}
func (f *fakePM) GetRoles(username string, projectIDOrName interface{}) []int {
	return f.roles[projectIDOrName.(string)]
}

func TestIsAuthenticated(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil)
	assert.False(t, ctx.IsAuthenticated())

	// authenticated
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, nil)
	assert.True(t, ctx.IsAuthenticated())
}

func TestGetUsername(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil)
	assert.Equal(t, "", ctx.GetUsername())

	// authenticated
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, nil)
	assert.Equal(t, "test", ctx.GetUsername())
}

func TestIsSysAdmin(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil)
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, non admin
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, nil)
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, admin
	ctx = NewSecurityContext(&models.User{
		Username:     "test",
		HasAdminRole: 1,
	}, nil)
	assert.True(t, ctx.IsSysAdmin())
}

func TestHasReadPerm(t *testing.T) {
	pm := &fakePM{
		public: "public_project",
		roles: map[string][]int{
			"has_read_perm_project": []int{common.RoleGuest},
		},
	}

	// public project, unauthenticated
	ctx := NewSecurityContext(nil, pm)
	assert.True(t, ctx.HasReadPerm("public_project"))

	// private project, unauthenticated
	ctx = NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasReadPerm("has_read_perm_project"))

	// private project, authenticated, has no perm
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.False(t, ctx.HasReadPerm("has_no_perm_project"))

	// private project, authenticated, has read perm
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.True(t, ctx.HasReadPerm("has_read_perm_project"))

	// private project, authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "test",
		HasAdminRole: 1,
	}, pm)
	assert.True(t, ctx.HasReadPerm("has_no_perm_project"))
}

func TestHasWritePerm(t *testing.T) {
	pm := &fakePM{
		roles: map[string][]int{
			"has_read_perm_project":  []int{common.RoleGuest},
			"has_write_perm_project": []int{common.RoleGuest, common.RoleDeveloper},
		},
	}

	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasWritePerm("has_write_perm_project"))

	// authenticated, has read perm
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.False(t, ctx.HasWritePerm("has_read_perm_project")) // authenticated, has read perm

	// authenticated, has write perm
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.True(t, ctx.HasWritePerm("has_write_perm_project"))

	// authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "test",
		HasAdminRole: 1,
	}, pm)
	assert.True(t, ctx.HasReadPerm("has_no_perm_project"))
}

func TestHasAllPerm(t *testing.T) {
	pm := &fakePM{
		roles: map[string][]int{
			"has_read_perm_project":  []int{common.RoleGuest},
			"has_write_perm_project": []int{common.RoleGuest, common.RoleDeveloper},
			"has_all_perm_project":   []int{common.RoleGuest, common.RoleDeveloper, common.RoleProjectAdmin},
		},
	}

	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasAllPerm("has_all_perm_project"))

	// authenticated, has read perm
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.False(t, ctx.HasAllPerm("has_read_perm_project"))

	// authenticated, has write perm
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.False(t, ctx.HasAllPerm("has_write_perm_project"))

	// authenticated, has all perms
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.True(t, ctx.HasAllPerm("has_all_perm_project"))

	// authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "test",
		HasAdminRole: 1,
	}, pm)
	assert.True(t, ctx.HasReadPerm("has_no_perm_project"))
}
