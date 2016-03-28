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

/*
import (
	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego/orm"
)


// GetUserByProject gets all members of the project.
func GetUserByProject(projectID int64, queryUser models.User) ([]models.User, error) {
	o := orm.NewOrm()
	u := []models.User{}
	sql := `select
			u.user_id, u.username, r.name rolename, r.role_id
		from user u left join user_project_role upr
		    on u.user_id = upr.user_id
		left join project_role pr
			on pr.pr_id = upr.pr_id
		left join role r
			on r.role_id = pr.role_id
		where u.deleted = 0
		  and pr.project_id = ? `

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
*/
