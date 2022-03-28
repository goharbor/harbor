//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package dao

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/member/models"
)

type mysqlDao struct {
	*dao
}

// NewMysqlDao ...
func NewMysqlDao() DAO {
	return &mysqlDao{}
}

func (d *mysqlDao) AddProjectMember(ctx context.Context, member models.Member) (int, error) {
	log.Debugf("Adding project member %+v", member)
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	if member.EntityID <= 0 {
		return 0, fmt.Errorf("invalid entity_id, member: %+v", member)
	}

	if member.ProjectID <= 0 {
		return 0, fmt.Errorf("invalid project_id, member: %+v", member)
	}

	delSQL := "delete from project_member where project_id = ? and entity_id = ? and entity_type = ? "
	_, err = o.Raw(delSQL, member.ProjectID, member.EntityID, member.EntityType).Exec()
	if err != nil {
		return 0, err
	}

	var pmid int

	sql := "insert into project_member (project_id, entity_id , role, entity_type) values (?, ?, ?, ?)"
	res, err := o.Raw(sql, member.ProjectID, member.EntityID, member.Role, member.EntityType).Exec()
	if err != nil {
		return 0, err
	}
	insertID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	pmid = int(insertID)

	return pmid, err
}
