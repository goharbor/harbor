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
	exist, err := dao.IsAdminRole(userID)
	if err != nil {
		log.Errorf("Error occurred in IsAdminRole, error: %v", err)
		return false
	}
	if exist {
		return true
	}
	roleList, err := dao.GetUserProjectRoles(userID, projectID)
	if err != nil {
		log.Errorf("Error occurred in GetUserProjectRoles, error: %v", err)
		return false
	}
	return len(roleList) > 0
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
