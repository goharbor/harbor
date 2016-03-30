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
	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/models"
)

// AddProjectMember inserts a record to table project_member
func AddProjectMember(projectID int64, userID int, role int) error {
	o := orm.NewOrm()

	sql := "insert into project_member (project_id, user_id , role) values (?, ?, ?)"

	_, err := o.Raw(sql, projectID, userID, role).Exec()

	return err
}

// UpdateProjectMember updates the record in table project_member
func UpdateProjectMember(projectID int64, userID int, role int) error {
	o := orm.NewOrm()

	sql := "update project_member set role = ? where project_id = ? and user_id = ?"

	_, err := o.Raw(sql, role, projectID, userID).Exec()

	return err
}

// DeleteProjectMember delete the record from table project_member
func DeleteProjectMember(projectID int64, userID int) error {
	o := orm.NewOrm()

	sql := "delete from project_member where project_id = ? and user_id = ?"

	if _, err := o.Raw(sql, projectID, userID).Exec(); err != nil {
		return err
	}

	return nil
}

// GetUserByProject gets all members of the project.
func GetUserByProject(projectID int64, queryUser models.User) ([]models.User, error) {
	o := orm.NewOrm()
	u := []models.User{}
	sql := `select u.user_id, u.username, r.name rolename, r.role_id
		from user u 
		join project_member pm 
		on pm.project_id = ? and u.user_id = pm.user_id 
		join role r
		on pm.role = r.role_id
		where u.deleted = 0`

	queryParam := make([]interface{}, 1)
	queryParam = append(queryParam, projectID)

	if queryUser.Username != "" {
		sql += " and u.username like ? "
		queryParam = append(queryParam, queryUser.Username)
	}
	sql += ` order by u.user_id `
	_, err := o.Raw(sql, queryParam).QueryRows(&u)
	return u, err
}
