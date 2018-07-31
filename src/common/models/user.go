// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"time"
)

// UserTable is the name of table in DB that holds the user object
const UserTable = "harbor_user"

// User holds the details of a user.
type User struct {
	UserID   int    `orm:"pk;auto;column(user_id)" json:"user_id"`
	Username string `orm:"column(username)" json:"username"`
	Email    string `orm:"column(email)" json:"email"`
	Password string `orm:"column(password)" json:"password"`
	Realname string `orm:"column(realname)" json:"realname"`
	Comment  string `orm:"column(comment)" json:"comment"`
	Deleted  bool   `orm:"column(deleted)" json:"deleted"`
	Rolename string `orm:"-" json:"role_name"`
	//if this field is named as "RoleID", beego orm can not map role_id
	//to it.
	Role int `orm:"-" json:"role_id"`
	//	RoleList     []Role `json:"role_list"`
	HasAdminRole bool         `orm:"column(sysadmin_flag)" json:"has_admin_role"`
	ResetUUID    string       `orm:"column(reset_uuid)" json:"reset_uuid"`
	Salt         string       `orm:"column(salt)" json:"-"`
	CreationTime time.Time    `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time    `orm:"column(update_time);auto_now" json:"update_time"`
	GroupList    []*UserGroup `orm:"-" json:"-"`
}

// UserQuery ...
type UserQuery struct {
	Username   string
	Email      string
	Pagination *Pagination
}

// TableName ...
func (u *User) TableName() string {
	return UserTable
}
