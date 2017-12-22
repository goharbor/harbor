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
	"sync"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/uaa"
	"github.com/vmware/harbor/src/ui/auth"
	"github.com/vmware/harbor/src/ui/config"
)

var lock = &sync.Mutex{}
var client uaa.Client

//GetClient returns the client instance, if the client is not created it creates one.
func GetClient() (uaa.Client, error) {
	lock.Lock()
	defer lock.Unlock()
	if client != nil {
		return client, nil
	}
	UAASettings, err := config.UAASettings()
	if err != nil {
		return nil, err
	}
	cfg := &uaa.ClientConfig{
		ClientID:      UAASettings.ClientID,
		ClientSecret:  UAASettings.ClientSecret,
		Endpoint:      UAASettings.Endpoint,
		SkipTLSVerify: !UAASettings.VerifyCert,
	}
	client, err = uaa.NewDefaultClient(cfg)
	return client, err
}

func doAuth(username, password string, client uaa.Client) (*models.User, error) {
	t, err := client.PasswordAuth(username, password)
	if t != nil && err == nil {
		//TODO: See if it's possible to get more information from token.
		u := &models.User{
			Username: username,
			Password: "1234567ab",
			Email:    username + "@placeholder.com",
			Realname: username,
		}
		return u, nil
	}
	return nil, err
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
			Password: "1234567ab",
			Email:    m.Principal + "@uaa.placeholder",
			Realname: m.Principal,
		}
		err = u.OnBoardUser(user)
		return user, err
	}
	return nil, err
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, return the user's profile.
func (u *Auth) OnBoardUser(user *models.User) error {
	return dao.OnBoardUser(user)
}

// // SearchUser -  search user on uaa server
func (u *Auth) SearchUser(username string) (*models.User, error) {
	if err := u.ensureClient(); err != nil {
		return nil, err
	}
	l, err := u.client.SearchUser(username)
	if err != nil {

	}
	if len(l) == 0 {
		return nil, fmt.Errorf("No entry found for username: %s", username)
	}
	if len(l) > 1 {
		return nil, fmt.Errorf("Multiple entries found for username: %s", username)
	}
	e := l[0]
	email := username + "@uaa.placeholder"
	if len(e.Emails) > 0 {
		email = e.Emails[0].Value
	}
	return &models.User{
		Username: username,
		Password: "1234567ab",
		Email:    email,
		Realname: username,
	}, nil

}

func (u *Auth) ensureClient() error {
	if u.client != nil {
		return nil
	}
	u.Lock()
	defer u.Unlock()
	if u.client == nil {
		c, err := GetClient()
		if err != nil {
			return err
		}
		u.client = c
	}
	return nil
}
func init() {
	auth.Register(common.UAAAuth, &Auth{
		client: nil,
	})
}
