// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"strconv"
	"strings"
)

const (
	deleteMarkerInit injectionMarker = iota
	deleteMarkerAfterDeleteFrom
	deleteMarkerAfterWhere
	deleteMarkerAfterOrderBy
	deleteMarkerAfterLimit
)

// NewDeleteBuilder creates a new DELETE builder.
func NewDeleteBuilder() *DeleteBuilder {
	return DefaultFlavor.NewDeleteBuilder()
}

func newDeleteBuilder() *DeleteBuilder {
	args := &Args{}
	return &DeleteBuilder{
		Cond: Cond{
			Args: args,
		},
		limit:     -1,
		args:      args,
		injection: newInjection(),
	}
}

// DeleteBuilder is a builder to build DELETE.
type DeleteBuilder struct {
	Cond

	table       string
	whereExprs  []string
	orderByCols []string
	order       string
	limit       int

	args *Args

	injection *injection
	marker    injectionMarker
}

var _ Builder = new(DeleteBuilder)

// DeleteFrom sets table name in DELETE.
func DeleteFrom(table string) *DeleteBuilder {
	return DefaultFlavor.NewDeleteBuilder().DeleteFrom(table)
}

// DeleteFrom sets table name in DELETE.
func (db *DeleteBuilder) DeleteFrom(table string) *DeleteBuilder {
	db.table = Escape(table)
	db.marker = deleteMarkerAfterDeleteFrom
	return db
}

// Where sets expressions of WHERE in DELETE.
func (db *DeleteBuilder) Where(andExpr ...string) *DeleteBuilder {
	db.whereExprs = append(db.whereExprs, andExpr...)
	db.marker = deleteMarkerAfterWhere
	return db
}

// OrderBy sets columns of ORDER BY in DELETE.
func (db *DeleteBuilder) OrderBy(col ...string) *DeleteBuilder {
	db.orderByCols = col
	db.marker = deleteMarkerAfterOrderBy
	return db
}

// Asc sets order of ORDER BY to ASC.
func (db *DeleteBuilder) Asc() *DeleteBuilder {
	db.order = "ASC"
	db.marker = deleteMarkerAfterOrderBy
	return db
}

// Desc sets order of ORDER BY to DESC.
func (db *DeleteBuilder) Desc() *DeleteBuilder {
	db.order = "DESC"
	db.marker = deleteMarkerAfterOrderBy
	return db
}

// Limit sets the LIMIT in DELETE.
func (db *DeleteBuilder) Limit(limit int) *DeleteBuilder {
	db.limit = limit
	db.marker = deleteMarkerAfterLimit
	return db
}

// String returns the compiled DELETE string.
func (db *DeleteBuilder) String() string {
	s, _ := db.Build()
	return s
}

// Build returns compiled DELETE string and args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (db *DeleteBuilder) Build() (sql string, args []interface{}) {
	return db.BuildWithFlavor(db.args.Flavor)
}

// BuildWithFlavor returns compiled DELETE string and args with flavor and initial args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (db *DeleteBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
	buf := &bytes.Buffer{}
	db.injection.WriteTo(buf, deleteMarkerInit)
	buf.WriteString("DELETE FROM ")
	buf.WriteString(db.table)
	db.injection.WriteTo(buf, deleteMarkerAfterDeleteFrom)

	if len(db.whereExprs) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(db.whereExprs, " AND "))

		db.injection.WriteTo(buf, deleteMarkerAfterWhere)
	}

	if len(db.orderByCols) > 0 {
		buf.WriteString(" ORDER BY ")
		buf.WriteString(strings.Join(db.orderByCols, ", "))

		if db.order != "" {
			buf.WriteRune(' ')
			buf.WriteString(db.order)
		}

		db.injection.WriteTo(buf, deleteMarkerAfterOrderBy)
	}

	if db.limit >= 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.Itoa(db.limit))

		db.injection.WriteTo(buf, deleteMarkerAfterLimit)
	}

	return db.args.CompileWithFlavor(buf.String(), flavor, initialArg...)
}

// SetFlavor sets the flavor of compiled sql.
func (db *DeleteBuilder) SetFlavor(flavor Flavor) (old Flavor) {
	old = db.args.Flavor
	db.args.Flavor = flavor
	return
}

// SQL adds an arbitrary sql to current position.
func (db *DeleteBuilder) SQL(sql string) *DeleteBuilder {
	db.injection.SQL(db.marker, sql)
	return db
}
