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

package models

const (
	//PROJECTADMIN project administrator
	PROJECTADMIN = 1
	//DEVELOPER developer
	DEVELOPER = 2
	//GUEST guest
	GUEST = 3
)

// Role holds the details of a role.
type Role struct {
	RoleID   int    `orm:"column(role_id)" json:"role_id"`
	RoleCode string `orm:"column(role_code)" json:"role_code"`
	Name     string `orm:"column(name)" json:"role_name"`

	RoleMask int `orm:"role_mask" json:"role_mask"`
}
