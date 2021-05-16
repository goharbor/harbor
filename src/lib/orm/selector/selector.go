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

package selector

import (
	"context"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	sqlbuilder "github.com/huandu/go-sqlbuilder"
)

// Cond sqlbuilder.Cond wrapper with more methods to build conditions
type Cond struct {
	*sqlbuilder.Cond
}

// InsensitiveContains represents search filed by case-insensitive
func (c *Cond) InsensitiveContains(field string, value interface{}) string {
	str := strings.Replace(fmt.Sprintf("%v", value), `%`, `\%`, -1)
	arg := fmt.Sprintf("%%%s%%", str)

	return fmt.Sprintf("UPPER(%s::text) LIKE UPPER(%s)", sqlbuilder.Escape(field), c.Args.Add(arg))
}

// Exists represents EXISTS logic like "EXISTS expr".
func (c *Cond) Exists(expr string) string {
	return fmt.Sprintf("EXISTS (%s)", expr)
}

// NotExists represents NOT EXISTS logic like "NOT EXISTS expr".
func (c *Cond) NotExists(expr string) string {
	return fmt.Sprintf("NOT EXISTS (%s)", expr)
}

// Vars returns returns a placeholder for value slice.
func (c *Cond) Vars(value ...interface{}) string {
	vs := make([]string, 0, len(value))

	for _, v := range value {
		vs = append(vs, c.Args.Add(v))
	}

	return strings.Join(vs, ", ")
}

// Builder sqlbuilder.SelectBuilder wrapper
type Builder struct {
	*sqlbuilder.SelectBuilder
}

// Selector the selector to run SELECT query
type Selector struct {
	sb *sqlbuilder.SelectBuilder

	options []Option
}

// Build returns the sql, args from the sql builder
func (s *Selector) Build(options ...Option) (string, []interface{}, error) {
	builder := &Builder{s.sb}
	cond := &Cond{&s.sb.Cond}

	for _, o := range append(s.options, options...) {
		if err := o(builder, cond); err != nil {
			return "", nil, err
		}
	}

	sql, args := builder.Build()

	return sql, args, nil
}

// Count query the count
func (s *Selector) Count(ctx context.Context) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	sql, args, err := s.Build(Count())
	if err != nil {
		return 0, err
	}

	var count int64
	err = o.Raw(sql, args).QueryRow(&count)

	return count, err
}

// QueryRow query data and map to container
func (s *Selector) QueryRow(ctx context.Context, containers ...interface{}) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	sql, args, err := s.Build()
	if err != nil {
		return err
	}

	return o.Raw(sql, args...).QueryRow(containers)
}

// QueryRows query data rows and map to container
func (s *Selector) QueryRows(ctx context.Context, containers ...interface{}) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	sql, args, err := s.Build()
	if err != nil {
		return 0, err
	}

	return o.Raw(sql, args...).QueryRows(containers...)
}

// New returns a selector
func New(ctx context.Context, options ...Option) *Selector {
	sb := sqlbuilder.NewSelectBuilder()
	sb.SetFlavor(sqlbuilder.PostgreSQL)

	return &Selector{sb: sb, options: options}
}

// Option the option func for the selector
type Option func(builder *Builder, cond *Cond) error

// Select option applies SELECT T0.col1, T0.col2 ... for the model to the selector
func Select(model interface{}) Option {
	return func(builder *Builder, cond *Cond) error {
		meta := orm.ParseModel(model)

		var columns []string
		for _, column := range meta.Columns {
			columns = append(columns, fmt.Sprintf("T0.%s", column.Name))
		}

		builder.Select(columns...)

		return nil
	}
}

// Count option applies SELECT COUNT(*) to the selector
func Count() Option {
	return func(builder *Builder, cond *Cond) error {
		builder.Select("COUNT(*)")

		return nil
	}
}

// From option applies FROM for the model to the selector
func From(tableOrModel interface{}) Option {
	return func(builder *Builder, cond *Cond) error {
		meta := orm.ParseModel(tableOrModel)

		builder.From(fmt.Sprintf("%s T0", meta.TableName))

		return nil
	}
}

// Filter option applies Where conds from filters for the model to the selector
func Filter(model interface{}, filters map[string]interface{}) Option {
	return func(builder *Builder, cond *Cond) error {
		meta := orm.ParseModel(model)

		filterable := map[string]string{}
		for key := range filters {
			col := meta.GetColumn(key)
			if !col.IsFilterable() {
				continue
			}

			filterable[col.Name] = key
		}

		for _, col := range meta.Columns {
			key, ok := filterable[col.Name]
			if !ok {
				continue
			}

			field := fmt.Sprintf("T0.%s", col.Name)
			value, _ := filters[key]

			columnFilter(builder, cond, field, value)
		}

		return nil
	}
}

func columnFilter(builder *Builder, cond *Cond, field string, value interface{}) {
	if f, ok := value.(*q.FuzzyMatchValue); ok {
		builder.Where(cond.InsensitiveContains(field, f.Value))
		return
	}

	if r, ok := value.(*q.Range); ok {
		if r.Min != nil {
			builder.Where(cond.GreaterEqualThan(field, r.Min))
		}
		if r.Max != nil {
			builder.Where(cond.LessEqualThan(field, r.Max))
		}
		return
	}

	if ol, ok := value.(*q.OrList); ok && len(ol.Values) > 0 {
		builder.Where(cond.In(field, ol.Values...))
		return
	}

	if _, ok := value.(*q.AndList); ok {
		// do nothing as and list needs to be handled by the logic of DAO
		return
	}

	builder.Where(cond.Equal(field, value))
}

// Pagination option applies LIMIT OFFSET from pageNumber and pageSize to the selector
func Pagination(pageNumber, pageSize int64) Option {
	return func(builder *Builder, cond *Cond) error {
		if pageSize > 0 {
			builder.Limit(int(pageSize))

			if pageNumber > 0 {
				builder.Offset(int((pageNumber - 1) * pageSize))
			}
		}

		return nil
	}
}

// Sorts option applies ORDER BY from sorts to the selector
func Sorts(model interface{}, sorts []*q.Sort) Option {
	return func(builder *Builder, cond *Cond) error {
		meta := orm.ParseModel(model)

		getSortings := func(sorts []*q.Sort) []string {
			var sortings []string

			for _, sort := range sorts {
				col := meta.GetColumn(sort.Key)
				if !col.IsSortable() {
					continue
				}

				sorting := fmt.Sprintf("T0.%s", sqlbuilder.Escape(col.Name))
				if sort.DESC {
					sorting = sorting + " DESC"
				}

				sortings = append(sortings, sorting)
			}

			return sortings
		}

		sortings := getSortings(sorts)

		// if no sorts are specified, apply the default sort setting if exists
		if len(sortings) == 0 {
			sortings = getSortings(meta.DefaultSorts)
		}

		if len(sortings) > 0 {
			builder.OrderBy(sortings...)
		}

		return nil
	}
}
