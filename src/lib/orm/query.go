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
	"sync"
	"unicode"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
)

// NewCondition alias function of orm.NewCondition
var NewCondition = orm.NewCondition

// Condition alias to orm.Condition
type Condition = orm.Condition

// Params alias to orm.Params
type Params = orm.Params

// ParamsList alias to orm.ParamsList
type ParamsList = orm.ParamsList

// QuerySeter alias to orm.QuerySeter
type QuerySeter = orm.QuerySeter

// Escape special characters
func Escape(str string) string {
	return dao.Escape(str)
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

// QuerySetter generates the query setter according to the query. "ignoredCols" is used to set the columns that will not be queried.
// Currently, it supports two ways to generate the query setter, the first one is to generate by the fields of the model,
// and the second one is to generate by the methods their name begins with `FilterBy` of the model.
// e.g. for the following model the queriable fields are  :
// "Field2", "customized_field2", "Field3", "field3" and "Field4" (or "field4").
// type Foo struct{
//   Field1 string `orm:"-"`
//   Field2 string `orm:"column(customized_field2)"`
//   Field3 string
// }
//
// func (f *Foo) FilterByField4(ctx context.Context, qs orm.QuerySeter, key string, value interface{}) orm.QuerySeter {
//   // The value is the raw value of key in q.Query
//	 return qs
// }
func QuerySetter(ctx context.Context, model interface{}, query *q.Query, ignoredCols ...string) (orm.QuerySeter, error) {
	val := reflect.ValueOf(model)
	if val.Kind() != reflect.Ptr {
		return nil, errors.Errorf("<orm.QuerySetter> cannot use non-ptr model struct `%s`", getFullName(reflect.Indirect(val).Type()))
	}

	ormer, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	qs := ormer.QueryTable(model)
	if query == nil {
		return qs, nil
	}

	ignored := map[string]bool{}
	for _, col := range ignoredCols {
		ignored[col] = true
	}

	columns := queriableColumns(model)
	methods := queriableMethods(model)
	for k, v := range query.Keywords {
		field := strings.SplitN(k, orm.ExprSep, 2)[0]
		if ignored[field] {
			continue
		}

		if columns[field] {
			qs = queryByColumn(qs, k, v)
		} else if method, ok := methods[snakeCase(field)]; ok {
			qs = queryByMethod(ctx, qs, k, v, method, val)
		}
	}

	if query.PageSize > 0 {
		qs = qs.Limit(query.PageSize)
		if query.PageNumber > 0 {
			qs = qs.Offset(query.PageSize * (query.PageNumber - 1))
		}
	}
	return qs, nil
}

// get reflect.Type name with package path.
func getFullName(typ reflect.Type) string {
	return typ.PkgPath() + "." + typ.Name()
}

// convert string to snake case
func snakeCase(str string) string {
	delim := '_'

	runes := []rune(str)

	var out []rune
	for i := 0; i < len(runes); i++ {
		if i > 0 &&
			(unicode.IsUpper(runes[i])) &&
			((i+1 < len(runes) && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, delim)
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

func queryByColumn(qs orm.QuerySeter, key string, value interface{}) orm.QuerySeter {
	// fuzzy match
	if f, ok := value.(*q.FuzzyMatchValue); ok {
		return qs.Filter(key+"__icontains", f.Value)
	}

	// range
	if r, ok := value.(*q.Range); ok {
		if r.Min != nil {
			qs = qs.Filter(key+"__gte", r.Min)
		}
		if r.Max != nil {
			qs = qs.Filter(key+"__lte", r.Max)
		}
		return qs
	}

	// or list
	if ol, ok := value.(*q.OrList); ok {
		if len(ol.Values) > 0 {
			qs = qs.Filter(key+"__in", ol.Values...)
		}
		return qs
	}

	// and list
	if _, ok := value.(*q.AndList); ok {
		// do nothing as and list needs to be handled by the logic of DAO
		return qs
	}

	// exact match
	return qs.Filter(key, value)
}

func queryByMethod(ctx context.Context, qs orm.QuerySeter, key string, value interface{}, methodName string, reflectVal reflect.Value) orm.QuerySeter {
	if mv := reflectVal.MethodByName(methodName); mv.IsValid() {
		switch method := mv.Interface().(type) {
		case func(context.Context, orm.QuerySeter, string, interface{}) orm.QuerySeter:
			return method(ctx, qs, key, value)
		default:
			return qs
		}
	}

	return qs
}

var (
	cache = sync.Map{}
)

// get model fields which are columns in orm
// e.g. for the following model the columns that can be queried are:
// "Field2", "customized_field2", "Field3" and "field3"
// type model struct{
//   Field1 string `orm:"-"`
//   Field2 string `orm:"column(customized_field2)"`
//   Field3 string
// }
func queriableColumns(model interface{}) map[string]bool {
	typ := reflect.Indirect(reflect.ValueOf(model)).Type()

	key := getFullName(typ) + "-columns"
	value, ok := cache.Load(key)
	if ok {
		return value.(map[string]bool)
	}

	cols := map[string]bool{}
	defer func() {
		cache.Store(key, cols)
	}()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
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

		if colName == "" {
			colName = snakeCase(field.Name)
		}

		cols[colName] = true
		cols[field.Name] = true
	}
	return cols
}

// get model methods which begin with `FilterBy`
func queriableMethods(model interface{}) map[string]string {
	val := reflect.ValueOf(model)

	key := getFullName(reflect.Indirect(val).Type()) + "-methods"
	value, ok := cache.Load(key)
	if ok {
		return value.(map[string]string)
	}

	methods := map[string]string{}
	defer func() {
		cache.Store(key, methods)
	}()

	prefix := "FilterBy"
	typ := val.Type()
	for i := 0; i < typ.NumMethod(); i++ {
		name := typ.Method(i).Name

		if !strings.HasPrefix(name, prefix) {
			continue
		}

		field := snakeCase(strings.TrimPrefix(name, prefix))
		if field != "" {
			methods[field] = name
		}
	}

	return methods
}
