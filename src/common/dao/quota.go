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
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
)

var (
	quotaOrderMap = map[string]string{
		"id":             "id asc",
		"+id":            "id asc",
		"-id":            "id desc",
		"creation_time":  "creation_time asc",
		"+creation_time": "creation_time asc",
		"-creation_time": "creation_time desc",
		"update_time":    "update_time asc",
		"+update_time":   "update_time asc",
		"-update_time":   "update_time desc",
	}
)

// AddQuota add quota to the database.
func AddQuota(quota models.Quota) (int64, error) {
	now := time.Now()
	quota.CreationTime = now
	quota.UpdateTime = now
	return GetOrmer().Insert(&quota)
}

// GetQuota returns quota by id.
func GetQuota(id int64) (*models.Quota, error) {
	q := models.Quota{ID: id}
	err := GetOrmer().Read(&q, "ID")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &q, err
}

// UpdateQuota update the quota.
func UpdateQuota(quota models.Quota) error {
	quota.UpdateTime = time.Now()
	_, err := GetOrmer().Update(&quota)
	return err
}

// ListQuotas returns quotas by query.
func ListQuotas(query ...*models.QuotaQuery) ([]*models.Quota, error) {
	condition, params := quotaQueryConditions(query...)
	sql := fmt.Sprintf(`select * %s`, condition)

	orderBy := quotaOrderBy(query...)
	if orderBy != "" {
		sql += ` order by ` + orderBy
	}

	if len(query) > 0 && query[0] != nil {
		page, size := query[0].Page, query[0].Size
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
	if _, err := GetOrmer().Raw(sql, params).QueryRows(&quotas); err != nil {
		return nil, err
	}

	return quotas, nil
}

func quotaQueryConditions(query ...*models.QuotaQuery) (string, []interface{}) {
	params := []interface{}{}
	sql := `from quota `
	if len(query) == 0 || query[0] == nil {
		return sql, params
	}

	sql += `where 1=1 `

	q := query[0]
	if q.Reference != "" {
		sql += `and reference = ? `
		params = append(params, q.Reference)
	}
	if q.ReferenceID != "" {
		sql += `and reference_id = ? `
		params = append(params, q.ReferenceID)
	}
	if len(q.ReferenceIDs) != 0 {
		sql += fmt.Sprintf(`and reference_id in (%s) `, paramPlaceholder(len(q.ReferenceIDs)))
		params = append(params, q.ReferenceIDs)
	}

	return sql, params
}

func quotaOrderBy(query ...*models.QuotaQuery) string {
	orderBy := ""

	if len(query) > 0 && query[0] != nil && query[0].Sort != "" {
		if val, ok := quotaOrderMap[query[0].Sort]; ok {
			orderBy = val
		} else {
			sort := query[0].Sort

			order := "asc"
			if sort[0] == '-' {
				order = "desc"
				sort = sort[1:]
			}

			prefix := "hard."
			if strings.HasPrefix(sort, prefix) {
				orderBy = fmt.Sprintf("hard->>'%s' %s", strings.TrimPrefix(sort, prefix), order)
			}
		}
	}

	return orderBy
}
