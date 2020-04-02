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

package orm

import (
	"context"
	"reflect"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/q"
)

// WithFilters generates the query setter according to the query. "ignoredCols" is used to set the
// columns that will not be queried. Here pagination is not applied.
func WithFilters(ctx context.Context, model interface{}, query *q.Query, ignoredCols ...string) (orm.QuerySeter, error) {
	ormer, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	qs := ormer.QueryTable(model)
	if query == nil {
		return qs, nil
	}

	// the program will panic when querying the columns that doesn't exist
	// list the supported columns first to avoid the panic
	cols := listQueriableCols(model, ignoredCols...)
	for k, v := range query.Keywords {
		col := strings.SplitN(k, orm.ExprSep, 2)[0]
		if _, exist := cols[col]; !exist {
			continue
		}

		// fuzzy match
		f, ok := v.(*q.FuzzyMatchValue)
		if ok {
			qs = qs.Filter(k+"__icontains", f.Value)
			continue
		}

		// range
		r, ok := v.(*q.Range)
		if ok {
			if r.Min != nil {
				qs = qs.Filter(k+"__gte", r.Min)
			}
			if r.Max != nil {
				qs = qs.Filter(k+"__lte", r.Max)
			}
			continue
		}

		// or list
		ol, ok := v.(*q.OrList)
		if ok {
			if len(ol.Values) > 0 {
				qs = qs.Filter(k+"__in", ol.Values...)
			}
			continue
		}

		// and list
		_, ok = v.(*q.AndList)
		if ok {
			// do nothing as and list needs to be handled by the logic of DAO
			continue
		}

		// exact match
		qs = qs.Filter(k, v)
	}

	return qs, nil
}

// QuerySetter generates the query setter according to the query. "ignoredCols" is used to set the
// columns that will not be queried
func QuerySetter(ctx context.Context, model interface{}, query *q.Query, ignoredCols ...string) (orm.QuerySeter, error) {
	qs, err := WithFilters(ctx, model, query, ignoredCols...)
	if err != nil {
		return nil, err
	}

	if query != nil && query.PageSize > 0 {
		qs = qs.Limit(query.PageSize)
		if query.PageNumber > 0 {
			qs = qs.Offset(query.PageSize * (query.PageNumber - 1))
		}
	}

	return qs, nil
}

// list the columns that can be queried
// e.g. for the following model the columns that can be queried are:
// "Field2", "customized_field2", "Field3" and "field3"
// type model struct{
//   Field1 string `orm:"-"`
//   Field2 string `orm:"column(customized_field2)"`
//   Field3 string
// }
//
// set "ignoredCols" to ignore the specified columns
func listQueriableCols(model interface{}, ignoredCols ...string) map[string]struct{} {
	if model == nil {
		return nil
	}
	ignored := map[string]struct{}{}
	for _, ig := range ignoredCols {
		ignored[ig] = struct{}{}
	}
	cols := map[string]struct{}{}
	t := reflect.Indirect(reflect.ValueOf(model)).Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		orm := field.Tag.Get("orm")
		if orm == "-" {
			continue
		}
		colName := ""
		for _, str := range strings.Split(orm, ";") {
			if strings.HasPrefix(str, "column") {
				str = strings.TrimPrefix(str, "column(")
				str = strings.TrimSuffix(str, ")")
				if len(str) > 0 {
					colName = str
					break
				}
			}
		}
		if len(colName) == 0 {
			// TODO convert the field.Name to snake_case
		}
		if _, exist := ignored[colName]; exist {
			continue
		}
		if _, exist := ignored[field.Name]; exist {
			continue
		}
		if len(colName) != 0 {
			cols[colName] = struct{}{}
		}
		cols[field.Name] = struct{}{}
	}
	return cols
}

// ParamPlaceholderForIn returns a string that contains placeholders for sql keyword "in"
// e.g. n=3, returns "?,?,?"
func ParamPlaceholderForIn(n int) string {
	placeholders := []string{}
	for i := 0; i < n; i++ {
		placeholders = append(placeholders, "?")
	}
	return strings.Join(placeholders, ",")
}

// Escape special characters
func Escape(str string) string {
	return dao.Escape(str)
}
