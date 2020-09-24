// Copyright 2018 Project Harbor Authors
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
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/auth"
)

// Auth implements Authenticator interface to authenticate user against DB.
type Auth struct {
	auth.DefaultAuthenticateHelper
}

// Authenticate calls dao to authenticate user.
func (d *Auth) Authenticate(m models.AuthModel) (*models.User, error) {
	u, err := dao.LoginByDb(m)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, auth.NewErrAuth("Invalid credentials")
	}
	return u, nil
}

// SearchUser - Check if user exist in local db
func (d *Auth) SearchUser(username string) (*models.User, error) {
	var queryCondition = models.User{
		Username: username,
	}

	return dao.GetUser(queryCondition)
}

// OnBoardUser -
func (d *Auth) OnBoardUser(u *models.User) error {
	return nil
}

func init() {
	auth.Register(common.DBAuth, &Auth{})
}
