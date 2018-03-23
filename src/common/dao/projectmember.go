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

package dao

import (
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/models"
)

// AddProjectMember inserts a record to table project_member
func AddProjectMember(projectID int64, userID int, role int, entityType string) (int, error) {
	o := GetOrmer()
	if !(entityType == common.UserMember || entityType == common.GroupMember) {
		entityType = common.UserMember
	}
	sql := `insert into project_member (project_id, entity_id , role, entity_type) values (?, ?, ?, ?)`
	_, err := o.Raw(sql, projectID, userID, role, entityType).Exec()
	if err != nil {
		return 0, err
	}
	var pmid int
	querySQL := `select id from project_member where project_id = ? and entity_id = ? and entity_type = ? limit 1`

	err = o.Raw(querySQL, projectID, userID, entityType).QueryRow(&pmid)
	if err != nil {
		return 0, err
	}
	return pmid, err
}

// UpdateProjectMember updates the record in table project_member
func UpdateProjectMember(projectID int64, userID int, role int, entityType string) error {
	o := GetOrmer()
	if !(entityType == common.UserMember || entityType == common.GroupMember) {
		entityType = common.UserMember
	}

	sql := `update project_member set role = ? where project_id = ? and entity_id = ?`

	_, err := o.Raw(sql, role, projectID, userID).Exec()

	return err
}

// DeleteProjectMember delete the record from table project_member
func DeleteProjectMember(projectID int64, userID int, entityType string) error {
	o := GetOrmer()

	if !(entityType == common.UserMember || entityType == common.GroupMember) {
		entityType = common.UserMember
	}

	sql := `delete from project_member where project_id = ? and entity_id = ? and entity_type = ?`

	if _, err := o.Raw(sql, projectID, userID, entityType).Exec(); err != nil {
		return err
	}

	return nil
}

// GetUserByProject gets all members of the project.
func GetUserByProject(projectID int64, queryUser models.User) ([]*models.UserMember, error) {
	o := GetOrmer()
	sql := `select pm.id as id, u.user_id, u.username, u.creation_time, u.update_time, r.name as rolename, 
		r.role_id as role, pm.entity_type as entity_type from user u join project_member pm 
		on pm.project_id = ? and u.user_id = pm.entity_id 
		join role r on pm.role = r.role_id where u.deleted = 0 and pm.entity_type = 'u' `

	queryParam := make([]interface{}, 1)
	queryParam = append(queryParam, projectID)

	if len(queryUser.Username) != 0 {
		sql += ` and u.username like ? `
		queryParam = append(queryParam, `%`+Escape(queryUser.Username)+`%`)
	}
	sql += ` order by u.username `

	members := []*models.UserMember{}
	_, err := o.Raw(sql, queryParam).QueryRows(&members)

	return members, err
}
