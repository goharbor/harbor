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
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/config"
	libErrors "github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/user"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
)

// 1.5 seconds
const frozenTime time.Duration = 1500 * time.Millisecond

var lock = NewUserLock(frozenTime)

// ErrorUserNotExist ...
var ErrorUserNotExist = errors.New("user does not exist")

// ErrorGroupNotExist ...
var ErrorGroupNotExist = errors.New("group does not exist")

// ErrDuplicateLDAPGroup ...
var ErrDuplicateLDAPGroup = errors.New("a LDAP user group with same DN already exist")

// ErrInvalidLDAPGroupDN ...
var ErrInvalidLDAPGroupDN = errors.New("the LDAP group DN is invalid")

// ErrNotSupported ...
var ErrNotSupported = errors.New("not supported")

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
	Authenticate(ctx context.Context, m models.AuthModel) (*models.User, error)
	// OnBoardUser will check if a user exists in user table, if not insert the user and
	// put the id in the pointer of user model, if it does exist, fill in the user model based
	// on the data record of the user
	OnBoardUser(ctx context.Context, u *models.User) error
	// OnBoardGroup Create a group in harbor DB, if altGroupName is not empty, take the altGroupName as groupName in harbor DB.
	OnBoardGroup(ctx context.Context, g *model.UserGroup, altGroupName string) error
	// SearchUser Get user information from account repository
	SearchUser(ctx context.Context, username string) (*models.User, error)
	// SearchGroup Search a group based on specific authentication
	SearchGroup(ctx context.Context, groupDN string) (*model.UserGroup, error)
	// PostAuthenticate Update user information after authenticate, such as Onboard or sync info etc
	PostAuthenticate(ctx context.Context, u *models.User) error
}

// DefaultAuthenticateHelper - default AuthenticateHelper implementation
type DefaultAuthenticateHelper struct {
}

// Authenticate ...
func (d *DefaultAuthenticateHelper) Authenticate(ctx context.Context, m models.AuthModel) (*models.User, error) {
	return nil, ErrNotSupported
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, fill in the user model based
// on the data record of the user
func (d *DefaultAuthenticateHelper) OnBoardUser(ctx context.Context, u *models.User) error {
	return ErrNotSupported
}

// SearchUser - Get user information from account repository
func (d *DefaultAuthenticateHelper) SearchUser(ctx context.Context, username string) (*models.User, error) {
	log.Errorf("Not support searching user, username: %s", username)
	return nil, libErrors.NotFoundError(ErrNotSupported).WithMessage("%s not found", username)
}

// PostAuthenticate - Update user information after authenticate, such as OnBoard or sync info etc
func (d *DefaultAuthenticateHelper) PostAuthenticate(ctx context.Context, u *models.User) error {
	return nil
}

// OnBoardGroup - OnBoardGroup, it will set the ID of the user group, if altGroupName is not empty, take the altGroupName as groupName in harbor DB.
func (d *DefaultAuthenticateHelper) OnBoardGroup(ctx context.Context, u *model.UserGroup, altGroupName string) error {
	return ErrNotSupported
}

// SearchGroup - Search ldap group by group key, groupKey is the unique attribute of group in authenticator, for LDAP, the key is group DN
func (d *DefaultAuthenticateHelper) SearchGroup(ctx context.Context, groupKey string) (*model.UserGroup, error) {
	log.Errorf("Not support searching group, group key: %s", groupKey)
	return nil, libErrors.NotFoundError(ErrNotSupported).WithMessage("%s not found", groupKey)
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
func Login(ctx context.Context, m models.AuthModel) (*models.User, error) {
	authMode, err := config.AuthMode(ctx)
	if err != nil {
		return nil, err
	}
	if authMode == "" || IsSuperUser(ctx, m.Principal) {
		authMode = common.DBAuth
	}
	log.Debug("Current AUTH_MODE is ", authMode)

	authenticator, ok := registry[authMode]
	if !ok {
		return nil, fmt.Errorf("unrecognized auth_mode: %s", authMode)
	}
	if lock.IsLocked(m.Principal) {
		log.Debugf("%s is locked due to login failure, login failed", m.Principal)
		return nil, nil
	}
	user, err := authenticator.Authenticate(ctx, m)
	if err != nil {
		if _, ok = err.(ErrAuth); ok {
			log.Debugf("Login failed, locking %s, and sleep for %v", m.Principal, frozenTime)
			lock.Lock(m.Principal)
			time.Sleep(frozenTime)
		}
		return nil, err
	}
	err = authenticator.PostAuthenticate(ctx, user)
	return user, err
}

func getHelper(ctx context.Context) (AuthenticateHelper, error) {
	authMode, err := config.AuthMode(ctx)
	if err != nil {
		return nil, err
	}
	AuthenticateHelper, ok := registry[authMode]
	if !ok {
		return nil, fmt.Errorf("can not get authenticator, authmode: %s", authMode)
	}
	return AuthenticateHelper, nil
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, return the user's profile.
func OnBoardUser(ctx context.Context, user *models.User) error {
	log.Debugf("OnBoardUser, user: %v", user.Username)
	helper, err := getHelper(ctx)
	if err != nil {
		return err
	}
	return helper.OnBoardUser(ctx, user)
}

// SearchUser --
func SearchUser(ctx context.Context, username string) (*models.User, error) {
	helper, err := getHelper(ctx)
	if err != nil {
		return nil, err
	}
	return helper.SearchUser(ctx, username)
}

// OnBoardGroup - Create a user group in harbor db, if altGroupName is not empty, take the altGroupName as groupName in harbor DB
func OnBoardGroup(ctx context.Context, userGroup *model.UserGroup, altGroupName string) error {
	helper, err := getHelper(ctx)
	if err != nil {
		return err
	}
	return helper.OnBoardGroup(ctx, userGroup, altGroupName)
}

// SearchGroup -- Search group in authenticator, groupKey is the unique attribute of group in authenticator, for LDAP, the key is group DN
func SearchGroup(ctx context.Context, groupKey string) (*model.UserGroup, error) {
	helper, err := getHelper(ctx)
	if err != nil {
		return nil, err
	}
	return helper.SearchGroup(ctx, groupKey)
}

// SearchAndOnBoardUser ... Search user and OnBoard user, if user exist, return the ID of current user.
func SearchAndOnBoardUser(ctx context.Context, username string) (int, error) {
	user, err := SearchUser(ctx, username)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, libErrors.NotFoundError(nil).WithMessage(fmt.Sprintf("user %s is not found", username))
	}
	err = OnBoardUser(ctx, user)
	if err != nil {
		return 0, err
	}
	return user.UserID, nil
}

// SearchAndOnBoardGroup ... if altGroupName is not empty, take the altGroupName as groupName in harbor DB
func SearchAndOnBoardGroup(ctx context.Context, groupKey, altGroupName string) (int, error) {
	userGroup, err := SearchGroup(ctx, groupKey)
	if err != nil {
		return 0, err
	}
	if userGroup == nil {
		return 0, ErrorGroupNotExist
	}
	if userGroup != nil {
		err = OnBoardGroup(ctx, userGroup, altGroupName)
	}
	return userGroup.ID, err
}

// PostAuthenticate -
func PostAuthenticate(ctx context.Context, u *models.User) error {
	helper, err := getHelper(ctx)
	if err != nil {
		return err
	}
	return helper.PostAuthenticate(ctx, u)
}

// IsSuperUser checks if the user is super user(conventionally id == 1) of Harbor
func IsSuperUser(ctx context.Context, username string) bool {
	u, err := user.Mgr.GetByName(ctx, username)
	if err != nil {
		// LDAP user can't be found before onboard to Harbor
		log.Debugf("Failed to get user from DB, username: %s, error: %v", username, err)
		return false
	}
	return u.UserID == 1
}
