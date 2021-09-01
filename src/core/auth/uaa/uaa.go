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

package uaa

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/uaa"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	userpkg "github.com/goharbor/harbor/src/pkg/user"
)

// Auth is the implementation of AuthenticateHelper to access uaa for authentication.
type Auth struct {
	sync.Mutex
	client uaa.Client
	auth.DefaultAuthenticateHelper
	userMgr userpkg.Manager
}

// Authenticate ...
func (u *Auth) Authenticate(ctx context.Context, m models.AuthModel) (*models.User, error) {
	if err := u.ensureClient(ctx); err != nil {
		return nil, err
	}
	t, err := u.client.PasswordAuth(m.Principal, m.Password)
	if err != nil {
		return nil, auth.NewErrAuth(err.Error())
	}
	user := &models.User{
		Username: m.Principal,
	}
	info, err2 := u.client.GetUserInfo(t.AccessToken)
	if err2 != nil {
		log.Warningf("Failed to extract user info from UAA, error: %v", err2)
	} else {
		user.Email = info.Email
		user.Realname = info.Name
	}
	return user, nil
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, return the user's profile.
func (u *Auth) OnBoardUser(ctx context.Context, user *models.User) error {
	user.Username = strings.TrimSpace(user.Username)
	if len(user.Username) == 0 {
		return fmt.Errorf("the Username is empty")
	}
	if len(user.Password) == 0 {
		user.Password = "1234567ab"
	}
	fillEmailRealName(user)
	user.Comment = "From UAA"
	return u.userMgr.Onboard(ctx, user)
}

func fillEmailRealName(user *models.User) {
	if len(user.Realname) == 0 {
		user.Realname = user.Username
	}
	if len(user.Email) == 0 && strings.Contains(user.Username, "@") {
		user.Email = user.Username
	}
}

// PostAuthenticate will check if user exists in DB, if not on Board user, if he does, update the profile.
func (u *Auth) PostAuthenticate(ctx context.Context, user *models.User) error {
	dbUser, err := u.userMgr.GetByName(ctx, user.Username)
	if errors.IsNotFoundErr(err) {
		return u.OnBoardUser(ctx, user)
	} else if err != nil {
		return err
	}
	user.UserID = dbUser.UserID
	user.SysAdminFlag = dbUser.SysAdminFlag
	fillEmailRealName(user)
	if err2 := u.userMgr.UpdateProfile(ctx, user, "Email", "Realname"); err2 != nil {
		log.Warningf("Failed to update user profile, user: %s, error: %v", user.Username, err2)
	}
	return nil
}

// SearchUser search user on uaa server, transform it to Harbor's user model
func (u *Auth) SearchUser(ctx context.Context, username string) (*models.User, error) {
	if err := u.ensureClient(ctx); err != nil {
		return nil, err
	}
	l, err := u.client.SearchUser(username)
	if err != nil {
		return nil, err
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) > 1 {
		return nil, fmt.Errorf("multiple entries found for username: %s", username)
	}
	e := l[0]
	email := ""
	if len(e.Emails) > 0 {
		email = e.Emails[0].Value
	}
	return &models.User{
		Username: username,
		Email:    email,
	}, nil
}

func (u *Auth) ensureClient(ctx context.Context) error {
	var cfg *uaa.ClientConfig
	UAASettings, err := config.UAASettings(ctx)
	//	log.Debugf("Uaa settings: %+v", UAASettings)
	if err != nil {
		log.Warningf("Failed to get UAA setting from Admin Server, error: %v", err)
	} else {
		cfg = &uaa.ClientConfig{
			ClientID:      UAASettings.ClientID,
			ClientSecret:  UAASettings.ClientSecret,
			Endpoint:      UAASettings.Endpoint,
			SkipTLSVerify: !UAASettings.VerifyCert,
			CARootPath:    os.Getenv("UAA_CA_ROOT"),
		}
	}
	if u.client != nil && cfg != nil {
		return u.client.UpdateConfig(cfg)
	}
	u.Lock()
	defer u.Unlock()
	if u.client == nil {
		c, err := uaa.NewDefaultClient(cfg)
		if err != nil {
			return err
		}
		u.client = c
	}
	return nil
}
func init() {
	auth.Register(common.UAAAuth, &Auth{
		userMgr: userpkg.New(),
	})
}
