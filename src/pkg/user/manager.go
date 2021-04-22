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
	"strings"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/user/dao"
	"github.com/goharbor/harbor/src/pkg/user/models"
)

var (
	// Mgr is the global project manager
	Mgr = New()
)

// Manager is used for user management
type Manager interface {
	// Get get user by user id
	Get(ctx context.Context, id int) (*models.User, error)
	// GetByName get user by username
	GetByName(ctx context.Context, username string) (*models.User, error)
	// List users according to the query
	List(ctx context.Context, query *q.Query) (models.Users, error)
	// Count counts the number of users according to the query
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Create creates the user, the password of input should be plaintext
	Create(ctx context.Context, user *models.User) (int, error)
	// Delete deletes the user by updating user's delete flag and update the name and Email
	Delete(ctx context.Context, id int) error
	// SetSysAdminFlag sets the system admin flag of the user in local DB
	SetSysAdminFlag(ctx context.Context, id int, admin bool) error
	// UpdateProfile updates the user's profile
	UpdateProfile(ctx context.Context, user *models.User) error
	// UpdatePassword updates user's password
	UpdatePassword(ctx context.Context, id int, newPassword string) error
	// VerifyLocalPassword verifies the password against the record in DB based on the input
	VerifyLocalPassword(ctx context.Context, username, password string) (bool, error)
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{dao: dao.New()}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Delete(ctx context.Context, id int) error {
	u, err := m.Get(ctx, id)
	if err != nil {
		return err
	}
	u.Username = fmt.Sprintf("%s#%d", u.Username, u.UserID)
	u.Email = fmt.Sprintf("%s#%d", u.Email, u.UserID)
	u.Deleted = true
	return m.dao.Update(ctx, u, "username", "email", "deleted")
}

func (m *manager) VerifyLocalPassword(ctx context.Context, username, password string) (bool, error) {
	u, err := m.GetByName(ctx, username)
	if err != nil {
		return false, err
	}
	return utils.Encrypt(password, u.Salt, u.PasswordVersion) == u.Password, nil
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) UpdateProfile(ctx context.Context, user *models.User) error {
	return m.dao.Update(ctx, user, "email", "realname", "comment")
}

func (m *manager) UpdatePassword(ctx context.Context, id int, newPassword string) error {
	user := &models.User{
		UserID: id,
	}
	injectPasswd(user, newPassword)
	return m.dao.Update(ctx, user, "salt", "password", "password_version")
}

func (m *manager) SetSysAdminFlag(ctx context.Context, id int, admin bool) error {
	u := &models.User{
		UserID:       id,
		SysAdminFlag: admin,
	}
	return m.dao.Update(ctx, u, "sysadmin_flag")
}

func (m *manager) Create(ctx context.Context, user *models.User) (int, error) {
	injectPasswd(user, user.Password)
	return m.dao.Create(ctx, user)
}

// Get get user by user id
func (m *manager) Get(ctx context.Context, id int) (*models.User, error) {
	users, err := m.dao.List(ctx, q.New(q.KeyWords{"user_id": id}))
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.NotFoundError(nil).WithMessage("user %d not found", id)
	}

	return users[0], nil
}

// Get get user by username
func (m *manager) GetByName(ctx context.Context, username string) (*models.User, error) {
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
func (m *manager) List(ctx context.Context, query *q.Query) (models.Users, error) {
	query = q.MustClone(query)
	excludeAdmin := true
	for key := range query.Keywords {
		str := strings.ToLower(key)
		if str == "user_id__in" {
			excludeAdmin = false
			break
		} else if str == "user_id" {
			excludeAdmin = false
			break
		}
	}
	if excludeAdmin {
		// Exclude admin account when not filter by UserIDs, see https://github.com/goharbor/harbor/issues/2527
		query.Keywords["user_id__gt"] = 1
	}
	return m.dao.List(ctx, query)
}

func injectPasswd(u *models.User, password string) {
	salt := utils.GenerateRandomString()
	u.Password = utils.Encrypt(password, salt, utils.SHA256)
	u.Salt = salt
	u.PasswordVersion = utils.SHA256
}
