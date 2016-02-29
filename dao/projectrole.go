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
	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego/orm"
)

// AddProjectRole ...
func AddProjectRole(projectRole models.ProjectRole) (int64, error) {
	o := orm.NewOrm()
	p, err := o.Raw("insert into project_role (project_id, role_id) values (?, ?)").Prepare()
	if err != nil {
		return 0, err
	}
	defer p.Close()
	r, err := p.Exec(projectRole.ProjectID, projectRole.RoleID)
	if err != nil {
		return 0, err
	}
	id, err := r.LastInsertId()
	return id, err
}

// AddUserProjectRole inserts role information to table project_role and user_project_role.
func AddUserProjectRole(userID int, projectID int64, roleID int) error {

	o := orm.NewOrm()

	var pr []models.ProjectRole

	var prID int

	sql := `select pr.pr_id, pr.project_id, pr.role_id from project_role pr where pr.project_id = ? and pr.role_id = ?`
	n, err := o.Raw(sql, projectID, roleID).QueryRows(&pr)
	if err != nil {
		return err
	}

	if n == 0 { //project role not found, insert a pr record
		p, err := o.Raw("insert into project_role (project_id, role_id) values (?, ?)").Prepare()
		if err != nil {
			return err
		}
		defer p.Close()
		r, err := p.Exec(projectID, roleID)
		if err != nil {
			return err
		}
		id, err := r.LastInsertId()
		if err != nil {
			return err
		}
		prID = int(id)
	} else if n > 0 {
		prID = pr[0].PrID
	}
	p, err := o.Raw("insert into user_project_role (user_id, pr_id) values (?, ?)").Prepare()
	if err != nil {
		return err
	}
	defer p.Close()
	_, err = p.Exec(userID, prID)
	return err
}

// DeleteUserProjectRoles ...
func DeleteUserProjectRoles(userID int, projectID int64) error {
	o := orm.NewOrm()
	sql := `delete from user_project_role where user_id = ? and pr_id in
		(select pr_id from project_role where project_id = ?)`
	p, err := o.Raw(sql).Prepare()
	if err != nil {
		return err
	}
	_, err = p.Exec(userID, projectID)
	return err
}
