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
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
)

func TestCreate(t *testing.T) {
	factory := &databaseAuthenticatorFactory{}
	_, err := factory.Create(nil)
	assert.Nil(t, err)
}

func TestAuthenticate(t *testing.T) {
	if err := initDatabase(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	authenticator := &database{}
	principal := ""
	credential := ""

	// only provide principal
	_, err := authenticator.Authenticate(nil, principal)
	assert.NotNil(t, err)

	// invalid principal
	_, err = authenticator.Authenticate(nil, principal, credential)
	assert.NotNil(t, err)

	// wrong credential
	principal = "admin"
	credential = "wrong_password"
	_, err = authenticator.Authenticate(nil, principal, credential)
	assert.NotNil(t, err)

	// valid principal and credential
	principal = "admin"
	credential = "Harbor12345"
	ctx, err := authenticator.Authenticate(nil, principal, credential)
	assert.Nil(t, err)
	u := ctx.Value(common.CtxKeyUser)
	user, ok := u.(models.User)
	if !ok {
		t.Fatalf("user got from context is not User type")
	}
	assert.Equal(t, "admin", user.Username)
}

func initDatabase() error {
	dbHost := os.Getenv("MYSQL_HOST")
	if len(dbHost) == 0 {
		return fmt.Errorf("environment variable MYSQL_HOST is not set")
	}
	dbUser := os.Getenv("MYSQL_USR")
	if len(dbUser) == 0 {
		return fmt.Errorf("environment variable MYSQL_USR is not set")
	}
	dbPortStr := os.Getenv("MYSQL_PORT")
	if len(dbPortStr) == 0 {
		return fmt.Errorf("environment variable MYSQL_PORT is not set")
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		return fmt.Errorf("invalid MYSQL_PORT: %v", err)
	}

	dbPassword := os.Getenv("MYSQL_PWD")
	dbDatabase := os.Getenv("MYSQL_DATABASE")
	if len(dbDatabase) == 0 {
		return fmt.Errorf("environment variable MYSQL_DATABASE is not set")
	}

	db := &models.Database{
		Type: "mysql",
		MySQL: &models.MySQL{
			Host:     dbHost,
			Port:     dbPort,
			Username: dbUser,
			Password: dbPassword,
			Database: dbDatabase,
		},
	}

	return dao.InitDatabase(db)
}
