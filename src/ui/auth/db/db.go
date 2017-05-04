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
	"context"
	"fmt"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"
)

// Name of database authenticator
const Name = "database"

// TODO remove the annotation
/*
func init() {
	auth.Register(Name, &databaseAuthenticatorFactory{})
}
*/
type database struct{}

// Authenticate users with the records in database and add user information
// into the context
// authInfo should contain princial and credential
// TODO add test case
func (d *database) Authenticate(ctx context.Context,
	authInfo ...string) (context.Context, error) {

	if ctx == nil {
		ctx = context.Background()
	}

	if len(authInfo) != 2 {
		return ctx, fmt.Errorf("invalid authInfo")
	}

	principal := authInfo[0]
	if len(principal) == 0 {
		return ctx, fmt.Errorf("invalid principal")
	}

	credential := authInfo[1]

	// TODO refactor AuthModel
	m := models.AuthModel{
		Principal: principal,
		Password:  credential,
	}

	user, err := dao.LoginByDb(m)
	if err != nil {
		return ctx, err
	}

	if user == nil {
		return ctx, fmt.Errorf("invalid principal or credential")
	}

	ctx = context.WithValue(ctx, common.CtxKeyUser, *user)
	log.Infof("user %s authenticated by database has been added into context", user.Username)
	return ctx, nil
}

type databaseAuthenticatorFactory struct{}

func (d *databaseAuthenticatorFactory) Create(parameter interface{}) (auth.Authenticator, error) {
	return &database{}, nil
}

// Auth implements Authenticator interface to authenticate user against DB.
type Auth struct{}

// Authenticate calls dao to authenticate user.
func (d *Auth) Authenticate(m models.AuthModel) (*models.User, error) {
	u, err := dao.LoginByDb(m)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func init() {
	auth.RegisterOld("db_auth", &Auth{})
}
