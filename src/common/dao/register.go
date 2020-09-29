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
	"errors"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
)

// Register is used for user to register, the password is encrypted before the record is inserted into database.
func Register(user models.User) (int64, error) {

	o := GetOrmer()
	now := time.Now()
	salt := utils.GenerateRandomString()
	sql := `insert into harbor_user
				(username, password, password_version, realname, email, comment, salt, sysadmin_flag, creation_time, update_time)
				 values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING user_id`
	var userID int64
	err := o.Raw(sql, user.Username, utils.Encrypt(user.Password, salt, utils.SHA256), utils.SHA256, user.Realname, user.Email,
		user.Comment, salt, user.SysAdminFlag, now, now).QueryRow(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil

}

// UserExists returns whether a user exists according username or Email.
func UserExists(user models.User, target string) (bool, error) {

	if user.Username == "" && user.Email == "" {
		return false, errors.New("user name and email are blank")
	}

	o := GetOrmer()

	sql := `select user_id from harbor_user where 1=1 `
	queryParam := make([]interface{}, 1)

	switch target {
	case "username":
		sql += ` and username = ? `
		queryParam = append(queryParam, user.Username)
	case "email":
		sql += ` and email = ? `
		queryParam = append(queryParam, user.Email)
	}

	var u []models.User
	n, err := o.Raw(sql, queryParam).QueryRows(&u)
	if err != nil {
		return false, err
	} else if n == 0 {
		return false, nil
	} else {
		return true, nil
	}
}
