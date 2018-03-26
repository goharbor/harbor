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

// Member holds the details of a member.
type Member struct {
	ID         int    `orm:"pk;column(id)" json:"id"`
	ProjectID  int64  `orm:"column(project_id)" json:"project_id"`
	Entityname string `orm:"column(entity_name)" json:"entity_name"`
	Rolename   string `json:"role_name"`
	Role       int    `json:"role_id"`
	EntityID   int    `orm:"column(entity_id)" json:"entity_id"`
	EntityType string `orm:"column(entity_type)" json:"entity_type"`
}

// UserMember ...
type UserMember struct {
	ID       int    `orm:"pk;column(user_id)" json:"user_id"`
	Username string `json:"username"`
	Rolename string `json:"role_name"`
	Role     int    `json:"role_id"`
}

// MemberReq -  Create Project Member Request
type MemberReq struct {
	ProjectID   int64     `json:"project_id"`
	Role        int       `json:"role_id,omitempty"`
	MemberUser  User      `json:"member_user,omitempty"`
	MemberGroup UserGroup `json:"member_group,omitempty"`
}
