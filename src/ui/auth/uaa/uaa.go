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
	"sync"

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
		ClientID:     UAASettings.ClientID,
		ClientSecret: UAASettings.ClientSecret,
		Endpoint:     UAASettings.Endpoint,
		//TODO: remove it
		SkipTLSVerify: true,
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
		err = dao.OnBoardUser(u)
		if err == nil {
			return u, nil
		}
	}
	return nil, err
}

// Auth is the implementation of Authenticator to access uaa for authentication.
type Auth struct{}

//Authenticate ...
func (u *Auth) Authenticate(m models.AuthModel) (*models.User, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}
	return doAuth(m.Principal, m.Password, client)
}

func init() {
	auth.Register(auth.UAAAuth, &Auth{})
}
