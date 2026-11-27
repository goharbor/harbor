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
	"sort"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/quota/types"
)

func (d *dao) AddUsage(ctx context.Context, reference, referenceID string, delta types.ResourceList, enforce bool) (applied bool, err error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return false, err
	}

	names := make([]string, 0, len(delta))
	for name := range delta {
		names = append(names, string(name))
	}
	sort.Strings(names)

	set := "u.used"
	var conds []string
	var setParams, condParams []any
	for _, name := range names {
		v := delta[types.ResourceName(name)]
		lit := orm.QuoteLiteral(name)
		set = fmt.Sprintf("jsonb_set(%s, ARRAY[%s], to_jsonb(COALESCE((u.used->>%s)::bigint, 0) + ?::bigint))", set, lit, lit)
		setParams = append(setParams, v)
		switch {
		case v < 0:
			// a refresh between a reservation and its rollback already dropped
			// the reservation from used; subtracting it again would erase
			// other usage, so leave the refreshed figure alone
			conds = append(conds, fmt.Sprintf("COALESCE((u.used->>%s)::bigint, 0) + ?::bigint >= 0", lit))
			condParams = append(condParams, v)
		case enforce && v > 0:
			// a resource missing from hard yields NULL and denies, as IsSafe does
			conds = append(conds, fmt.Sprintf("((q.hard->>%s)::bigint = %d OR COALESCE((u.used->>%s)::bigint, 0) + ?::bigint <= (q.hard->>%s)::bigint)",
				lit, types.UNLIMITED, lit, lit))
			condParams = append(condParams, v)
		}
	}

	sql := fmt.Sprintf(`UPDATE quota_usage u SET used = %s, version = CASE WHEN u.version = 9223372036854775807 THEN 0 ELSE u.version + 1 END, update_time = now()
FROM quota q WHERE q.id = u.id AND u.reference = ? AND u.reference_id = ?`, set)
	for _, c := range conds {
		sql += " AND " + c
	}

	params := append(setParams, reference, referenceID)
	params = append(params, condParams...)

	res, err := o.Raw(sql, params...).Exec()
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
