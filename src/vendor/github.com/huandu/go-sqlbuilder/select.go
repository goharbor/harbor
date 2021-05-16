// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const (
	selectMarkerInit injectionMarker = iota
	selectMarkerAfterSelect
	selectMarkerAfterFrom
	selectMarkerAfterJoin
	selectMarkerAfterWhere
	selectMarkerAfterGroupBy
	selectMarkerAfterOrderBy
	selectMarkerAfterLimit
	selectMarkerAfterFor
)

// JoinOption is the option in JOIN.
type JoinOption string

// Join options.
const (
	FullJoin       JoinOption = "FULL"
	FullOuterJoin  JoinOption = "FULL OUTER"
	InnerJoin      JoinOption = "INNER"
	LeftJoin       JoinOption = "LEFT"
	LeftOuterJoin  JoinOption = "LEFT OUTER"
	RightJoin      JoinOption = "RIGHT"
	RightOuterJoin JoinOption = "RIGHT OUTER"
)

// NewSelectBuilder creates a new SELECT builder.
func NewSelectBuilder() *SelectBuilder {
	return DefaultFlavor.NewSelectBuilder()
}

func newSelectBuilder() *SelectBuilder {
	args := &Args{}
	return &SelectBuilder{
		Cond: Cond{
			Args: args,
		},
		limit:     -1,
		offset:    -1,
		args:      args,
		injection: newInjection(),
	}
}

// SelectBuilder is a builder to build SELECT.
type SelectBuilder struct {
	Cond

	distinct    bool
	tables      []string
	selectCols  []string
	joinOptions []JoinOption
	joinTables  []string
	joinExprs   [][]string
	whereExprs  []string
	havingExprs []string
	groupByCols []string
	orderByCols []string
	order       string
	limit       int
	offset      int
	forWhat     string

	args *Args

	injection *injection
	marker    injectionMarker
}

var _ Builder = new(SelectBuilder)

// Select sets columns in SELECT.
func Select(col ...string) *SelectBuilder {
	return DefaultFlavor.NewSelectBuilder().Select(col...)
}

// Select sets columns in SELECT.
func (sb *SelectBuilder) Select(col ...string) *SelectBuilder {
	sb.selectCols = col
	sb.marker = selectMarkerAfterSelect
	return sb
}

// Distinct marks this SELECT as DISTINCT.
func (sb *SelectBuilder) Distinct() *SelectBuilder {
	sb.distinct = true
	sb.marker = selectMarkerAfterSelect
	return sb
}

// From sets table names in SELECT.
func (sb *SelectBuilder) From(table ...string) *SelectBuilder {
	sb.tables = table
	sb.marker = selectMarkerAfterFrom
	return sb
}

// Join sets expressions of JOIN in SELECT.
//
// It builds a JOIN expression like
//     JOIN table ON onExpr[0] AND onExpr[1] ...
func (sb *SelectBuilder) Join(table string, onExpr ...string) *SelectBuilder {
	sb.marker = selectMarkerAfterJoin
	return sb.JoinWithOption("", table, onExpr...)
}

// JoinWithOption sets expressions of JOIN with an option.
//
// It builds a JOIN expression like
//     option JOIN table ON onExpr[0] AND onExpr[1] ...
//
// Here is a list of supported options.
//     - FullJoin: FULL JOIN
//     - FullOuterJoin: FULL OUTER JOIN
//     - InnerJoin: INNER JOIN
//     - LeftJoin: LEFT JOIN
//     - LeftOuterJoin: LEFT OUTER JOIN
//     - RightJoin: RIGHT JOIN
//     - RightOuterJoin: RIGHT OUTER JOIN
func (sb *SelectBuilder) JoinWithOption(option JoinOption, table string, onExpr ...string) *SelectBuilder {
	sb.joinOptions = append(sb.joinOptions, option)
	sb.joinTables = append(sb.joinTables, table)
	sb.joinExprs = append(sb.joinExprs, onExpr)
	sb.marker = selectMarkerAfterJoin
	return sb
}

// Where sets expressions of WHERE in SELECT.
func (sb *SelectBuilder) Where(andExpr ...string) *SelectBuilder {
	sb.whereExprs = append(sb.whereExprs, andExpr...)
	sb.marker = selectMarkerAfterWhere
	return sb
}

// Having sets expressions of HAVING in SELECT.
func (sb *SelectBuilder) Having(andExpr ...string) *SelectBuilder {
	sb.havingExprs = append(sb.havingExprs, andExpr...)
	sb.marker = selectMarkerAfterGroupBy
	return sb
}

// GroupBy sets columns of GROUP BY in SELECT.
func (sb *SelectBuilder) GroupBy(col ...string) *SelectBuilder {
	sb.groupByCols = col
	sb.marker = selectMarkerAfterGroupBy
	return sb
}

// OrderBy sets columns of ORDER BY in SELECT.
func (sb *SelectBuilder) OrderBy(col ...string) *SelectBuilder {
	sb.orderByCols = col
	sb.marker = selectMarkerAfterOrderBy
	return sb
}

// Asc sets order of ORDER BY to ASC.
func (sb *SelectBuilder) Asc() *SelectBuilder {
	sb.order = "ASC"
	sb.marker = selectMarkerAfterOrderBy
	return sb
}

// Desc sets order of ORDER BY to DESC.
func (sb *SelectBuilder) Desc() *SelectBuilder {
	sb.order = "DESC"
	sb.marker = selectMarkerAfterOrderBy
	return sb
}

// Limit sets the LIMIT in SELECT.
func (sb *SelectBuilder) Limit(limit int) *SelectBuilder {
	sb.limit = limit
	sb.marker = selectMarkerAfterLimit
	return sb
}

// Offset sets the LIMIT offset in SELECT.
func (sb *SelectBuilder) Offset(offset int) *SelectBuilder {
	sb.offset = offset
	sb.marker = selectMarkerAfterLimit
	return sb
}

// ForUpdate adds FOR UPDATE at the end of SELECT statement.
func (sb *SelectBuilder) ForUpdate() *SelectBuilder {
	sb.forWhat = "UPDATE"
	sb.marker = selectMarkerAfterFor
	return sb
}

// ForShare adds FOR SHARE at the end of SELECT statement.
func (sb *SelectBuilder) ForShare() *SelectBuilder {
	sb.forWhat = "SHARE"
	sb.marker = selectMarkerAfterFor
	return sb
}

// As returns an AS expression.
func (sb *SelectBuilder) As(name, alias string) string {
	return fmt.Sprintf("%s AS %s", name, alias)
}

// BuilderAs returns an AS expression wrapping a complex SQL.
// According to SQL syntax, SQL built by builder is surrounded by parens.
func (sb *SelectBuilder) BuilderAs(builder Builder, alias string) string {
	return fmt.Sprintf("(%s) AS %s", sb.Var(builder), alias)
}

// String returns the compiled SELECT string.
func (sb *SelectBuilder) String() string {
	s, _ := sb.Build()
	return s
}

// Build returns compiled SELECT string and args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (sb *SelectBuilder) Build() (sql string, args []interface{}) {
	return sb.BuildWithFlavor(sb.args.Flavor)
}

// BuildWithFlavor returns compiled SELECT string and args with flavor and initial args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (sb *SelectBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
	buf := &bytes.Buffer{}
	sb.injection.WriteTo(buf, selectMarkerInit)
	buf.WriteString("SELECT ")

	if sb.distinct {
		buf.WriteString("DISTINCT ")
	}

	buf.WriteString(strings.Join(sb.selectCols, ", "))
	sb.injection.WriteTo(buf, selectMarkerAfterSelect)

	buf.WriteString(" FROM ")
	buf.WriteString(strings.Join(sb.tables, ", "))
	sb.injection.WriteTo(buf, selectMarkerAfterFrom)

	for i := range sb.joinTables {
		if option := sb.joinOptions[i]; option != "" {
			buf.WriteRune(' ')
			buf.WriteString(string(option))
		}

		buf.WriteString(" JOIN ")
		buf.WriteString(sb.joinTables[i])

		if exprs := sb.joinExprs[i]; len(exprs) > 0 {
			buf.WriteString(" ON ")
			buf.WriteString(strings.Join(sb.joinExprs[i], " AND "))
		}

	}

	if len(sb.joinTables) > 0 {
		sb.injection.WriteTo(buf, selectMarkerAfterJoin)
	}

	if len(sb.whereExprs) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(sb.whereExprs, " AND "))

		sb.injection.WriteTo(buf, selectMarkerAfterWhere)
	}

	if len(sb.groupByCols) > 0 {
		buf.WriteString(" GROUP BY ")
		buf.WriteString(strings.Join(sb.groupByCols, ", "))

		if len(sb.havingExprs) > 0 {
			buf.WriteString(" HAVING ")
			buf.WriteString(strings.Join(sb.havingExprs, " AND "))
		}

		sb.injection.WriteTo(buf, selectMarkerAfterGroupBy)
	}

	if len(sb.orderByCols) > 0 {
		buf.WriteString(" ORDER BY ")
		buf.WriteString(strings.Join(sb.orderByCols, ", "))

		if sb.order != "" {
			buf.WriteRune(' ')
			buf.WriteString(sb.order)
		}

		sb.injection.WriteTo(buf, selectMarkerAfterOrderBy)
	}

	if sb.limit >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.Itoa(sb.limit))
	}

	if MySQL == flavor && sb.limit >= 0 || PostgreSQL == flavor {
		if sb.offset >= 0 {
			buf.WriteString(" OFFSET ")
			buf.WriteString(strconv.Itoa(sb.offset))
		}
	}

	if sb.limit >= 0 {
		sb.injection.WriteTo(buf, selectMarkerAfterLimit)
	}

	if sb.forWhat != "" {
		buf.WriteString(" FOR ")
		buf.WriteString(sb.forWhat)

		sb.injection.WriteTo(buf, selectMarkerAfterFor)
	}

	return sb.args.CompileWithFlavor(buf.String(), flavor, initialArg...)
}

// SetFlavor sets the flavor of compiled sql.
func (sb *SelectBuilder) SetFlavor(flavor Flavor) (old Flavor) {
	old = sb.args.Flavor
	sb.args.Flavor = flavor
	return
}

// SQL adds an arbitrary sql to current position.
func (sb *SelectBuilder) SQL(sql string) *SelectBuilder {
	sb.injection.SQL(sb.marker, sql)
	return sb
}
