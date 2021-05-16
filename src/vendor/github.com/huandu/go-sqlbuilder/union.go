// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"strconv"
	"strings"
)

const (
	unionDistinct = " UNION " // Default union type is DISTINCT.
	unionAll      = " UNION ALL "
)

const (
	unionMarkerInit injectionMarker = iota
	unionMarkerAfterUnion
	unionMarkerAfterOrderBy
	unionMarkerAfterLimit
)

// NewUnionBuilder creates a new UNION builder.
func NewUnionBuilder() *UnionBuilder {
	return DefaultFlavor.NewUnionBuilder()
}

func newUnionBuilder() *UnionBuilder {
	return &UnionBuilder{
		limit:  -1,
		offset: -1,

		args:      &Args{},
		injection: newInjection(),
	}
}

// UnionBuilder is a builder to build UNION.
type UnionBuilder struct {
	opt         string
	builders    []Builder
	orderByCols []string
	order       string
	limit       int
	offset      int

	args *Args

	injection *injection
	marker    injectionMarker
}

var _ Builder = new(UnionBuilder)

// Union unions all builders together using UNION operator.
func Union(builders ...Builder) *UnionBuilder {
	return DefaultFlavor.NewUnionBuilder().Union(builders...)
}

// Union unions all builders together using UNION operator.
func (ub *UnionBuilder) Union(builders ...Builder) *UnionBuilder {
	return ub.union(unionDistinct, builders...)
}

// UnionAll unions all builders together using UNION ALL operator.
func UnionAll(builders ...Builder) *UnionBuilder {
	return DefaultFlavor.NewUnionBuilder().UnionAll(builders...)
}

// UnionAll unions all builders together using UNION ALL operator.
func (ub *UnionBuilder) UnionAll(builders ...Builder) *UnionBuilder {
	return ub.union(unionAll, builders...)
}

func (ub *UnionBuilder) union(opt string, builders ...Builder) *UnionBuilder {
	ub.opt = opt
	ub.builders = builders
	ub.marker = unionMarkerAfterUnion
	return ub
}

// OrderBy sets columns of ORDER BY in SELECT.
func (ub *UnionBuilder) OrderBy(col ...string) *UnionBuilder {
	ub.orderByCols = col
	ub.marker = unionMarkerAfterOrderBy
	return ub
}

// Asc sets order of ORDER BY to ASC.
func (ub *UnionBuilder) Asc() *UnionBuilder {
	ub.order = "ASC"
	ub.marker = unionMarkerAfterOrderBy
	return ub
}

// Desc sets order of ORDER BY to DESC.
func (ub *UnionBuilder) Desc() *UnionBuilder {
	ub.order = "DESC"
	ub.marker = unionMarkerAfterOrderBy
	return ub
}

// Limit sets the LIMIT in SELECT.
func (ub *UnionBuilder) Limit(limit int) *UnionBuilder {
	ub.limit = limit
	ub.marker = unionMarkerAfterLimit
	return ub
}

// Offset sets the LIMIT offset in SELECT.
func (ub *UnionBuilder) Offset(offset int) *UnionBuilder {
	ub.offset = offset
	ub.marker = unionMarkerAfterLimit
	return ub
}

// String returns the compiled SELECT string.
func (ub *UnionBuilder) String() string {
	s, _ := ub.Build()
	return s
}

// Build returns compiled SELECT string and args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (ub *UnionBuilder) Build() (sql string, args []interface{}) {
	return ub.BuildWithFlavor(ub.args.Flavor)
}

// BuildWithFlavor returns compiled SELECT string and args with flavor and initial args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (ub *UnionBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
	buf := &bytes.Buffer{}
	ub.injection.WriteTo(buf, unionMarkerInit)

	if len(ub.builders) > 0 {
		needParen := flavor != SQLite

		if needParen {
			buf.WriteRune('(')
		}

		buf.WriteString(ub.Var(ub.builders[0]))

		if needParen {
			buf.WriteRune(')')
		}

		for _, b := range ub.builders[1:] {
			buf.WriteString(ub.opt)

			if needParen {
				buf.WriteRune('(')
			}

			buf.WriteString(ub.Var(b))

			if needParen {
				buf.WriteRune(')')
			}
		}
	}

	ub.injection.WriteTo(buf, unionMarkerAfterUnion)

	if len(ub.orderByCols) > 0 {
		buf.WriteString(" ORDER BY ")
		buf.WriteString(strings.Join(ub.orderByCols, ", "))

		if ub.order != "" {
			buf.WriteRune(' ')
			buf.WriteString(ub.order)
		}

		ub.injection.WriteTo(buf, unionMarkerAfterOrderBy)
	}

	if ub.limit >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.Itoa(ub.limit))

	}

	if MySQL == flavor && ub.limit >= 0 || PostgreSQL == flavor {
		if ub.offset >= 0 {
			buf.WriteString(" OFFSET ")
			buf.WriteString(strconv.Itoa(ub.offset))
		}
	}

	if ub.limit >= 0 {
		ub.injection.WriteTo(buf, unionMarkerAfterLimit)
	}

	return ub.args.CompileWithFlavor(buf.String(), flavor, initialArg...)
}

// SetFlavor sets the flavor of compiled sql.
func (ub *UnionBuilder) SetFlavor(flavor Flavor) (old Flavor) {
	old = ub.args.Flavor
	ub.args.Flavor = flavor
	return
}

// Var returns a placeholder for value.
func (ub *UnionBuilder) Var(arg interface{}) string {
	return ub.args.Add(arg)
}

// SQL adds an arbitrary sql to current position.
func (ub *UnionBuilder) SQL(sql string) *UnionBuilder {
	ub.injection.SQL(ub.marker, sql)
	return ub
}
