// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"errors"
	"fmt"
)

// Supported flavors.
const (
	invalidFlavor Flavor = iota

	MySQL
	PostgreSQL
	SQLite
)

var (
	// DefaultFlavor is the default flavor for all builders.
	DefaultFlavor = MySQL
)

var (
	// ErrInterpolateNotImplemented means the method or feature is not implemented right now.
	ErrInterpolateNotImplemented = errors.New("go-sqlbuilder: interpolation for this flavor is not implemented")

	// ErrInterpolateMissingArgs means there are some args missing in query, so it's not possible to
	// prepare a query with such args.
	ErrInterpolateMissingArgs = errors.New("go-sqlbuilder: not enough args when interpolating")

	// ErrInterpolateUnsupportedArgs means that some types of the args are not supported.
	ErrInterpolateUnsupportedArgs = errors.New("go-sqlbuilder: unsupported args when interpolating")
)

// Flavor is the flag to control the format of compiled sql.
type Flavor int

// String returns the name of f.
func (f Flavor) String() string {
	switch f {
	case MySQL:
		return "MySQL"
	case PostgreSQL:
		return "PostgreSQL"
	case SQLite:
		return "SQLite"
	}

	return "<invalid>"
}

// Interpolate parses sql returned by `Args#Compile` or `Builder`,
// and interpolate args to replace placeholders in the sql.
//
// If there are some args missing in sql, e.g. the number of placeholders are larger than len(args),
// returns ErrMissingArgs error.
func (f Flavor) Interpolate(sql string, args []interface{}) (string, error) {
	switch f {
	case MySQL:
		return mysqlInterpolate(sql, args...)
	case PostgreSQL:
		return postgresqlInterpolate(sql, args...)
	case SQLite:
		return sqliteInterpolate(sql, args...)
	}

	return "", ErrInterpolateNotImplemented
}

// NewCreateTableBuilder creates a new CREATE TABLE builder with flavor.
func (f Flavor) NewCreateTableBuilder() *CreateTableBuilder {
	b := newCreateTableBuilder()
	b.SetFlavor(f)
	return b
}

// NewDeleteBuilder creates a new DELETE builder with flavor.
func (f Flavor) NewDeleteBuilder() *DeleteBuilder {
	b := newDeleteBuilder()
	b.SetFlavor(f)
	return b
}

// NewInsertBuilder creates a new INSERT builder with flavor.
func (f Flavor) NewInsertBuilder() *InsertBuilder {
	b := newInsertBuilder()
	b.SetFlavor(f)
	return b
}

// NewSelectBuilder creates a new SELECT builder with flavor.
func (f Flavor) NewSelectBuilder() *SelectBuilder {
	b := newSelectBuilder()
	b.SetFlavor(f)
	return b
}

// NewUpdateBuilder creates a new UPDATE builder with flavor.
func (f Flavor) NewUpdateBuilder() *UpdateBuilder {
	b := newUpdateBuilder()
	b.SetFlavor(f)
	return b
}

// NewUnionBuilder creates a new UNION builder with flavor.
func (f Flavor) NewUnionBuilder() *UnionBuilder {
	b := newUnionBuilder()
	b.SetFlavor(f)
	return b
}

// Quote adds quote for name to make sure the name can be used safely
// as table name or field name.
//
// * For MySQL, use back quote (`) to quote name;
// * For PostgreSQL and SQLite, use double quote (") to quote name.
func (f Flavor) Quote(name string) string {
	switch f {
	case MySQL:
		return fmt.Sprintf("`%s`", name)
	case PostgreSQL, SQLite:
		return fmt.Sprintf(`"%s"`, name)
	}

	return name
}
