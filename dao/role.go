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

	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/models"
)

// GetUserProjectRoles returns roles that the user has according to the project.
func GetUserProjectRoles(userID int, projectID int64) ([]models.Role, error) {

	o := orm.NewOrm()

	sql := `select *
		from role
		where role_id = 
			(
				select role
				from project_member
				where project_id = ? and user_id = ?
			)`

	var roleList []models.Role
	_, err := o.Raw(sql, projectID, userID).QueryRows(&roleList)

	if err != nil {
		return nil, err
	}
	return roleList, nil
}

// IsAdminRole returns whether the user is admin.
func IsAdminRole(userIDOrUsername interface{}) (bool, error) {
	u := models.User{}

	switch v := userIDOrUsername.(type) {
	case int:
		u.UserID = v
	case string:
		u.Username = v
	default:
		return false, fmt.Errorf("invalid parameter, only int and string are supported: %v", userIDOrUsername)
	}

	if u.UserID == NonExistUserID && len(u.Username) == 0 {
		return false, nil
	}

	user, err := GetUser(u)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	return user.HasAdminRole == 1, nil
}
