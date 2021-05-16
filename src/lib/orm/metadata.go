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
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

var (
	// cache the parsed models
	cache = sync.Map{}
)

// Column the columns of the model
type Column struct {
	Name       string // the column name in db
	filterable bool   // True when the column is searchable
	sortable   bool   // True when the column is orderable
}

// IsFilterable returns true when column is filterable
func (col *Column) IsFilterable() bool {
	return col != nil && col.filterable
}

// IsSortable returns true when column is sortable
func (col *Column) IsSortable() bool {
	return col != nil && col.sortable
}

// FilterFunc type alias to filter funcs for orm.QuerySeter
type FilterFunc = func(context.Context, orm.QuerySeter, string, interface{}) orm.QuerySeter

// Metadata metadata of model
type Metadata struct {
	TableName    string
	DefaultSorts []*q.Sort
	Columns      []*Column // ordered columns

	columnIndexes map[string]*Column // help to find column by column name or field name
	filterFuncs   map[string]FilterFunc
}

// AddColumn add column to the metadata
func (m *Metadata) AddColumn(name string, filterable bool, sortable bool, alias ...string) {
	col := &Column{
		Name:       name,
		filterable: filterable,
		sortable:   sortable,
	}
	m.Columns = append(m.Columns, col)

	m.columnIndexes[name] = col
	for _, a := range alias {
		m.columnIndexes[a] = col
	}
}

// GetColumn get column of the model by db column name or filed name in model
func (m *Metadata) GetColumn(columnOrField string) *Column {
	for _, key := range []string{columnOrField, snakeCase(columnOrField)} {
		if col, ok := m.columnIndexes[key]; ok {
			return col
		}
	}

	return nil
}

// GetFilterFunc get the filter func for the key
func (m *Metadata) GetFilterFunc(key string) (FilterFunc, bool) {
	for _, key := range []string{key, snakeCase(key)} {
		if f, ok := m.filterFuncs[key]; ok {
			return f, true
		}
	}

	return nil, false
}

// ParseModel parse the definition of the provided model(fields/methods/annotations) and return the parsed metadata
func ParseModel(model interface{}) *Metadata {
	// pointer type
	ptr := reflect.TypeOf(model)
	// struct type
	t := ptr.Elem()

	// get the metadata from cache first
	fullName := getFullName(t)
	cacheMetadata, exist := cache.Load(fullName)
	if exist {
		return cacheMetadata.(*Metadata)
	}

	// pointer value
	v := reflect.ValueOf(model)
	metadata := &Metadata{
		TableName:     getTableName(v),
		columnIndexes: map[string]*Column{},
		filterFuncs:   map[string]FilterFunc{},
	}
	// parse fields of the provided model
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		orm := field.Tag.Get("orm")
		// isn't the database column, skip
		if orm == "-" {
			continue
		}

		filterable := parseFilterable(field)
		defaultSort, sortable := parseSortable(field)
		column := parseColumn(field)

		metadata.AddColumn(column, filterable, sortable, field.Name)

		metadata.filterFuncs[column] = columnFilterFunc
		metadata.filterFuncs[field.Name] = columnFilterFunc

		if defaultSort != nil {
			metadata.DefaultSorts = []*q.Sort{defaultSort}
		}
	}

	// parse filter methods of the provided model
	for i := 0; i < ptr.NumMethod(); i++ {
		methodName := ptr.Method(i).Name
		if !strings.HasPrefix(methodName, "FilterBy") {
			continue
		}
		methodValue := v.MethodByName(methodName)
		if !methodValue.IsValid() {
			continue
		}

		filterFunc, ok := methodValue.Interface().(FilterFunc)
		if !ok {
			fmt.Printf("%s method is not filter func", methodName)
			continue
		}
		field := strings.TrimPrefix(methodName, "FilterBy")

		metadata.filterFuncs[field] = filterFunc
		metadata.filterFuncs[snakeCase(field)] = filterFunc
	}

	// parse default sorts method
	methodValue := v.MethodByName("GetDefaultSorts")
	if methodValue.IsValid() {
		values := methodValue.Call(nil)
		if len(values) == 1 {
			if sorts, ok := values[0].Interface().([]*q.Sort); ok && len(sorts) > 0 {
				metadata.DefaultSorts = sorts
			}
		}
	}

	cache.Store(fullName, metadata)
	return metadata
}

// parseFilterable parses whether the field is filterable according to the field annotation
// For the following struct definition, "Field1" isn't filterable and "Field2" is filterable
// type Model struct {
//	 Field1 string `filter:"false"`
//	 Field2 string
// }
func parseFilterable(field reflect.StructField) bool {
	return field.Tag.Get("filter") != "false"
}

// parseSortable parses whether the field is sortable according to the field annotation
// If the field is sortable and is also specified as the default sort, return a q.Sort model as well
// For the following struct definition, "Field1" isn't sortable and "Field2", "Field2", "Field4", "Field5" are all sortable
// type Model struct {
//	 Field1 string `sort:"false"`
//	 Field2 string `sort:"true;default"`
//	 Field3 string `sort:"true;default:desc"`
//	 Field4 string `sort:"default"`
//   Field5 string
// }
func parseSortable(field reflect.StructField) (*q.Sort, bool) {
	var defaultSort *q.Sort
	for _, item := range strings.Split(field.Tag.Get("sort"), ";") {
		// isn't sortable, return directly
		if item == "false" {
			return nil, false
		}
		if !strings.HasPrefix(item, "default") {
			continue
		}
		defaultSort = &q.Sort{
			Key:  field.Name,
			DESC: false,
		}
		if strings.TrimPrefix(item, "default") == ":desc" {
			defaultSort.DESC = true
		}
	}
	return defaultSort, true
}

// parseColumn parses the column name according to the field annotation
// type Model struct {
//	 Field1 string `orm:"column(customized_field1)"`
//	 Field2 string
// }
// It returns "customized_field1" for "Field1" and returns "field2" for "Field2"
func parseColumn(field reflect.StructField) string {
	column := ""
	for _, item := range strings.Split(field.Tag.Get("orm"), ";") {
		if !strings.HasPrefix(item, "column") {
			continue
		}
		item = strings.TrimPrefix(item, "column(")
		item = strings.TrimSuffix(item, ")")
		column = item
		break
	}
	if len(column) == 0 {
		column = snakeCase(field.Name)
	}
	return column
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

// snakeString and getTableName copy from the beego orm
// snake string, XxYy to xx_yy , XxYY to xx_y_y
func snakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

func getTableName(val reflect.Value) string {
	if fun := val.MethodByName("TableName"); fun.IsValid() {
		vals := fun.Call([]reflect.Value{})
		// has return and the first val is string
		if len(vals) > 0 && vals[0].Kind() == reflect.String {
			return vals[0].String()
		}
	}
	return snakeString(reflect.Indirect(val).Type().Name())
}

func columnFilterFunc(ctx context.Context, qs orm.QuerySeter, key string, value interface{}) orm.QuerySeter {
	// fuzzy match
	if f, ok := value.(*q.FuzzyMatchValue); ok {
		return qs.Filter(key+"__icontains", Escape(f.Value))
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
