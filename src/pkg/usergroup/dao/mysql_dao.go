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
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
)

// NewMysqlDao ...
func NewMysqlDao() DAO {
	return &mysqlDao{}
}

type mysqlDao struct {
	*dao
}

// Add - Add User Group
func (d *mysqlDao) Add(ctx context.Context, userGroup model.UserGroup) (int, error) {
	query := q.New(q.KeyWords{"GroupName": userGroup.GroupName, "GroupType": common.HTTPGroupType})
	userGroupList, err := d.Query(ctx, query)
	if err != nil {
		return 0, ErrGroupNameDup
	}
	if len(userGroupList) > 0 {
		return 0, ErrGroupNameDup
	}
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	var id int
	now := time.Now()
	sql := "insert into user_group (group_name, group_type, ldap_group_dn, creation_time, update_time) values (?, ?, ?, ?, ?)"

	res, err := o.Raw(sql, userGroup.GroupName, userGroup.GroupType, utils.TrimLower(userGroup.LdapGroupDN), now, now).Exec()
	if err != nil {
		return 0, err
	}
	insertID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	id = int(insertID)

	return id, nil
}
