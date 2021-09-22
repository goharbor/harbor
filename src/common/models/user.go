// Copyright Project Harbor Authors
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

// User holds the details of a user.
type User struct {
	UserID          int    `json:"user_id"`
	Username        string `json:"username" sort:"default"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordVersion string `json:"password_version"`
	Realname        string `json:"realname"`
	Comment         string `json:"comment"`
	Deleted         bool   `json:"deleted"`
	Rolename        string `json:"role_name"`
	Role            int    `json:"role_id"`
	SysAdminFlag    bool   `json:"sysadmin_flag"`
	// AdminRoleInAuth to store the admin privilege granted by external authentication provider
	AdminRoleInAuth bool      `json:"admin_role_in_auth"`
	ResetUUID       string    `json:"reset_uuid"`
	Salt            string    `json:"-"`
	CreationTime    time.Time `json:"creation_time"`
	UpdateTime      time.Time `json:"update_time"`
	GroupIDs        []int     `json:"-"`
	OIDCUserMeta    *OIDCUser `json:"oidc_user_meta,omitempty"`
}

type Users []*User

// MapByUserID returns map which key is UserID of the user and value is the user itself
func (users Users) MapByUserID() map[int]*User {
	m := map[int]*User{}
	for _, user := range users {
		m[user.UserID] = user
	}
	return m
}
