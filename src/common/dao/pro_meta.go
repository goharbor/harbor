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
	"fmt"
	"strings"
	"time"

	"github.com/vmware/harbor/src/common/models"
)

// Using raw sql instead of CRUD objects as beego does not support composite primary key

// AddProjectMetadata adds metadata for a project
func AddProjectMetadata(meta *models.ProjectMetadata) error {
	now := time.Now()
	sql := `insert into project_metadata
				(project_id, name, value, creation_time, update_time, deleted)
				 values (?, ?, ?, ?, ?, 0)`
	_, err := GetOrmer().Raw(sql, meta.ProjectID, meta.Name, meta.Value,
		now, now).Exec()
	return err
}

// DeleteProjectMetadata deleted metadata of a project. If name is absent
// all metadatas will be deleted, otherwise only the metadatas specified
// by name will be deleted
func DeleteProjectMetadata(projectID int64, name ...string) error {
	params := make([]interface{}, 1)
	sql := `update project_metadata 
			set deleted = 1 
			where project_id = ?`
	params = append(params, projectID)

	if len(name) > 0 {
		sql += fmt.Sprintf(` and name in ( %s )`, paramPlaceholder(len(name)))
		params = append(params, name)
	}

	_, err := GetOrmer().Raw(sql, params).Exec()
	return err
}

// UpdateProjectMetadata updates metadata of a project
func UpdateProjectMetadata(meta *models.ProjectMetadata) error {
	sql := `update project_metadata 
				set value = ?, update_time = ? 
				where project_id = ? and name = ? and deleted = 0`
	_, err := GetOrmer().Raw(sql, meta.Value, time.Now(), meta.ProjectID,
		meta.Name).Exec()
	return err
}

// GetProjectMetadata returns the metadata of a project. If name is absent
// all metadatas will be returned, otherwise only the metadatas specified
// by name will be returned
func GetProjectMetadata(projectID int64, name ...string) ([]*models.ProjectMetadata, error) {
	proMetas := []*models.ProjectMetadata{}
	params := make([]interface{}, 1)

	sql := `select * from project_metadata 
				where project_id = ? and deleted = 0`
	params = append(params, projectID)

	if len(name) > 0 {
		sql += fmt.Sprintf(` and name in ( %s )`, paramPlaceholder(len(name)))
		params = append(params, name)
	}

	_, err := GetOrmer().Raw(sql, params).QueryRows(&proMetas)
	return proMetas, err
}

func paramPlaceholder(n int) string {
	placeholders := []string{}
	for i := 0; i < n; i++ {
		placeholders = append(placeholders, "?")
	}
	return strings.Join(placeholders, ",")
}
