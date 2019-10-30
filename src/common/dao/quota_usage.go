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
	"github.com/goharbor/harbor/src/pkg/types"
)

var (
	quotaUsageOrderMap = map[string]string{
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

// AddQuotaUsage add quota usage to the database.
func AddQuotaUsage(quotaUsage models.QuotaUsage) (int64, error) {
	now := time.Now()
	quotaUsage.CreationTime = now
	quotaUsage.UpdateTime = now
	return GetOrmer().Insert(&quotaUsage)
}

// GetQuotaUsage returns quota usage by id.
func GetQuotaUsage(id int64) (*models.QuotaUsage, error) {
	q := models.QuotaUsage{ID: id}
	err := GetOrmer().Read(&q, "ID")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &q, err
}

// UpdateQuotaUsage update the quota usage.
func UpdateQuotaUsage(quotaUsage models.QuotaUsage) error {
	quotaUsage.UpdateTime = time.Now()
	_, err := GetOrmer().Update(&quotaUsage)
	return err
}

// ListQuotaUsages returns quota usages by query.
func ListQuotaUsages(query ...*models.QuotaUsageQuery) ([]*models.QuotaUsage, error) {
	condition, params := quotaUsageQueryConditions(query...)
	sql := fmt.Sprintf(`select * %s`, condition)

	orderBy := quotaUsageOrderBy(query...)
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

	var quotaUsages []*models.QuotaUsage
	if _, err := GetOrmer().Raw(sql, params).QueryRows(&quotaUsages); err != nil {
		return nil, err
	}

	return quotaUsages, nil
}

func quotaUsageQueryConditions(query ...*models.QuotaUsageQuery) (string, []interface{}) {
	params := []interface{}{}
	sql := `from quota_usage `
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
		sql += fmt.Sprintf(`and reference_id in (%s) `, ParamPlaceholderForIn(len(q.ReferenceIDs)))
		params = append(params, q.ReferenceIDs)
	}

	return sql, params
}

func quotaUsageOrderBy(query ...*models.QuotaUsageQuery) string {
	orderBy := ""

	if len(query) > 0 && query[0] != nil && query[0].Sort != "" {
		if val, ok := quotaUsageOrderMap[query[0].Sort]; ok {
			orderBy = val
		} else {
			sort := query[0].Sort

			order := "ASC"
			if sort[0] == '-' {
				order = "DESC"
				sort = sort[1:]
			}

			prefix := "used."
			if strings.HasPrefix(sort, prefix) {
				resource := strings.TrimPrefix(sort, prefix)
				if types.IsValidResource(types.ResourceName(resource)) {
					field := fmt.Sprintf("%s->>'%s'", strings.TrimSuffix(prefix, "."), resource)
					orderBy = fmt.Sprintf("(%s) %s", castQuantity(field), order)
				}
			}
		}
	}

	return orderBy
}
