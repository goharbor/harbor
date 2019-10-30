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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota/driver"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/types"
)

var (
	quotaOrderMap = map[string]string{
		"creation_time":  "b.creation_time asc",
		"+creation_time": "b.creation_time asc",
		"-creation_time": "b.creation_time desc",
		"update_time":    "b.update_time asc",
		"+update_time":   "b.update_time asc",
		"-update_time":   "b.update_time desc",
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

// Quota quota mode for api
type Quota struct {
	ID           int64            `orm:"pk;auto;column(id)" json:"id"`
	Ref          driver.RefObject `json:"ref"`
	Reference    string           `orm:"column(reference)" json:"-"`
	ReferenceID  string           `orm:"column(reference_id)" json:"-"`
	Hard         string           `orm:"column(hard);type(jsonb)" json:"-"`
	Used         string           `orm:"column(used);type(jsonb)" json:"-"`
	CreationTime time.Time        `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time        `orm:"column(update_time);auto_now" json:"update_time"`
}

// MarshalJSON ...
func (q *Quota) MarshalJSON() ([]byte, error) {
	hard, err := types.NewResourceList(q.Hard)
	if err != nil {
		return nil, err
	}

	used, err := types.NewResourceList(q.Used)
	if err != nil {
		return nil, err
	}

	type Alias Quota
	return json.Marshal(&struct {
		*Alias
		Hard types.ResourceList `json:"hard"`
		Used types.ResourceList `json:"used"`
	}{
		Alias: (*Alias)(q),
		Hard:  hard,
		Used:  used,
	})
}

// ListQuotas returns quotas by query.
func ListQuotas(query ...*models.QuotaQuery) ([]*Quota, error) {
	condition, params := quotaQueryConditions(query...)

	sql := fmt.Sprintf(`
SELECT
  a.id,
  a.reference,
  a.reference_id,
  a.hard,
  b.used,
  b.creation_time,
  b.update_time
FROM
  quota AS a
  JOIN quota_usage AS b ON a.id = b.id %s`, condition)

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

	var quotas []*Quota
	if _, err := GetOrmer().Raw(sql, params).QueryRows(&quotas); err != nil {
		return nil, err
	}

	for _, quota := range quotas {
		d, ok := driver.Get(quota.Reference)
		if !ok {
			continue
		}

		ref, err := d.Load(quota.ReferenceID)
		if err != nil {
			log.Warning(fmt.Sprintf("Load quota reference object (%s, %s) failed: %v", quota.Reference, quota.ReferenceID, err))
			continue
		}

		quota.Ref = ref
	}

	return quotas, nil
}

// GetTotalOfQuotas returns total of quotas
func GetTotalOfQuotas(query ...*models.QuotaQuery) (int64, error) {
	condition, params := quotaQueryConditions(query...)
	sql := fmt.Sprintf("SELECT COUNT(1) FROM quota AS a JOIN quota_usage AS b ON a.id = b.id %s", condition)

	var count int64
	if err := GetOrmer().Raw(sql, params).QueryRow(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func quotaQueryConditions(query ...*models.QuotaQuery) (string, []interface{}) {
	params := []interface{}{}
	sql := ""
	if len(query) == 0 || query[0] == nil {
		return sql, params
	}

	sql += `WHERE 1=1 `

	q := query[0]
	if q.ID != 0 {
		sql += `AND a.id = ? `
		params = append(params, q.ID)
	}
	if q.Reference != "" {
		sql += `AND a.reference = ? `
		params = append(params, q.Reference)
	}
	if q.ReferenceID != "" {
		sql += `AND a.reference_id = ? `
		params = append(params, q.ReferenceID)
	}

	if len(q.ReferenceIDs) != 0 {
		sql += fmt.Sprintf(`AND a.reference_id IN (%s) `, ParamPlaceholderForIn(len(q.ReferenceIDs)))
		params = append(params, q.ReferenceIDs)
	}

	return sql, params
}

func castQuantity(field string) string {
	// cast -1 to max int64 when order by field
	return fmt.Sprintf("CAST( (CASE WHEN (%[1]s) IS NULL THEN '0' WHEN (%[1]s) = '-1' THEN '9223372036854775807' ELSE (%[1]s) END) AS BIGINT )", field)
}

func quotaOrderBy(query ...*models.QuotaQuery) string {
	orderBy := "b.creation_time DESC"

	if len(query) > 0 && query[0] != nil && query[0].Sort != "" {
		if val, ok := quotaOrderMap[query[0].Sort]; ok {
			orderBy = val
		} else {
			sort := query[0].Sort

			order := "ASC"
			if sort[0] == '-' {
				order = "DESC"
				sort = sort[1:]
			}

			prefixes := []string{"hard.", "used."}
			for _, prefix := range prefixes {
				if strings.HasPrefix(sort, prefix) {
					resource := strings.TrimPrefix(sort, prefix)
					if types.IsValidResource(types.ResourceName(resource)) {
						field := fmt.Sprintf("%s->>'%s'", strings.TrimSuffix(prefix, "."), resource)
						orderBy = fmt.Sprintf("(%s) %s", castQuantity(field), order)
						break
					}
				}
			}
		}
	}

	return orderBy
}
