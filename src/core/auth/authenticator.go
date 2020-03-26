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

package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
)

// 1.5 seconds
const frozenTime time.Duration = 1500 * time.Millisecond

var lock = NewUserLock(frozenTime)

// ErrorUserNotExist ...
var ErrorUserNotExist = errors.New("User does not exist")

// ErrorGroupNotExist ...
var ErrorGroupNotExist = errors.New("Group does not exist")

// ErrDuplicateLDAPGroup ...
var ErrDuplicateLDAPGroup = errors.New("An LDAP user group with same DN already exist")

// ErrInvalidLDAPGroupDN ...
var ErrInvalidLDAPGroupDN = errors.New("The LDAP group DN is invalid")

// ErrAuth is the type of error to indicate a failed authentication due to user's error.
type ErrAuth struct {
	details string
}

// Error ...
func (ea ErrAuth) Error() string {
	return fmt.Sprintf("Failed to authenticate user, due to error '%s'", ea.details)
}

// NewErrAuth ...
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
	// Create a group in harbor DB, if altGroupName is not empty, take the altGroupName as groupName in harbor DB.
	OnBoardGroup(g *models.UserGroup, altGroupName string) error
	// Get user information from account repository
	SearchUser(username string) (*models.User, error)
	// Search a group based on specific authentication
	SearchGroup(groupDN string) (*models.UserGroup, error)
	// Update user information after authenticate, such as OnBoard or sync info etc
	PostAuthenticate(u *models.User) error
}

// DefaultAuthenticateHelper - default AuthenticateHelper implementation
type DefaultAuthenticateHelper struct {
}

// Authenticate ...
func (d *DefaultAuthenticateHelper) Authenticate(m models.AuthModel) (*models.User, error) {
	return nil, errors.New("Not supported")
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, fill in the user model based
// on the data record of the user
func (d *DefaultAuthenticateHelper) OnBoardUser(u *models.User) error {
	return errors.New("Not supported")
}

// SearchUser - Get user information from account repository
func (d *DefaultAuthenticateHelper) SearchUser(username string) (*models.User, error) {
	return nil, errors.New("Not supported")
}

// PostAuthenticate - Update user information after authenticate, such as OnBoard or sync info etc
func (d *DefaultAuthenticateHelper) PostAuthenticate(u *models.User) error {
	return nil
}

// OnBoardGroup - OnBoardGroup, it will set the ID of the user group, if altGroupName is not empty, take the altGroupName as groupName in harbor DB.
func (d *DefaultAuthenticateHelper) OnBoardGroup(u *models.UserGroup, altGroupName string) error {
	return errors.New("Not supported")
}

// SearchGroup - Search ldap group by group key, groupKey is the unique attribute of group in authenticator, for LDAP, the key is group DN
func (d *DefaultAuthenticateHelper) SearchGroup(groupKey string) (*models.UserGroup, error) {
	return nil, errors.New("Not supported")
}

var registry = make(map[string]AuthenticateHelper)

// Register add different authenticators to registry map.
func Register(name string, h AuthenticateHelper) {
	if _, dup := registry[name]; dup {
		log.Infof("authenticator: %s has been registered,skip", name)
		return
	}
	registry[name] = h
	log.Debugf("Registered authentication helper for auth mode: %s", name)
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

// OnBoardGroup - Create a user group in harbor db, if altGroupName is not empty, take the altGroupName as groupName in harbor DB
func OnBoardGroup(userGroup *models.UserGroup, altGroupName string) error {
	helper, err := getHelper()
	if err != nil {
		return err
	}
	return helper.OnBoardGroup(userGroup, altGroupName)
}

// SearchGroup -- Search group in authenticator, groupKey is the unique attribute of group in authenticator, for LDAP, the key is group DN
func SearchGroup(groupKey string) (*models.UserGroup, error) {
	helper, err := getHelper()
	if err != nil {
		return nil, err
	}
	return helper.SearchGroup(groupKey)
}

// SearchAndOnBoardUser ... Search user and OnBoard user, if user exist, return the ID of current user.
func SearchAndOnBoardUser(username string) (int, error) {
	user, err := SearchUser(username)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, ErrorUserNotExist
	}
	err = OnBoardUser(user)
	if err != nil {
		return 0, err
	}
	return user.UserID, nil
}

// SearchAndOnBoardGroup ... if altGroupName is not empty, take the altGroupName as groupName in harbor DB
func SearchAndOnBoardGroup(groupKey, altGroupName string) (int, error) {
	userGroup, err := SearchGroup(groupKey)
	if err != nil {
		return 0, err
	}
	if userGroup == nil {
		return 0, ErrorGroupNotExist
	}
	if userGroup != nil {
		err = OnBoardGroup(userGroup, altGroupName)
	}
	return userGroup.ID, err
}

// PostAuthenticate -
func PostAuthenticate(u *models.User) error {
	helper, err := getHelper()
	if err != nil {
		return err
	}
	return helper.PostAuthenticate(u)
}
