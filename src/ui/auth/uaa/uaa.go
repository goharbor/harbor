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

package uaa

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/uaa"
	"github.com/vmware/harbor/src/ui/auth"
	"github.com/vmware/harbor/src/ui/config"
)

//CreateClient create a UAA Client instance based on system configuration.
func CreateClient() (uaa.Client, error) {
	UAASettings, err := config.UAASettings()
	if err != nil {
		return nil, err
	}
	cfg := &uaa.ClientConfig{
		ClientID:      UAASettings.ClientID,
		ClientSecret:  UAASettings.ClientSecret,
		Endpoint:      UAASettings.Endpoint,
		SkipTLSVerify: !UAASettings.VerifyCert,
		CARootPath:    os.Getenv("UAA_CA_ROOT"),
	}
	return uaa.NewDefaultClient(cfg)
}

// Auth is the implementation of AuthenticateHelper to access uaa for authentication.
type Auth struct {
	sync.Mutex
	client uaa.Client
}

//Authenticate ...
func (u *Auth) Authenticate(m models.AuthModel) (*models.User, error) {
	if err := u.ensureClient(); err != nil {
		return nil, err
	}
	t, err := u.client.PasswordAuth(m.Principal, m.Password)
	if t != nil && err == nil {
		//TODO: See if it's possible to get more information from token.
		user := &models.User{
			Username: m.Principal,
		}
		err = u.OnBoardUser(user)
		return user, err
	}
	return nil, err
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, return the user's profile.
func (u *Auth) OnBoardUser(user *models.User) error {
	user.Username = strings.TrimSpace(user.Username)
	if len(user.Username) == 0 {
		return fmt.Errorf("The Username is empty")
	}
	if len(user.Password) == 0 {
		user.Password = "1234567ab"
	}
	if len(user.Realname) == 0 {
		user.Realname = user.Username
	}
	if len(user.Email) == 0 {
		//TODO: handle the case when user.Username itself is an email address.
		user.Email = user.Username + "@uaa.placeholder"
	}
	user.Comment = "From UAA"
	return dao.OnBoardUser(user)
}

// SearchUser search user on uaa server, transform it to Harbor's user model
func (u *Auth) SearchUser(username string) (*models.User, error) {
	if err := u.ensureClient(); err != nil {
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
		return nil, fmt.Errorf("Multiple entries found for username: %s", username)
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

func (u *Auth) ensureClient() error {
	if u.client != nil {
		return nil
	}
	u.Lock()
	defer u.Unlock()
	if u.client == nil {
		c, err := CreateClient()
		if err != nil {
			return err
		}
		u.client = c
	}
	return nil
}
func init() {
	auth.Register(common.UAAAuth, &Auth{})
}
