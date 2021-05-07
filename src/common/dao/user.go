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

package dao

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/models"
)

// GetUser ...
func GetUser(query models.User) (*models.User, error) {

	o := GetOrmer()

	sql := `select user_id, username, password, password_version, email, realname, comment, reset_uuid, salt,
		sysadmin_flag, creation_time, update_time
		from harbor_user u
		where deleted = false `
	queryParam := make([]interface{}, 1)
	if query.UserID != 0 {
		sql += ` and user_id = ? `
		queryParam = append(queryParam, query.UserID)
	}

	if query.Username != "" {
		sql += ` and username = ? `
		queryParam = append(queryParam, query.Username)
	}

	if query.ResetUUID != "" {
		sql += ` and reset_uuid = ? `
		queryParam = append(queryParam, query.ResetUUID)
	}

	if query.Email != "" {
		sql += ` and email = ? `
		queryParam = append(queryParam, query.Email)
	}

	var u []models.User
	n, err := o.Raw(sql, queryParam).QueryRows(&u)

	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}

	if n > 1 {
		return nil, fmt.Errorf("got more than one user when executing: %s param: %v", sql, queryParam)
	}

	return &u[0], nil
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, return the user's profile.
// This is used for ldap and uaa authentication, such the user can have an ID in Harbor.
func OnBoardUser(u *models.User) error {
	o := GetOrmer()
	created, id, err := o.ReadOrCreate(u, "Username")
	if err != nil {
		return err
	}
	if created {
		u.UserID = int(id)
		// current orm framework doesn't support to fetch a pointer or sql.NullString with QueryRow
		// https://github.com/astaxie/beego/issues/3767
		if len(u.Email) == 0 {
			_, err = o.Raw("update harbor_user set email = null where user_id = ? ", id).Exec()
			if err != nil {
				return err
			}
		}
	} else {
		existing, err := GetUser(*u)
		if err != nil {
			return err
		}
		u.Email = existing.Email
		u.SysAdminFlag = existing.SysAdminFlag
		u.Realname = existing.Realname
		u.UserID = existing.UserID
	}
	return nil
}

// CleanUser - Clean this user information from DB
func CleanUser(id int64) error {
	_, err := GetOrmer().QueryTable(&models.User{}).Filter("UserID", id).Delete()
	return err
}
