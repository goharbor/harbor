//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package dao

import (
	"context"
	"database/sql"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/user/models"
	"time"
)

// User holds the details of a user.
// only used in DAO, for other place, use the User model in common/models
type User struct {
	UserID   int    `orm:"pk;auto;column(user_id)" json:"user_id"`
	Username string `orm:"column(username)" json:"username" sort:"default"`
	// Email defined as sql.NullString because sometimes email is missing in LDAP/OIDC auth,
	// set it to null to avoid unique constraint check
	Email           sql.NullString `orm:"column(email)" json:"email"`
	Password        string         `orm:"column(password)" json:"password"`
	PasswordVersion string         `orm:"column(password_version)" json:"password_version"`
	Realname        string         `orm:"column(realname)" json:"realname"`
	Comment         string         `orm:"column(comment)" json:"comment"`
	Deleted         bool           `orm:"column(deleted)" json:"deleted"`
	SysAdminFlag    bool           `orm:"column(sysadmin_flag)" json:"sysadmin_flag"`
	ResetUUID       string         `orm:"column(reset_uuid)" json:"reset_uuid"`
	Salt            string         `orm:"column(salt)" json:"-"`
	CreationTime    time.Time      `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime      time.Time      `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName ...
func (u *User) TableName() string {
	return models.UserTable
}

// toDBUser ...
func toDBUser(u *commonmodels.User) *User {
	user := &User{}

	user.UserID = u.UserID
	user.Username = u.Username
	user.Email = sql.NullString{}
	if u.Email != "" {
		user.Email = sql.NullString{String: u.Email, Valid: true}
	}
	user.Password = u.Password
	user.PasswordVersion = u.PasswordVersion
	user.Realname = u.Realname
	user.Comment = u.Comment
	user.Deleted = u.Deleted
	user.SysAdminFlag = u.SysAdminFlag
	user.ResetUUID = u.ResetUUID
	user.Salt = u.Salt
	user.CreationTime = u.CreationTime
	user.UpdateTime = u.UpdateTime
	return user
}

// toCommonUser ...
func toCommonUser(u *User) *commonmodels.User {
	user := &commonmodels.User{}
	user.UserID = u.UserID
	user.Username = u.Username
	user.Email = u.Email.String

	user.Password = u.Password
	user.PasswordVersion = u.PasswordVersion
	user.Realname = u.Realname
	user.Comment = u.Comment
	user.Deleted = u.Deleted
	user.SysAdminFlag = u.SysAdminFlag
	user.ResetUUID = u.ResetUUID
	user.Salt = u.Salt
	user.CreationTime = u.CreationTime
	user.UpdateTime = u.UpdateTime
	user.GroupIDs = make([]int, 0)
	return user
}

// FilterByUsernameOrEmail generates the query setter to match username or email column to the same value
func (u *User) FilterByUsernameOrEmail(ctx context.Context, qs orm.QuerySeter, key string, value interface{}) orm.QuerySeter {
	usernameOrEmail, ok := value.(string)
	if !ok {
		return qs
	}
	subCond := orm.NewCondition()
	subCond = subCond.Or("Username", usernameOrEmail).Or("Email", usernameOrEmail)

	conds := qs.GetCond()
	if conds == nil {
		conds = orm.NewCondition()
	}
	qs = qs.SetCond(conds.AndCond(subCond))
	return qs
}
