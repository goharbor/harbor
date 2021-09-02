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
	"context"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/user"
)

// Auth implements Authenticator interface to authenticate user against DB.
type Auth struct {
	auth.DefaultAuthenticateHelper
	userMgr user.Manager
}

// Authenticate calls dao to authenticate user.
func (d *Auth) Authenticate(ctx context.Context, m models.AuthModel) (*models.User, error) {
	u, err := d.userMgr.MatchLocalPassword(ctx, m.Principal, m.Password)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, auth.NewErrAuth("Invalid credentials")
	}
	return u, nil
}

// SearchUser - Check if user exist in local db
func (d *Auth) SearchUser(ctx context.Context, username string) (*models.User, error) {
	u, err := d.userMgr.GetByName(ctx, username)
	if errors.IsNotFoundErr(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return u, err
}

// OnBoardUser -
func (d *Auth) OnBoardUser(ctx context.Context, u *models.User) error {
	return nil
}

func init() {
	auth.Register(common.DBAuth, &Auth{
		userMgr: user.New(),
	})
}
