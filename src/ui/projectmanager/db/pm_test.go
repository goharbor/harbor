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

func TestIsPublic(t *testing.T) {
	pms := &ProjectManager{}
	// project name
	assert.True(t, pms.IsPublic("library"))
	// project ID
	assert.True(t, pms.IsPublic(int64(1)))
	// non exist project
	assert.False(t, pms.IsPublic("non_exist_project"))
	// invalid type
	assert.False(t, pms.IsPublic(1))
}

func TestGetRoles(t *testing.T) {
	pm := &ProjectManager{}

	// non exist user
	assert.Equal(t, []int{},
		pm.GetRoles("non_exist_user", int64(1)))

	// project ID
	assert.Equal(t, []int{common.RoleProjectAdmin},
		pm.GetRoles("admin", int64(1)))

	// project name
	assert.Equal(t, []int{common.RoleProjectAdmin},
		pm.GetRoles("admin", "library"))

	// non exist project
	assert.Equal(t, []int{},
		pm.GetRoles("admin", "non_exist_project"))

	// invalid type
	assert.Equal(t, []int{}, pm.GetRoles("admin", 1))
}
