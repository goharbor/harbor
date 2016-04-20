/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package dao

import (
	"database/sql"
	"errors"

	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils"

	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/utils/log"
)

// GetUser ...
func GetUser(query models.User) (*models.User, error) {

	o := orm.NewOrm()

	sql := `select user_id, username, email, realname, comment, reset_uuid, salt,
		sysadmin_flag, creation_time, update_time
		from user u
		where deleted = 0 `
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

	var u []models.User
	n, err := o.Raw(sql, queryParam).QueryRows(&u)

	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}

	return &u[0], nil
}

// LoginByDb is used for user to login with database auth mode.
func LoginByDb(auth models.AuthModel) (*models.User, error) {
	o := orm.NewOrm()

	var users []models.User
	n, err := o.Raw(`select * from user where (username = ? or email = ?) and deleted = 0`,
		auth.Principal, auth.Principal).QueryRows(&users)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}

	user := users[0]

	if user.Password != utils.Encrypt(auth.Password, user.Salt) {
		return nil, nil
	}

	user.Password = "" //do not return the password

	return &user, nil
}

// ListUsers lists all users according to different conditions.
func ListUsers(query models.User) ([]models.User, error) {
	o := orm.NewOrm()
	u := []models.User{}
	sql := `select  user_id, username, email, realname, comment, reset_uuid, salt,
		sysadmin_flag, creation_time, update_time
		from user u
		where u.deleted = 0 and u.user_id != 1 `

	queryParam := make([]interface{}, 1)
	if query.Username != "" {
		sql += ` and username like ? `
		queryParam = append(queryParam, query.Username)
	}
	sql += ` order by user_id desc `

	_, err := o.Raw(sql, queryParam).QueryRows(&u)
	return u, err
}

// ToggleUserAdminRole gives a user admin role.
func ToggleUserAdminRole(u models.User) error {
	o := orm.NewOrm()

	sql := `update user set sysadmin_flag =not sysadmin_flag where user_id = ?`

	r, err := o.Raw(sql, u.UserID).Exec()
	if err != nil {
		return err
	}

	if _, err := r.RowsAffected(); err != nil {
		return err
	}

	return nil
}

// ChangeUserPassword ...
func ChangeUserPassword(u models.User, oldPassword ...string) (err error) {
	if len(oldPassword) > 1 {
		return errors.New("Wrong numbers of params.")
	}

	o := orm.NewOrm()

	var r sql.Result
	if len(oldPassword) == 0 {
		//In some cases, it may no need to check old password, just as Linux change password policies.
		r, err = o.Raw(`update user set password=?, salt=? where user_id=?`, utils.Encrypt(u.Password, u.Salt), u.Salt, u.UserID).Exec()
	} else {
		r, err = o.Raw(`update user set password=?, salt=? where user_id=? and password = ?`, utils.Encrypt(u.Password, u.Salt), u.Salt, u.UserID, utils.Encrypt(oldPassword[0], u.Salt)).Exec()
	}

	if err != nil {
		return err
	}
	c, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if c == 0 {
		return errors.New("No record has been modified, change password failed.")
	}

	return nil
}

// ResetUserPassword ...
func ResetUserPassword(u models.User) error {
	o := orm.NewOrm()
	r, err := o.Raw(`update user set password=?, reset_uuid=? where reset_uuid=?`, utils.Encrypt(u.Password, u.Salt), "", u.ResetUUID).Exec()
	if err != nil {
		return err
	}
	count, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("No record be changed, reset password failed.")
	}
	return nil
}

// UpdateUserResetUUID ...
func UpdateUserResetUUID(u models.User) error {
	o := orm.NewOrm()
	_, err := o.Raw(`update user set reset_uuid=? where email=?`, u.ResetUUID, u.Email).Exec()
	return err
}

// CheckUserPassword checks whether the password is correct.
func CheckUserPassword(query models.User) (*models.User, error) {

	currentUser, err := GetUser(query)

	if err != nil {
		return nil, err
	}

	if currentUser == nil {
		return nil, nil
	}

	sql := `select user_id, username, salt from user where deleted = 0`

	queryParam := make([]interface{}, 1)

	if query.UserID != 0 {
		sql += ` and password = ? and user_id = ?`
		queryParam = append(queryParam, utils.Encrypt(query.Password, currentUser.Salt))
		queryParam = append(queryParam, query.UserID)
	} else {
		sql += ` and username = ? and password = ?`
		queryParam = append(queryParam, currentUser.Username)
		queryParam = append(queryParam, utils.Encrypt(query.Password, currentUser.Salt))
	}
	o := orm.NewOrm()
	var user []models.User

	n, err := o.Raw(sql, queryParam).QueryRows(&user)

	if err != nil {
		return nil, err
	}

	if n == 0 {
		log.Warning("User principal does not match password. Current:", currentUser)
		return nil, nil
	}

	return &user[0], nil
}

// DeleteUser ...
func DeleteUser(userID int) error {
	o := orm.NewOrm()
	_, err := o.Raw(`update user set deleted = 1 where user_id = ?`, userID).Exec()
	return err
}
