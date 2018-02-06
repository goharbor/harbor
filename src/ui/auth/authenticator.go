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

package auth

import (
	"fmt"
	"time"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

// 1.5 seconds
const frozenTime time.Duration = 1500 * time.Millisecond

var lock = NewUserLock(frozenTime)

//ErrAuth is the type of error to indicate a failed authentication due to user's error.
type ErrAuth struct {
	details string
}

//Error ...
func (ea ErrAuth) Error() string {
	return fmt.Sprintf("Failed to authenticate user, due to error '%s'", ea.details)
}

//NewErrAuth ...
func NewErrAuth(msg string) ErrAuth {
	return ErrAuth{details: msg}
}

// AuthenticateHelper provides interface for user management in different auth modes.
type AuthenticateHelper interface {

	// Authenticate authenticate the user based on data in m.  Only when the error returned is an instance
	// of ErrAuth, it will be considered a bad credentials, other errors will be treated as server side error.
	Authenticate(m models.AuthModel) (*models.User, error)
	// OnBoardUser will check if a user exists in user table, if not insert the user and
	// put the id in the pointer of user model, if it does exist, fill in the user model based
	// on the data record of the user
	OnBoardUser(u *models.User) error
	// Get user information from account repository
	SearchUser(username string) (*models.User, error)
	// Update user information after authenticate, such as Onboard or sync info etc
	PostAuthenticate(u *models.User) error
}

// DefaultAuthenticateHelper - default AuthenticateHelper implementation
type DefaultAuthenticateHelper struct {
}

// Authenticate ...
func (d *DefaultAuthenticateHelper) Authenticate(m models.AuthModel) (*models.User, error) {
	return nil, nil
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, fill in the user model based
// on the data record of the user
func (d *DefaultAuthenticateHelper) OnBoardUser(u *models.User) error {
	return nil
}

//SearchUser - Get user information from account repository
func (d *DefaultAuthenticateHelper) SearchUser(username string) (*models.User, error) {
	return nil, nil
}

//PostAuthenticate - Update user information after authenticate, such as Onboard or sync info etc
func (d *DefaultAuthenticateHelper) PostAuthenticate(u *models.User) error {
	return nil
}

var registry = make(map[string]AuthenticateHelper)

// Register add different authenticators to registry map.
func Register(name string, h AuthenticateHelper) {
	if _, dup := registry[name]; dup {
		log.Infof("authenticator: %s has been registered,skip", name)
		return
	}
	registry[name] = h
	log.Debugf("Registered authencation helper for auth mode: %s", name)
}

// Login authenticates user credentials based on setting.
func Login(m models.AuthModel) (*models.User, error) {

	authMode, err := config.AuthMode()
	if err != nil {
		return nil, err
	}
	if authMode == "" || dao.IsSuperUser(m.Principal) {
		authMode = common.DBAuth
	}
	log.Debug("Current AUTH_MODE is ", authMode)

	authenticator, ok := registry[authMode]
	if !ok {
		return nil, fmt.Errorf("Unrecognized auth_mode: %s", authMode)
	}
	if lock.IsLocked(m.Principal) {
		log.Debugf("%s is locked due to login failure, login failed", m.Principal)
		return nil, nil
	}
	user, err := authenticator.Authenticate(m)
	if err != nil {
		if _, ok = err.(ErrAuth); ok {
			log.Debugf("Login failed, locking %s, and sleep for %v", m.Principal, frozenTime)
			lock.Lock(m.Principal)
			time.Sleep(frozenTime)
		}
		return nil, err
	}

	err = authenticator.PostAuthenticate(user)

	return user, err
}

func getHelper() (AuthenticateHelper, error) {
	authMode, err := config.AuthMode()
	if err != nil {
		return nil, err
	}
	AuthenticateHelper, ok := registry[authMode]
	if !ok {
		return nil, fmt.Errorf("Can not get authenticator, authmode: %s", authMode)
	}
	return AuthenticateHelper, nil
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, return the user's profile.
func OnBoardUser(user *models.User) error {
	log.Debugf("OnBoardUser, user %+v", user)
	helper, err := getHelper()
	if err != nil {
		return err
	}
	return helper.OnBoardUser(user)
}

// SearchUser --
func SearchUser(username string) (*models.User, error) {
	helper, err := getHelper()
	if err != nil {
		return nil, err
	}
	return helper.SearchUser(username)
}

// SearchAndOnboardUser ...
func SearchAndOnboardUser(username string) (int, error) {
	user, err := SearchUser(username)
	if err != nil {
		return 0, err
	}
	if user != nil {
		err = OnBoardUser(user)
		if err != nil {
			return 0, err
		}
	}
	return user.UserID, nil
}

// PostAuthenticate -
func PostAuthenticate(u *models.User) error {
	helper, err := getHelper()
	if err != nil {
		return err
	}
	return helper.PostAuthenticate(u)
}
