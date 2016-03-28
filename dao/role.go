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
	"fmt"

	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego/orm"
)

type role int

// Start from 2 to guarantee the compatibility with former code
const (
	ProjectAdmin role = 2
	Developer         = 3
	Guest             = 4
)

var roleList = make(map[role]*models.Role)

// IntToRole is used to convert int to role.
func IntToRole(i int) (r role, err error) {
	switch i {
	case 2:
		r = ProjectAdmin
	case 3:
		r = Developer
	case 4:
		r = Guest
	default:
		err = fmt.Errorf("no role is correspondent with the input: %d", i)
	}
	return
}

// GetUserProjectRoles returns roles that the user has according to the project.
func GetUserProjectRoles(userQuery models.User, projectID int64) ([]models.Role, error) {

	o := orm.NewOrm()

	sql := `select *
		from role
		where role_id = 
			(
				select role
				from project_member
				where project_id = ? and user_id = ?
			)`
	queryParam := make([]interface{}, 1)
	queryParam = append(queryParam, userQuery.UserID)

	var roleList []models.Role
	_, err := o.Raw(sql, projectID, userQuery.UserID).QueryRows(&roleList)

	if err != nil {
		return nil, err
	}
	return roleList, nil
}

// IsAdminRole returns whether the user  is admin.
func IsAdminRole(userID int) (bool, error) {

	user, err := GetUser(models.User{UserID: userID})
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	return user.HasAdminRole == 1, nil
}

func getRole(r role) (*models.Role, error) {
	if roleList[r] != nil {
		return roleList[r], nil
	}

	o := orm.NewOrm()
	var roles []*models.Role

	sql := "select role_id, role_code, name, role_mask from role"

	_, err := o.Raw(sql).QueryRows(&roles)
	if err != nil {
		return nil, err
	}

	for _, rr := range roles {
		if rr.RoleCode == "MDRWS" {
			roleList[ProjectAdmin] = rr
			continue
		}

		if rr.RoleCode == "RWS" {
			roleList[Developer] = rr
			continue
		}

		if rr.RoleCode == "RS" {
			roleList[Guest] = rr
			continue
		}
	}

	if roleList[r] == nil {
		return nil, fmt.Errorf("unsupported role type: %v", r)
	}

	return roleList[r], nil
}
