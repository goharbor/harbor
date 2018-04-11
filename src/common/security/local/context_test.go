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

package local

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/dao/project"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/promgr"
	"github.com/vmware/harbor/src/ui/promgr/pmsdriver/local"
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
	dbHost := os.Getenv("MYSQL_HOST")
	if len(dbHost) == 0 {
		log.Fatalf("environment variable MYSQL_HOST is not set")
	}
	dbPortStr := os.Getenv("MYSQL_PORT")
	if len(dbPortStr) == 0 {
		log.Fatalf("environment variable MYSQL_PORT is not set")
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("invalid MYSQL_PORT: %v", err)
	}
	dbUser := os.Getenv("MYSQL_USR")
	if len(dbUser) == 0 {
		log.Fatalf("environment variable MYSQL_USR is not set")
	}

	dbPassword := os.Getenv("MYSQL_PWD")
	dbDatabase := os.Getenv("MYSQL_DATABASE")
	if len(dbDatabase) == 0 {
		log.Fatalf("environment variable MYSQL_DATABASE is not set")
	}

	database := &models.Database{
		Type: "mysql",
		MySQL: &models.MySQL{
			Host:     dbHost,
			Port:     dbPort,
			Username: dbUser,
			Password: dbPassword,
			Database: dbDatabase,
		},
	}

	log.Infof("MYSQL_HOST: %s, MYSQL_USR: %s, MYSQL_PORT: %d, MYSQL_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)

	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

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
		HasAdminRole: 1,
	}, nil)
	assert.True(t, ctx.IsSysAdmin())
}

func TestIsSolutionUser(t *testing.T) {
	ctx := NewSecurityContext(nil, nil)
	assert.False(t, ctx.IsSolutionUser())
}

func TestHasReadPerm(t *testing.T) {
	// public project
	ctx := NewSecurityContext(nil, pm)
	assert.True(t, ctx.HasReadPerm("library"))

	// private project, unauthenticated
	ctx = NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasReadPerm(private.Name))

	// private project, authenticated, has no perm
	ctx = NewSecurityContext(&models.User{
		Username: "test",
	}, pm)
	assert.False(t, ctx.HasReadPerm(private.Name))

	// private project, authenticated, has read perm
	ctx = NewSecurityContext(guestUser, pm)
	assert.True(t, ctx.HasReadPerm(private.Name))

	// private project, authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "admin",
		HasAdminRole: 1,
	}, pm)
	assert.True(t, ctx.HasReadPerm(private.Name))
}

func TestHasWritePerm(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasWritePerm(private.Name))

	// authenticated, has read perm
	ctx = NewSecurityContext(guestUser, pm)
	assert.False(t, ctx.HasWritePerm(private.Name))

	// authenticated, has write perm
	ctx = NewSecurityContext(developerUser, pm)
	assert.True(t, ctx.HasWritePerm(private.Name))

	// authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "admin",
		HasAdminRole: 1,
	}, pm)
	assert.True(t, ctx.HasReadPerm(private.Name))
}

func TestHasAllPerm(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, pm)
	assert.False(t, ctx.HasAllPerm(private.Name))

	// authenticated, has all perms
	ctx = NewSecurityContext(projectAdminUser, pm)
	assert.True(t, ctx.HasAllPerm(private.Name))

	// authenticated, system admin
	ctx = NewSecurityContext(&models.User{
		Username:     "admin",
		HasAdminRole: 1,
	}, pm)
	assert.True(t, ctx.HasAllPerm(private.Name))
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
