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

package dao

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/quota/models"
)

// NewMysqlDao returns an instance of the mysql DAO
func NewMysqlDao() DAO {
	return &mysqlDao{}
}

type mysqlDao struct {
	*dao
}

func (d *mysqlDao) List(ctx context.Context, query *q.Query) ([]*models.Quota, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	condition, params := listConditions(query)

	sql := fmt.Sprintf(`
SELECT
  a.id,
  a.reference,
  a.reference_id,
  a.hard,
  a.version as hard_version,
  b.used,
  b.version as used_version,
  b.creation_time,
  b.update_time
FROM
  quota AS a
  JOIN quota_usage AS b ON a.id = b.id %s`, condition)

	orderBy := listOrderByForMysql(query)
	if orderBy != "" {
		sql += ` order by ` + orderBy
	}

	if query != nil {
		page, size := query.PageNumber, query.PageSize
		if size > 0 {
			sql += ` limit ?`
			params = append(params, size)
			if page > 0 {
				sql += ` offset ?`
				params = append(params, size*(page-1))
			}
		}
	}

	var quotas []*models.Quota
	if _, err := o.Raw(sql, params).QueryRows(&quotas); err != nil {
		return nil, err
	}

	return quotas, nil
}
