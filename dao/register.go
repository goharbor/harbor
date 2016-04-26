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
	"errors"
	"regexp"
	"time"

	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils"

	"github.com/astaxie/beego/orm"
)

// Register is used for user to register, the password is encrypted before the record is inserted into database.
func Register(user models.User) (int64, error) {

	err := validate(user)
	if err != nil {
		return 0, err
	}

	o := orm.NewOrm()

	p, err := o.Raw("insert into user (username, password, realname, email, comment, salt, sysadmin_flag, creation_time, update_time) values (?, ?, ?, ?, ?, ?, ?, ?, ?)").Prepare()
	if err != nil {
		return 0, err
	}
	defer p.Close()

	salt, err := GenerateRandomString()
	if err != nil {
		return 0, err
	}

	now := time.Now()
	r, err := p.Exec(user.Username, utils.Encrypt(user.Password, salt), user.Realname, user.Email, user.Comment, salt, user.HasAdminRole, now, now)

	if err != nil {
		return 0, err
	}
	userID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func validate(user models.User) error {

	if isIllegalLength(user.Username, 0, 20) {
		return errors.New("Username with illegal length.")
	}
	if isContainIllegalChar(user.Username, []string{",", "~", "#", "$", "%"}) {
		return errors.New("Username contains illegal characters.")
	}

	if exist, _ := UserExists(models.User{Username: user.Username}, "username"); exist {
		return errors.New("Username already exists.")
	}

	if len(user.Email) > 0 {
		if m, _ := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, user.Email); !m {
			return errors.New("Email with illegal format.")
		}
		if exist, _ := UserExists(models.User{Email: user.Email}, "email"); exist {
			return errors.New("Email already exists.")
		}
	}

	if isIllegalLength(user.Realname, 0, 20) {
		return errors.New("Realname with illegal length.")
	}

	if isContainIllegalChar(user.Realname, []string{",", "~", "#", "$", "%"}) {
		return errors.New("Realname contains illegal characters.")
	}

	if isIllegalLength(user.Password, 0, 20) {
		return errors.New("Password with illegal length.")
	}

	if isIllegalLength(user.Comment, -1, 30) {
		return errors.New("Comment with illegal length.")
	}
	return nil
}

// UserExists returns whether a user exists according username or Email.
func UserExists(user models.User, target string) (bool, error) {

	if user.Username == "" && user.Email == "" {
		return false, errors.New("User name and email are blank.")
	}

	o := orm.NewOrm()

	sql := `select user_id from user where 1=1 `
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
