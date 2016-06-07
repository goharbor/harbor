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

package api

import (
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

func checkProjectPermission(userID int, projectID int64) bool {
	roles, err := listRoles(userID, projectID)
	if err != nil {
		log.Errorf("error occurred in getProjectPermission: %v", err)
		return false
	}
	return len(roles) > 0
}

func hasProjectAdminRole(userID int, projectID int64) bool {
	roles, err := listRoles(userID, projectID)
	if err != nil {
		log.Errorf("error occurred in getProjectPermission: %v", err)
		return false
	}

	for _, role := range roles {
		if role.RoleID == models.PROJECTADMIN {
			return true
		}
	}

	return false
}

//sysadmin has all privileges to all projects
func listRoles(userID int, projectID int64) ([]models.Role, error) {
	roles := make([]models.Role, 0, 1)
	isSysAdmin, err := dao.IsAdminRole(userID)
	if err != nil {
		return roles, err
	}
	if isSysAdmin {
		role, err := dao.GetRoleByID(models.PROJECTADMIN)
		if err != nil {
			return roles, err
		}
		roles = append(roles, *role)
		return roles, nil
	}

	rs, err := dao.GetUserProjectRoles(userID, projectID)
	if err != nil {
		return roles, err
	}
	roles = append(roles, rs...)
	return roles, nil
}

func checkUserExists(name string) int {
	u, err := dao.GetUser(models.User{Username: name})
	if err != nil {
		log.Errorf("Error occurred in GetUser, error: %v", err)
		return 0
	}
	if u != nil {
		return u.UserID
	}
	return 0
}
