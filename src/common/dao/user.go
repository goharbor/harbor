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
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
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

// LoginByDb is used for user to login with database auth mode.
func LoginByDb(auth models.AuthModel) (*models.User, error) {
	var users []models.User
	o := GetOrmer()

	n, err := o.Raw(`select * from harbor_user where (username = ? or email = ?) and deleted = false`,
		auth.Principal, auth.Principal).QueryRows(&users)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}

	user := users[0]

	if !matchPassword(&user, auth.Password) {
		return nil, nil
	}
	user.Password = "" // do not return the password
	return &user, nil
}

// GetTotalOfUsers ...
func GetTotalOfUsers(query *models.UserQuery) (int64, error) {
	return userQueryConditions(query).Count()
}

// ListUsers lists all users according to different conditions.
func ListUsers(query *models.UserQuery) ([]models.User, error) {
	qs := userQueryConditions(query)
	if query != nil && query.Pagination != nil {
		offset := (query.Pagination.Page - 1) * query.Pagination.Size
		qs = qs.Offset(offset).Limit(query.Pagination.Size)
	}
	users := []models.User{}
	_, err := qs.OrderBy("username").All(&users)
	return users, err
}

func userQueryConditions(query *models.UserQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(&models.User{}).Filter("deleted", 0)

	if query == nil {
		// Exclude admin account, see https://github.com/goharbor/harbor/issues/2527
		return qs.Filter("user_id__gt", 1)
	}

	if len(query.UserIDs) > 0 {
		qs = qs.Filter("user_id__in", query.UserIDs)
	} else {
		// Exclude admin account when not filter by UserIDs, see https://github.com/goharbor/harbor/issues/2527
		qs = qs.Filter("user_id__gt", 1)
	}

	if len(query.Username) > 0 {
		qs = qs.Filter("username__contains", query.Username)
	}

	if len(query.Email) > 0 {
		qs = qs.Filter("email__contains", query.Email)
	}

	return qs
}

// ToggleUserAdminRole gives a user admin role.
func ToggleUserAdminRole(userID int, hasAdmin bool) error {
	o := GetOrmer()
	queryParams := make([]interface{}, 1)
	sql := `update harbor_user set sysadmin_flag = ? where user_id = ?`
	queryParams = append(queryParams, hasAdmin)
	queryParams = append(queryParams, userID)
	r, err := o.Raw(sql, queryParams).Exec()
	if err != nil {
		return err
	}

	if _, err := r.RowsAffected(); err != nil {
		return err
	}

	return nil
}

// ChangeUserPassword ...
func ChangeUserPassword(u models.User) error {
	u.UpdateTime = time.Now()
	u.Salt = utils.GenerateRandomString()
	u.Password = utils.Encrypt(u.Password, u.Salt, utils.SHA256)
	var err error
	if u.PasswordVersion == utils.SHA1 {
		u.PasswordVersion = utils.SHA256
		_, err = GetOrmer().Update(&u, "Password", "PasswordVersion", "Salt", "UpdateTime")
	} else {
		_, err = GetOrmer().Update(&u, "Password", "Salt", "UpdateTime")
	}
	return err
}

// ResetUserPassword ...
func ResetUserPassword(u models.User, rawPassword string) error {
	var rowsAffected int64
	var err error
	u.UpdateTime = time.Now()
	u.Password = utils.Encrypt(rawPassword, u.Salt, utils.SHA256)
	u.ResetUUID = ""
	if u.PasswordVersion == utils.SHA1 {
		u.PasswordVersion = utils.SHA256
		rowsAffected, err = GetOrmer().Update(&u, "Password", "PasswordVersion", "ResetUUID", "UpdateTime")
	} else {
		rowsAffected, err = GetOrmer().Update(&u, "Password", "ResetUUID", "UpdateTime")
	}
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no record be changed, reset password failed")
	}
	return nil
}

// UpdateUserResetUUID ...
func UpdateUserResetUUID(u models.User) error {
	o := GetOrmer()
	_, err := o.Raw(`update harbor_user set reset_uuid=? where email=?`, u.ResetUUID, u.Email).Exec()
	return err
}

// DeleteUser ...
func DeleteUser(userID int) error {
	o := GetOrmer()

	user, err := GetUser(models.User{
		UserID: userID,
	})
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%s#%d", user.Username, user.UserID)
	email := fmt.Sprintf("%s#%d", user.Email, user.UserID)

	_, err = o.Raw(`update harbor_user
		set deleted = true, username = ?, email = ?
		where user_id = ?`, name, email, userID).Exec()
	return err
}

// ChangeUserProfile - Update user in local db,
// cols to specify the columns need to update,
// Email, and RealName, Comment are updated by default.
func ChangeUserProfile(user models.User, cols ...string) error {
	o := GetOrmer()
	if len(cols) == 0 {
		cols = []string{"Email", "Realname", "Comment"}
	}
	if _, err := o.Update(&user, cols...); err != nil {
		log.Errorf("update user failed, error: %v", err)
		return err
	}
	return nil
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

// IsSuperUser checks if the user is super user(conventionally id == 1) of Harbor
func IsSuperUser(username string) bool {
	u, err := GetUser(models.User{
		Username: username,
	})
	log.Debugf("Check if user %s is super user", username)
	if err != nil {
		log.Errorf("Failed to get user from DB, username: %s, error: %v", username, err)
		return false
	}
	return u != nil && u.UserID == 1
}

// CleanUser - Clean this user information from DB
func CleanUser(id int64) error {
	_, err := GetOrmer().QueryTable(&models.User{}).Filter("UserID", id).Delete()
	return err
}

// MatchPassword returns true is password matched
func matchPassword(u *models.User, password string) bool {
	return utils.Encrypt(password, u.Salt, u.PasswordVersion) == u.Password
}
