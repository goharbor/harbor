// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"strings"
)

const (
	createTableMarkerInit injectionMarker = iota
	createTableMarkerAfterCreate
	createTableMarkerAfterDefine
	createTableMarkerAfterOption
)

// NewCreateTableBuilder creates a new CREATE TABLE builder.
func NewCreateTableBuilder() *CreateTableBuilder {
	return DefaultFlavor.NewCreateTableBuilder()
}

func newCreateTableBuilder() *CreateTableBuilder {
	args := &Args{}
	return &CreateTableBuilder{
		verb:      "CREATE TABLE",
		args:      args,
		injection: newInjection(),
		marker:    createTableMarkerInit,
	}
}

// CreateTableBuilder is a builder to build CREATE TABLE.
type CreateTableBuilder struct {
	verb        string
	ifNotExists bool
	table       string
	defs        [][]string
	options     [][]string

	args *Args

	injection *injection
	marker    injectionMarker
}

var _ Builder = new(CreateTableBuilder)

// CreateTable sets the table name in CREATE TABLE.
func CreateTable(table string) *CreateTableBuilder {
	return DefaultFlavor.NewCreateTableBuilder().CreateTable(table)
}

// CreateTable sets the table name in CREATE TABLE.
func (ctb *CreateTableBuilder) CreateTable(table string) *CreateTableBuilder {
	ctb.table = Escape(table)
	ctb.marker = createTableMarkerAfterCreate
	return ctb
}

// CreateTempTable sets the table name and changes the verb of ctb to CREATE TEMPORARY TABLE.
func (ctb *CreateTableBuilder) CreateTempTable(table string) *CreateTableBuilder {
	ctb.verb = "CREATE TEMPORARY TABLE"
	ctb.table = Escape(table)
	ctb.marker = createTableMarkerAfterCreate
	return ctb
}

// IfNotExists adds IF NOT EXISTS before table name in CREATE TABLE.
func (ctb *CreateTableBuilder) IfNotExists() *CreateTableBuilder {
	ctb.ifNotExists = true
	return ctb
}

// Define adds definition of a column or index in CREATE TABLE.
func (ctb *CreateTableBuilder) Define(def ...string) *CreateTableBuilder {
	ctb.defs = append(ctb.defs, def)
	ctb.marker = createTableMarkerAfterDefine
	return ctb
}

// Option adds a table option in CREATE TABLE.
func (ctb *CreateTableBuilder) Option(opt ...string) *CreateTableBuilder {
	ctb.options = append(ctb.options, opt)
	ctb.marker = createTableMarkerAfterOption
	return ctb
}

// String returns the compiled INSERT string.
func (ctb *CreateTableBuilder) String() string {
	s, _ := ctb.Build()
	return s
}

// Build returns compiled CREATE TABLE string and args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (ctb *CreateTableBuilder) Build() (sql string, args []interface{}) {
	return ctb.BuildWithFlavor(ctb.args.Flavor)
}

// BuildWithFlavor returns compiled CREATE TABLE string and args with flavor and initial args.
// They can be used in `DB#Query` of package `database/sql` directly.
func (ctb *CreateTableBuilder) BuildWithFlavor(flavor Flavor, initialArg ...interface{}) (sql string, args []interface{}) {
	buf := &bytes.Buffer{}
	ctb.injection.WriteTo(buf, createTableMarkerInit)
	buf.WriteString(ctb.verb)

	if ctb.ifNotExists {
		buf.WriteString(" IF NOT EXISTS")
	}

	buf.WriteRune(' ')
	buf.WriteString(ctb.table)
	ctb.injection.WriteTo(buf, createTableMarkerAfterCreate)

	if len(ctb.defs) > 0 {
		buf.WriteString(" (")

		defs := make([]string, 0, len(ctb.defs))

		for _, def := range ctb.defs {
			defs = append(defs, strings.Join(def, " "))
		}

		buf.WriteString(strings.Join(defs, ", "))
		buf.WriteRune(')')

		ctb.injection.WriteTo(buf, createTableMarkerAfterDefine)
	}

	if len(ctb.options) > 0 {
		buf.WriteRune(' ')

		opts := make([]string, 0, len(ctb.options))

		for _, opt := range ctb.options {
			opts = append(opts, strings.Join(opt, " "))
		}

		buf.WriteString(strings.Join(opts, ", "))
		ctb.injection.WriteTo(buf, createTableMarkerAfterOption)
	}

	return ctb.args.CompileWithFlavor(buf.String(), flavor, initialArg...)
}

// SetFlavor sets the flavor of compiled sql.
func (ctb *CreateTableBuilder) SetFlavor(flavor Flavor) (old Flavor) {
	old = ctb.args.Flavor
	ctb.args.Flavor = flavor
	return
}

// Var returns a placeholder for value.
func (ctb *CreateTableBuilder) Var(arg interface{}) string {
	return ctb.args.Add(arg)
}

// SQL adds an arbitrary sql to current position.
func (ctb *CreateTableBuilder) SQL(sql string) *CreateTableBuilder {
	ctb.injection.SQL(ctb.marker, sql)
	return ctb
}
