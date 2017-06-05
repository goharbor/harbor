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

package db

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
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

	os.Exit(m.Run())
}

func TestGet(t *testing.T) {
	pm := &ProjectManager{}

	// project name
	project, err := pm.Get("library")
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, "library", project.Name)

	// project ID
	project, err = pm.Get(int64(1))
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, int64(1), project.ProjectID)

	// non-exist project
	project, err = pm.Get("non-exist-project")
	assert.Nil(t, err)
	assert.Nil(t, project)

	// invalid type
	project, err = pm.Get(true)
	assert.NotNil(t, err)
}

func TestExist(t *testing.T) {
	pm := &ProjectManager{}

	// exist project
	exist, err := pm.Exist("library")
	assert.Nil(t, err)
	assert.True(t, exist)

	// non-exist project
	exist, err = pm.Exist("non-exist-project")
	assert.Nil(t, err)
	assert.False(t, exist)
}

func TestIsPublic(t *testing.T) {
	pms := &ProjectManager{}
	// public project
	public, err := pms.IsPublic("library")
	assert.Nil(t, err)
	assert.True(t, public)
	// non exist project
	public, err = pms.IsPublic("non_exist_project")
	assert.Nil(t, err)
	assert.False(t, public)
}

func TestGetRoles(t *testing.T) {
	pm := &ProjectManager{}

	// non exist user
	roles, err := pm.GetRoles("non_exist_user", int64(1))
	assert.Nil(t, err)
	assert.Equal(t, []int{}, roles)

	// exist project
	roles, err = pm.GetRoles("admin", "library")
	assert.Nil(t, err)
	assert.Equal(t, []int{common.RoleProjectAdmin}, roles)

	// non-exist project
	roles, err = pm.GetRoles("admin", "non_exist_project")
	assert.Nil(t, err)
	assert.Equal(t, []int{}, roles)
}

func TestGetPublic(t *testing.T) {
	pm := &ProjectManager{}
	projects, err := pm.GetPublic()
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(projects))

	for _, project := range projects {
		assert.Equal(t, 1, project.Public)
	}
}

func TestGetByMember(t *testing.T) {
	pm := &ProjectManager{}
	projects, err := pm.GetByMember("admin")
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(projects))
}

func TestCreateAndDelete(t *testing.T) {
	pm := &ProjectManager{}

	// nil project
	_, err := pm.Create(nil)
	assert.NotNil(t, err)

	// nil project name
	_, err = pm.Create(&models.Project{
		OwnerID: 1,
	})
	assert.NotNil(t, err)

	// nil owner id and nil owner name
	_, err = pm.Create(&models.Project{
		Name:      "test",
		OwnerName: "non_exist_user",
	})
	assert.NotNil(t, err)

	// valid project, owner id
	id, err := pm.Create(&models.Project{
		Name:    "test",
		OwnerID: 1,
	})
	assert.Nil(t, err)
	assert.Nil(t, pm.Delete(id))

	// valid project, owner name
	id, err = pm.Create(&models.Project{
		Name:      "test",
		OwnerName: "admin",
	})
	assert.Nil(t, err)
	assert.Nil(t, pm.Delete(id))
}

func TestUpdate(t *testing.T) {
	pm := &ProjectManager{}

	id, err := pm.Create(&models.Project{
		Name:    "test",
		OwnerID: 1,
	})
	assert.Nil(t, err)
	defer pm.Delete(id)

	project, err := pm.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, 0, project.Public)

	project.Public = 1
	assert.Nil(t, pm.Update(id, project))

	project, err = pm.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, 1, project.Public)
}

func TestGetTotal(t *testing.T) {
	pm := &ProjectManager{}

	id, err := pm.Create(&models.Project{
		Name:    "get_total_test",
		OwnerID: 1,
		Public:  1,
	})
	assert.Nil(t, err)
	defer pm.Delete(id)

	// get by name
	total, err := pm.GetTotal(&models.ProjectQueryParam{
		Name: "get_total_test",
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), total)

	// get by owner
	total, err = pm.GetTotal(&models.ProjectQueryParam{
		Owner: "admin",
	})
	assert.Nil(t, err)
	assert.NotEqual(t, 0, total)

	// get by public
	value := true
	total, err = pm.GetTotal(&models.ProjectQueryParam{
		Public: &value,
	})
	assert.Nil(t, err)
	assert.NotEqual(t, 0, total)
}

func TestGetAll(t *testing.T) {
	pm := &ProjectManager{}

	id, err := pm.Create(&models.Project{
		Name:    "get_all_test",
		OwnerID: 1,
		Public:  1,
	})
	assert.Nil(t, err)
	defer pm.Delete(id)

	// get by name
	projects, err := pm.GetAll(&models.ProjectQueryParam{
		Name: "get_all_test",
	})
	assert.Nil(t, err)
	assert.Equal(t, id, projects[0].ProjectID)

	// get by owner
	projects, err = pm.GetAll(&models.ProjectQueryParam{
		Owner: "admin",
	})
	assert.Nil(t, err)
	exist := false
	for _, project := range projects {
		if project.ProjectID == id {
			exist = true
			break
		}
	}
	assert.True(t, exist)

	// get by public
	value := true
	projects, err = pm.GetAll(&models.ProjectQueryParam{
		Public: &value,
	})
	assert.Nil(t, err)
	exist = false
	for _, project := range projects {
		if project.ProjectID == id {
			exist = true
			break
		}
	}
	assert.True(t, exist)
}

func TestGetHasReadPerm(t *testing.T) {
	pm := &ProjectManager{}

	// do not pass username
	projects, err := pm.GetHasReadPerm()
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(projects))
	exist := false
	for _, project := range projects {
		if project.ProjectID == 1 {
			exist = true
			break
		}
	}
	assert.True(t, exist)

	// username is nil
	projects, err = pm.GetHasReadPerm("")
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(projects))
	exist = false
	for _, project := range projects {
		if project.ProjectID == 1 {
			exist = true
			break
		}
	}
	assert.True(t, exist)

	// valid username
	id, err := pm.Create(&models.Project{
		Name:    "get_has_read_perm_test",
		OwnerID: 1,
		Public:  0,
	})
	assert.Nil(t, err)
	defer pm.Delete(id)

	projects, err = pm.GetHasReadPerm("admin")
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(projects))
	exist1 := false
	exist2 := false
	for _, project := range projects {
		if project.ProjectID == 1 {
			exist1 = true
		}
		if project.ProjectID == id {
			exist2 = true
		}
	}
	assert.True(t, exist1)
	assert.True(t, exist2)
}
