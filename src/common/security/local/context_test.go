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
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/group"
	"github.com/goharbor/harbor/src/common/dao/project"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/core/promgr/pmsdriver/local"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	private = &models.Project{
		Name:    "private_project",
		OwnerID: 1,
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

	pm = promgr.NewDefaultProjectManager(local.NewDriver(), true)
)

func TestMain(m *testing.M) {

	test.InitDatabaseFromEnv()

	// regiser users
	id, err := dao.Register(*projectAdminUser)
	if err != nil {
		log.Fatalf("failed to register user: %v", err)
	}
	projectAdminUser.UserID = int(id)
	defer dao.DeleteUser(int(id))

	id, err = dao.Register(*developerUser)
	if err != nil {
		log.Fatalf("failed to register user: %v", err)
	}
	developerUser.UserID = int(id)
	defer dao.DeleteUser(int(id))

	id, err = dao.Register(*guestUser)
	if err != nil {
		log.Fatalf("failed to register user: %v", err)
	}
	guestUser.UserID = int(id)
	defer dao.DeleteUser(int(id))

	// add project
	id, err = dao.AddProject(*private)
	if err != nil {
		log.Fatalf("failed to add project: %v", err)
	}
	private.ProjectID = id
	defer dao.DeleteProject(id)

	var projectAdminPMID, developerUserPMID, guestUserPMID int
	// add project members
	projectAdminPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  private.ProjectID,
		EntityID:   projectAdminUser.UserID,
		EntityType: common.UserMember,
		Role:       common.RoleProjectAdmin,
	})
	if err != nil {
		log.Fatalf("failed to add member: %v", err)
	}
	defer project.DeleteProjectMemberByID(projectAdminPMID)

	developerUserPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  private.ProjectID,
		EntityID:   developerUser.UserID,
		EntityType: common.UserMember,
		Role:       common.RoleDeveloper,
	})
	if err != nil {
		log.Fatalf("failed to add member: %v", err)
	}
	defer project.DeleteProjectMemberByID(developerUserPMID)
	guestUserPMID, err = project.AddProjectMember(models.Member{
		ProjectID:  private.ProjectID,
		EntityID:   guestUser.UserID,
		EntityType: common.UserMember,
		Role:       common.RoleGuest,
	})
	if err != nil {
		log.Fatalf("failed to add member: %v", err)
	}
	defer project.DeleteProjectMemberByID(guestUserPMID)
	os.Exit(m.Run())
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
		SysAdminFlag: true,
	}, nil)
	assert.True(t, ctx.IsSysAdmin())
}

func TestIsSolutionUser(t *testing.T) {
	ctx := NewSecurityContext(nil, nil)
	assert.False(t, ctx.IsSolutionUser())
}

func TestHasPullPerm(t *testing.T) {
	// public project
	ctx := NewSecurityContext(nil, pm)

	resource := rbac.NewProjectNamespace(1).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(rbac.ActionPull, resource))

	// private project, unauthenticated
	ctx = NewSecurityContext(nil, pm)
	resource = rbac.NewProjectNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.False(t, ctx.Can(rbac.ActionPull, resource))

	// private project, authenticated, has no perm
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.False(t, ctx.Can(rbac.ActionPull, resource))

	// private project, authenticated, has read perm
	ctx = NewSecurityContext(guestUser, pm)
	assert.True(t, ctx.Can(rbac.ActionPull, resource))

	// private project, authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "admin",
		SysAdminFlag: true,
	}, pm)
	assert.True(t, ctx.Can(rbac.ActionPull, resource))
}

func TestHasPushPerm(t *testing.T) {
	resource := rbac.NewProjectNamespace(private.ProjectID).Resource(rbac.ResourceRepository)

	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	assert.False(t, ctx.Can(rbac.ActionPush, resource))

	// authenticated, has read perm
	ctx = NewSecurityContext(guestUser, pm)
	assert.False(t, ctx.Can(rbac.ActionPush, resource))

	// authenticated, has write perm
	ctx = NewSecurityContext(developerUser, pm)
	assert.True(t, ctx.Can(rbac.ActionPush, resource))

	// authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "admin",
		SysAdminFlag: true,
	}, pm)
	assert.True(t, ctx.Can(rbac.ActionPush, resource))
}

func TestHasPushPullPerm(t *testing.T) {
	resource := rbac.NewProjectNamespace(private.ProjectID).Resource(rbac.ResourceRepository)

	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	assert.False(t, ctx.Can(rbac.ActionPush, resource) && ctx.Can(rbac.ActionPull, resource))

	// authenticated, has all perms
	ctx = NewSecurityContext(projectAdminUser, pm)
	assert.True(t, ctx.Can(rbac.ActionPush, resource) && ctx.Can(rbac.ActionPull, resource))

	// authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "admin",
		SysAdminFlag: true,
	}, pm)
	assert.True(t, ctx.Can(rbac.ActionPush, resource) && ctx.Can(rbac.ActionPull, resource))
}

func TestHasPushPullPermWithGroup(t *testing.T) {
	PrepareGroupTest()
	project, err := dao.GetProjectByName("group_project")
	if err != nil {
		t.Errorf("Error occurred when GetProjectByName: %v", err)
	}
	developer, err := dao.GetUser(models.User{Username: "sample01"})
	if err != nil {
		t.Errorf("Error occurred when GetUser: %v", err)
	}

	userGroups, err := group.QueryUserGroup(models.UserGroup{GroupType: common.LDAPGroupType, LdapGroupDN: "cn=harbor_user,dc=example,dc=com"})
	if err != nil {
		t.Errorf("Failed to query user group %v", err)
	}
	if len(userGroups) < 1 {
		t.Errorf("Failed to retrieve user group")
	}

	developer.GroupIDs = []int{userGroups[0].ID}

	resource := rbac.NewProjectNamespace(project.ProjectID).Resource(rbac.ResourceRepository)

	ctx := NewSecurityContext(developer, pm)
	assert.True(t, ctx.Can(rbac.ActionPush, resource))
	assert.True(t, ctx.Can(rbac.ActionPull, resource))
}

func TestGetMyProjects(t *testing.T) {
	ctx := NewSecurityContext(guestUser, pm)
	projects, err := ctx.GetMyProjects()
	require.Nil(t, err)
	assert.Equal(t, 1, len(projects))
	assert.Equal(t, private.ProjectID, projects[0].ProjectID)
}

func TestGetProjectRoles(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	roles := ctx.GetProjectRoles(private.Name)
	assert.Equal(t, 0, len(roles))

	// authenticated, project name of ID is nil
	ctx = NewSecurityContext(guestUser, pm)
	roles = ctx.GetProjectRoles(nil)
	assert.Equal(t, 0, len(roles))

	// authenticated, has read perm
	ctx = NewSecurityContext(guestUser, pm)
	roles = ctx.GetProjectRoles(private.Name)
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, common.RoleGuest, roles[0])

	// authenticated, has write perm
	ctx = NewSecurityContext(developerUser, pm)
	roles = ctx.GetProjectRoles(private.Name)
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, common.RoleDeveloper, roles[0])

	// authenticated, has all perms
	ctx = NewSecurityContext(projectAdminUser, pm)
	roles = ctx.GetProjectRoles(private.Name)
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, common.RoleProjectAdmin, roles[0])
}
func PrepareGroupTest() {
	initSqls := []string{
		`insert into user_group (group_name, group_type, ldap_group_dn) values ('harbor_group_01', 1, 'cn=harbor_user,dc=example,dc=com')`,
		`insert into harbor_user (username, email, password, realname) values ('sample01', 'sample01@example.com', 'harbor12345', 'sample01')`,
		`insert into project (name, owner_id) values ('group_project', 1)`,
		`insert into project (name, owner_id) values ('group_project_private', 1)`,
		`insert into project_metadata (project_id, name, value) values ((select project_id from project where name = 'group_project'), 'public', 'false')`,
		`insert into project_metadata (project_id, name, value) values ((select project_id from project where name = 'group_project_private'), 'public', 'false')`,
		`insert into project_member (project_id, entity_id, entity_type, role) values ((select project_id from project where name = 'group_project'), (select id from user_group where group_name = 'harbor_group_01'),'g', 2)`,
	}

	clearSqls := []string{
		`delete from project_metadata where project_id in (select project_id from project where name in ('group_project', 'group_project_private'))`,
		`delete from project where name in ('group_project', 'group_project_private')`,
		`delete from project_member where project_id in (select project_id from project where name in ('group_project', 'group_project_private'))`,
		`delete from user_group where group_name = 'harbor_group_01'`,
		`delete from harbor_user where username = 'sample01'`,
	}
	dao.PrepareTestData(clearSqls, initSqls)
}

func TestSecurityContext_GetRolesByGroup(t *testing.T) {
	PrepareGroupTest()
	project, err := dao.GetProjectByName("group_project")
	if err != nil {
		t.Errorf("Error occurred when GetProjectByName: %v", err)
	}
	developer, err := dao.GetUser(models.User{Username: "sample01"})
	if err != nil {
		t.Errorf("Error occurred when GetUser: %v", err)
	}
	userGroups, err := group.QueryUserGroup(models.UserGroup{GroupType: common.LDAPGroupType, LdapGroupDN: "cn=harbor_user,dc=example,dc=com"})
	if err != nil {
		t.Errorf("Failed to query user group %v", err)
	}
	if len(userGroups) < 1 {
		t.Errorf("Failed to retrieve user group")
	}

	developer.GroupIDs = []int{userGroups[0].ID}
	type fields struct {
		user *models.User
		pm   promgr.ProjectManager
	}
	type args struct {
		projectIDOrName interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []int
	}{
		{"Developer", fields{user: developer, pm: pm}, args{project.ProjectID}, []int{2}},
		{"Guest", fields{user: guestUser, pm: pm}, args{project.ProjectID}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SecurityContext{
				user: tt.fields.user,
				pm:   tt.fields.pm,
			}
			if got := s.GetRolesByGroup(tt.args.projectIDOrName); !dao.ArrayEqual(got, tt.want) {
				t.Errorf("SecurityContext.GetRolesByGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecurityContext_GetMyProjects(t *testing.T) {
	type fields struct {
		user *models.User
		pm   promgr.ProjectManager
	}
	tests := []struct {
		name     string
		fields   fields
		wantSize int
		wantErr  bool
	}{
		{"Admin", fields{user: projectAdminUser, pm: pm}, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SecurityContext{
				user: tt.fields.user,
				pm:   tt.fields.pm,
			}
			got, err := s.GetMyProjects()
			if (err != nil) != tt.wantErr {
				t.Errorf("SecurityContext.GetMyProjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantSize {
				t.Errorf("SecurityContext.GetMyProjects() = %v, want %v", len(got), tt.wantSize)
			}
		})
	}
}

func Test_mergeRoles(t *testing.T) {
	type args struct {
		rolesA []int
		rolesB []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"normal", args{[]int{3, 4}, []int{1, 2, 3, 4}}, []int{1, 2, 3, 4}},
		{"empty", args{[]int{}, []int{}}, []int{}},
		{"left empty", args{[]int{}, []int{1, 2, 3, 4}}, []int{1, 2, 3, 4}},
		{"right empty", args{[]int{1, 2, 3, 4}, []int{}}, []int{1, 2, 3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeRoles(tt.args.rolesA, tt.args.rolesB); !test.CheckSetsEqual(got, tt.want) {
				t.Errorf("mergeRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}
