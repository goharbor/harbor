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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/models"
)

var (
	public = &models.Project{
		Name:   "public_project",
		Public: 1,
	}

	private = &models.Project{
		Name:   "private_project",
		Public: 0,
	}

	read = &models.Project{
		Name: "has_read_perm_project",
	}

	write = &models.Project{
		Name: "has_write_perm_project",
	}

	all = &models.Project{
		Name: "has_all_perm_project",
	}
)

type fakePM struct {
	projects []*models.Project
	roles    map[string][]int
}

func (f *fakePM) IsPublic(projectIDOrName interface{}) bool {
	for _, project := range f.projects {
		if project.Name == projectIDOrName.(string) {
			return project.Public == 1
		}
	}
	return false
}
func (f *fakePM) GetRoles(username string, projectIDOrName interface{}) []int {
	return f.roles[projectIDOrName.(string)]
}
func (f *fakePM) Get(projectIDOrName interface{}) *models.Project {
	for _, project := range f.projects {
		if project.Name == projectIDOrName.(string) {
			return project
		}
	}
	return nil
}
func (f *fakePM) Exist(projectIDOrName interface{}) bool {
	for _, project := range f.projects {
		if project.Name == projectIDOrName.(string) {
			return true
		}
	}
	return false
}

// nil implement
func (f *fakePM) GetPublic() []*models.Project {
	return []*models.Project{}
}

// nil implement
func (f *fakePM) GetByMember(username string) []*models.Project {
	return []*models.Project{}
}

// nil implement
func (f *fakePM) Create(*models.Project) (int64, error) {
	return 0, fmt.Errorf("not support")
}

// nil implement
func (f *fakePM) Delete(projectIDOrName interface{}) error {
	return fmt.Errorf("not support")
}

// nil implement
func (f *fakePM) Update(projectIDOrName interface{}, project *models.Project) error {
	return fmt.Errorf("not support")
}

// nil implement
func (f *fakePM) GetAll(owner, name, public, member string, role int, page,
	size int64) ([]*models.Project, int64) {
	return []*models.Project{}, 0
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
		projects: []*models.Project{public, private, read},
		roles: map[string][]int{
			"has_read_perm_project": []int{common.RoleGuest},
		},
	}

	// non-exist project
	ctx := NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasReadPerm("non_exist_project"))

	// public project
	ctx = NewSecurityContext(nil, pm)
	assert.True(t, ctx.HasReadPerm("public_project"))

	// private project, unauthenticated
	ctx = NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasReadPerm("private_project"))

	// private project, authenticated, has no perm
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.False(t, ctx.HasReadPerm("private_project"))

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
	assert.True(t, ctx.HasReadPerm("private_project"))

	// non-exist project, authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "test",
		HasAdminRole: 1,
	}, pm)
	assert.False(t, ctx.HasReadPerm("non_exist_project"))
}

func TestHasWritePerm(t *testing.T) {
	pm := &fakePM{
		projects: []*models.Project{read, write, private},
		roles: map[string][]int{
			"has_read_perm_project":  []int{common.RoleGuest},
			"has_write_perm_project": []int{common.RoleGuest, common.RoleDeveloper},
		},
	}

	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasWritePerm("has_write_perm_project"))

	// authenticated, non-exist project
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.False(t, ctx.HasWritePerm("non_exist_project"))

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
	assert.True(t, ctx.HasReadPerm("private_project"))

	// authenticated, system admin, non-exist project
	ctx = NewSecurityContext(&models.User{
		Username:     "test",
		HasAdminRole: 1,
	}, pm)
	assert.False(t, ctx.HasReadPerm("non_exist_project"))
}

func TestHasAllPerm(t *testing.T) {
	pm := &fakePM{
		projects: []*models.Project{read, write, all, private},
		roles: map[string][]int{
			"has_read_perm_project":  []int{common.RoleGuest},
			"has_write_perm_project": []int{common.RoleGuest, common.RoleDeveloper},
			"has_all_perm_project":   []int{common.RoleGuest, common.RoleDeveloper, common.RoleProjectAdmin},
		},
	}

	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasAllPerm("has_all_perm_project"))

	// authenticated, non-exist project
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.False(t, ctx.HasAllPerm("non_exist_project"))

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
	assert.True(t, ctx.HasAllPerm("private_project"))

	// authenticated, system admin, non-exist project
	ctx = NewSecurityContext(&models.User{
		Username:     "test",
		HasAdminRole: 1,
	}, pm)
	assert.False(t, ctx.HasAllPerm("non_exist_project"))
}
