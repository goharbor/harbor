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

	"github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

var (
	// cache the parsed models
	cache = sync.Map{}
)

type key struct {
	Name       string
	Filterable bool
	FilterFunc func(context.Context, orm.QuerySeter, string, interface{}) orm.QuerySeter
	Sortable   bool
}

type metadata struct {
	Keys         map[string]*key
	DefaultSorts []*q.Sort
}

func (m *metadata) Filterable(key string) (*key, bool) {
	k, exist := m.Keys[key]
	return k, exist
}

func (m *metadata) Sortable(key string) bool {
	k, exist := m.Keys[key]
	if !exist {
		return false
	}
	return k.Sortable
}

// parse the definition of the provided model(fields/methods/annotations) and return the parsed metadata
func parseModel(model interface{}) *metadata {
	// pointer type
	ptr := reflect.TypeOf(model)
	// struct type
	t := ptr.Elem()

	// get the metadata from cache first
	fullName := getFullName(t)
	cacheMetadata, exist := cache.Load(fullName)
	if exist {
		return cacheMetadata.(*metadata)
	}

	// pointer value
	v := reflect.ValueOf(model)
	metadata := &metadata{
		Keys: map[string]*key{},
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

		metadata.Keys[field.Name] = &key{
			Name:       field.Name,
			Filterable: filterable,
			Sortable:   sortable,
		}
		metadata.Keys[column] = &key{
			Name:       column,
			Filterable: filterable,
			Sortable:   sortable,
		}
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
		filterFunc, ok := methodValue.Interface().(func(context.Context, orm.QuerySeter, string, interface{}) orm.QuerySeter)
		if !ok {
			continue
		}
		field := strings.TrimPrefix(methodName, "FilterBy")
		metadata.Keys[field] = &key{
			Name:       field,
			Filterable: true,
			FilterFunc: filterFunc,
		}
		snakeCaseField := snakeCase(field)
		metadata.Keys[snakeCaseField] = &key{
			Name:       snakeCaseField,
			Filterable: true,
			FilterFunc: filterFunc,
		}
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
