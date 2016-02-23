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

	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

func GetUser(query models.User) (*models.User, error) {

	o := orm.NewOrm()

	sql := `select user_id, username, email, realname, reset_uuid, salt,
			ifnull((select pr.role_id  
			  from project_role pr 
			     left join user_project_role upr on upr.pr_id = pr.pr_id
			  where pr.role_id = 1
			     and upr.user_id = u.user_id),0) as has_admin_role
		from user u
		where deleted = 0 `
	queryParam := make([]interface{}, 1)
	if query.UserId != 0 {
		sql += ` and user_id = ? `
		queryParam = append(queryParam, query.UserId)
	}

	if query.Username != "" {
		sql += ` and username = ? `
		queryParam = append(queryParam, query.Username)
	}

	if query.ResetUuid != "" {
		sql += ` and reset_uuid = ? `
		queryParam = append(queryParam, query.ResetUuid)
	}

	var u []models.User
	n, err := o.Raw(sql, queryParam).QueryRows(&u)

	if err != nil {
		return nil, err
	} else if n == 0 {
		return nil, nil
	} else {
		return &u[0], nil
	}
}

func LoginByDb(auth models.AuthModel) (*models.User, error) {

	query := models.User{Username: auth.Principal, Email: auth.Principal}

	o := orm.NewOrm()
	var u []models.User
	n, err := o.Raw(`select username from user where (username = ? or email = ?)`, query.Username, query.Email).QueryRows(&u)
	if err != nil {
		return nil, err
	} else if n == 0 {
		beego.Warning("User does not exist. Principal:", auth.Principal)
		return nil, nil
	} else {
		u[0].Password = auth.Password
		return CheckUserPassword(u[0])
	}

}

func ListUsers(query models.User) ([]models.User, error) {
	o := orm.NewOrm()
	u := []models.User{}
	sql := `select  u.user_id, u.username, u.email, ifnull((select pr.role_id  
			  from project_role pr 
			     left join user_project_role upr on upr.pr_id = pr.pr_id
			  where pr.role_id = 1
			     and upr.user_id = u.user_id),0) as has_admin_role
		 from user u
		    where u.deleted = 0 and u.user_id != 1 `

	queryParam := make([]interface{}, 1)
	if query.Username != "" {
		sql += ` and u.username like ? `
		queryParam = append(queryParam, query.Username)
	}
	sql += ` order by u.user_id desc `

	_, err := o.Raw(sql, queryParam).QueryRows(&u)
	return u, err
}

func ToggleUserAdminRole(u models.User) error {

	projectRole := models.ProjectRole{PrId: 1} //admin project role

	o := orm.NewOrm()

	var pr []models.ProjectRole

	n, err := o.Raw(`select user_id from user_project_role where user_id = ? and pr_id = ? `, u.UserId, projectRole.PrId).QueryRows(&pr)
	if err != nil {
		return err
	}

	var sql string
	if n == 0 {
		sql = `insert into user_project_role (user_id, pr_id) values (?, ?)`
	} else {
		sql = `delete from user_project_role where user_id = ? and pr_id = ?`
	}

	p, err := o.Raw(sql).Prepare()
	if err != nil {
		return err
	}
	defer p.Close()
	_, err = p.Exec(u.UserId, projectRole.PrId)

	return err
}

func ChangeUserPassword(u models.User, oldPassword string) error {
	o := orm.NewOrm()
	_, err := o.Raw(`update user set password=?, salt=? where user_id=? and password = ?`, utils.Encrypt(u.Password, u.Salt), u.Salt, u.UserId, utils.Encrypt(oldPassword, u.Salt)).Exec()
	return err
}

func ResetUserPassword(u models.User) error {
	o := orm.NewOrm()
	r, err := o.Raw(`update user set password=?, reset_uuid=? where reset_uuid=?`, utils.Encrypt(u.Password, u.Salt), "", u.ResetUuid).Exec()
	if err != nil {
		return err
	}
	count, err := r.RowsAffected()
	if count == 0 {
		return errors.New("No record be changed.")
	}
	return err
}

func UpdateUserResetUuid(u models.User) error {
	o := orm.NewOrm()
	_, err := o.Raw(`update user set reset_uuid=? where email=?`, u.ResetUuid, u.Email).Exec()
	return err
}

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

	if query.UserId != 0 {
		sql += ` and password = ? and user_id = ?`
		queryParam = append(queryParam, utils.Encrypt(query.Password, currentUser.Salt))
		queryParam = append(queryParam, query.UserId)
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
	} else if n == 0 {
		beego.Warning("User principal does not match password. Current:", currentUser)
		return nil, nil
	} else {
		return &user[0], nil
	}
}

func DeleteUser(userId int) error {
	o := orm.NewOrm()
	_, err := o.Raw(`update user set deleted = 1 where user_id = ?`, userId).Exec()
	return err
}
