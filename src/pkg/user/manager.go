// Copyright Project Harbor Authors
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

package user

import (
	"context"
	"fmt"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/user/dao"
	"github.com/goharbor/harbor/src/pkg/user/models"
	"strings"
)

var (
	// Mgr is the global project manager
	Mgr = New()
)

// Manager is used for user management
type Manager interface {
	// Get get user by user id
	Get(ctx context.Context, id int) (*commonmodels.User, error)
	// GetByName get user by username, it will return an error if the user does not exist
	GetByName(ctx context.Context, username string) (*commonmodels.User, error)
	// List users according to the query
	List(ctx context.Context, query *q.Query, options ...models.Option) (commonmodels.Users, error)
	// Count counts the number of users according to the query
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Create creates the user, the password of input should be plaintext
	Create(ctx context.Context, user *commonmodels.User) (int, error)
	// Delete deletes the user by updating user's delete flag and update the name and Email
	Delete(ctx context.Context, id int) error
	// SetSysAdminFlag sets the system admin flag of the user in local DB
	SetSysAdminFlag(ctx context.Context, id int, admin bool) error
	// UpdateProfile updates the user's profile
	UpdateProfile(ctx context.Context, user *commonmodels.User, col ...string) error
	// UpdatePassword updates user's password
	UpdatePassword(ctx context.Context, id int, newPassword string) error
	// MatchLocalPassword tries to match the record in DB based on the input, the first return value is
	// the user model corresponding to the entry in DB
	MatchLocalPassword(ctx context.Context, username, password string) (*commonmodels.User, error)
	// Onboard will check if a user exists in user table, if not insert the user and
	// put the id in the pointer of user model, if it does exist, return the user's profile.
	// This is used for ldap and uaa authentication, such the user can have an ID in Harbor.
	Onboard(ctx context.Context, user *commonmodels.User) error
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{dao: dao.New()}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Onboard(ctx context.Context, user *commonmodels.User) error {
	u, err := m.GetByName(ctx, user.Username)
	if err == nil {
		user.Email = u.Email
		user.SysAdminFlag = u.SysAdminFlag
		user.Realname = u.Realname
		user.UserID = u.UserID
		return nil
	} else if !errors.IsNotFoundErr(err) {
		return err
	}
	// User does not exists, insert the user record.
	// Given this func is ALWAYS called in a tx, the conflict error can rollback the tx to ensure the consistency
	id, err2 := m.Create(ctx, user)
	if err2 != nil {
		return err2
	}
	user.UserID = id
	return nil
}

func (m *manager) Delete(ctx context.Context, id int) error {
	u, err := m.Get(ctx, id)
	if err != nil {
		return err
	}
	u.Username = lib.Truncate(u.Username, fmt.Sprintf("#%d", u.UserID), 255)
	u.Email = lib.Truncate(u.Email, fmt.Sprintf("#%d", u.UserID), 255)
	u.Deleted = true
	return m.dao.Update(ctx, u, "username", "email", "deleted")
}

func (m *manager) MatchLocalPassword(ctx context.Context, usernameOrEmail, password string) (*commonmodels.User, error) {
	l, err := m.dao.List(ctx, q.New(q.KeyWords{"username_or_email": usernameOrEmail}))
	if err != nil {
		return nil, err
	}
	for _, entry := range l {
		if utils.Encrypt(password, entry.Salt, entry.PasswordVersion) == entry.Password {
			entry.Password = ""
			return entry, nil
		}
	}
	return nil, nil
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) UpdateProfile(ctx context.Context, user *commonmodels.User, cols ...string) error {
	if cols == nil || len(cols) == 0 {
		cols = []string{"Email", "Realname", "Comment"}
	}
	return m.dao.Update(ctx, user, cols...)
}

func (m *manager) UpdatePassword(ctx context.Context, id int, newPassword string) error {
	user := &commonmodels.User{
		UserID: id,
	}
	injectPasswd(user, newPassword)
	return m.dao.Update(ctx, user, "salt", "password", "password_version")
}

func (m *manager) SetSysAdminFlag(ctx context.Context, id int, admin bool) error {
	u := &commonmodels.User{
		UserID:       id,
		SysAdminFlag: admin,
	}
	return m.dao.Update(ctx, u, "sysadmin_flag")
}

func (m *manager) Create(ctx context.Context, user *commonmodels.User) (int, error) {
	injectPasswd(user, user.Password)
	return m.dao.Create(ctx, user)
}

// Get get user by user id
func (m *manager) Get(ctx context.Context, id int) (*commonmodels.User, error) {
	users, err := m.dao.List(ctx, q.New(q.KeyWords{"user_id": id}))
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.NotFoundError(nil).WithMessage("user %d not found", id)
	}

	return users[0], nil
}

// GetByName get user by username
func (m *manager) GetByName(ctx context.Context, username string) (*commonmodels.User, error) {
	users, err := m.dao.List(ctx, q.New(q.KeyWords{"username": username}))
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.NotFoundError(nil).WithMessage("user %s not found", username)
	}

	return users[0], nil
}

// List users according to the query
func (m *manager) List(ctx context.Context, query *q.Query, options ...models.Option) (commonmodels.Users, error) {
	query = q.MustClone(query)
	for key := range query.Keywords {
		str := strings.ToLower(key)
		if str == "user_id__in" {
			options = append(options, models.WithDefaultAdmin())
			break
		} else if str == "user_id" {
			options = append(options, models.WithDefaultAdmin())
			break
		}
	}
	opts := models.NewOptions(options...)
	if !opts.IncludeDefaultAdmin {
		query.Keywords["user_id__gt"] = 1
	}
	return m.dao.List(ctx, query)
}

func injectPasswd(u *commonmodels.User, password string) {
	salt := utils.GenerateRandomString()
	u.Password = utils.Encrypt(password, salt, utils.SHA256)
	u.Salt = salt
	u.PasswordVersion = utils.SHA256
}
