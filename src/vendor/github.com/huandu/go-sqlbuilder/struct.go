// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"bytes"
	"math"
	"reflect"
	"regexp"
	"strings"
)

var (
	// DBTag is the struct tag to describe the name for a field in struct.
	DBTag = "db"

	// FieldTag is the struct tag to describe the tag name for a field in struct.
	// Use "," to separate different tags.
	FieldTag = "fieldtag"

	// FieldOpt is the options for a struct field.
	// As db column can contain "," in theory, field options should be provided in a separated tag.
	FieldOpt = "fieldopt"
)

const (
	fieldOptWithQuote = "withquote"
	fieldOptOmitEmpty = "omitempty"

	optName   = "optName"
	optParams = "optParams"
)

var optRegex = regexp.MustCompile(`(?P<` + optName + `>\w+)(\((?P<` + optParams + `>.*)\))?`)

// Struct represents a struct type.
//
// All methods in Struct are thread-safe.
// We can define a global variable to hold a Struct and use it in any goroutine.
type Struct struct {
	Flavor Flavor

	structType         reflect.Type
	structFieldsParser structFieldsieldsParser
}

var emptyStruct Struct

// NewStruct analyzes type information in structValue
// and creates a new Struct with all structValue fields.
// If structValue is not a struct, NewStruct returns a dummy Sturct.
func NewStruct(structValue interface{}) *Struct {
	t := reflect.TypeOf(structValue)
	t = dereferencedType(t)

	if t.Kind() != reflect.Struct {
		return &emptyStruct
	}

	return &Struct{
		Flavor:             DefaultFlavor,
		structType:         t,
		structFieldsParser: makeDefaultFieldsParser(t),
	}
}

// For sets the default flavor of s and returns a shadow copy of s.
// The original s.Flavor is not changed.
func (s *Struct) For(flavor Flavor) *Struct {
	copy := *s
	copy.Flavor = flavor
	return &copy
}

// WithFieldMapper returns a new Struct based on s with custom field mapper.
// The original s is not changed.
func (s *Struct) WithFieldMapper(mapper FieldMapperFunc) *Struct {
	if s.structType == nil {
		return &emptyStruct
	}

	copy := *s
	copy.structFieldsParser = makeCustomFieldsParser(s.structType, mapper)
	return &copy
}

// SelectFrom creates a new `SelectBuilder` with table name.
// By default, all exported fields of the s are listed as columns in SELECT.
//
// Caller is responsible to set WHERE condition to find right record.
func (s *Struct) SelectFrom(table string) *SelectBuilder {
	return s.SelectFromForTag(table, "")
}

// SelectFromForTag creates a new `SelectBuilder` with table name for a specified tag.
// By default, all fields of the s tagged with tag are listed as columns in SELECT.
//
// Caller is responsible to set WHERE condition to find right record.
func (s *Struct) SelectFromForTag(table string, tag string) *SelectBuilder {
	sf := s.structFieldsParser()
	sb := s.Flavor.NewSelectBuilder()
	sb.From(table)

	if sf.taggedFields == nil {
		return sb
	}

	fields, ok := sf.taggedFields[tag]

	if ok {
		fields = s.quoteFields(sf, fields)

		buf := &bytes.Buffer{}
		cols := make([]string, 0, len(fields))

		for _, field := range fields {
			buf.WriteString(table)
			buf.WriteRune('.')
			buf.WriteString(field)
			cols = append(cols, buf.String())
			buf.Reset()
		}

		sb.Select(cols...)
	} else {
		sb.Select("*")
	}

	return sb
}

// Update creates a new `UpdateBuilder` with table name.
// By default, all exported fields of the s is assigned in UPDATE with the field values from value.
// If value's type is not the same as that of s, Update returns a dummy `UpdateBuilder` with table name.
//
// Caller is responsible to set WHERE condition to match right record.
func (s *Struct) Update(table string, value interface{}) *UpdateBuilder {
	return s.UpdateForTag(table, "", value)
}

// UpdateForTag creates a new `UpdateBuilder` with table name.
// By default, all fields of the s tagged with tag is assigned in UPDATE with the field values from value.
// If value's type is not the same as that of s, UpdateForTag returns a dummy `UpdateBuilder` with table name.
//
// Caller is responsible to set WHERE condition to match right record.
func (s *Struct) UpdateForTag(table string, tag string, value interface{}) *UpdateBuilder {
	sf := s.structFieldsParser()
	ub := s.Flavor.NewUpdateBuilder()
	ub.Update(table)

	if sf.taggedFields == nil {
		return ub
	}

	fields, ok := sf.taggedFields[tag]

	if !ok {
		return ub
	}

	v := reflect.ValueOf(value)
	v = dereferencedValue(v)

	if v.Type() != s.structType {
		return ub
	}

	quoted := s.quoteFields(sf, fields)
	assignments := make([]string, 0, len(fields))

	for i, f := range fields {
		name := sf.fieldAlias[f]
		val := v.FieldByName(name)

		if isEmptyValue(val) {
			if omitEmptyTagMap, ok := sf.omitEmptyFields[f]; ok {
				if omitEmptyTagMap.containsAny("", tag) {
					continue
				}
			}
		} else {
			val = dereferencedValue(val)
		}
		data := val.Interface()
		assignments = append(assignments, ub.Assign(quoted[i], data))
	}

	ub.Set(assignments...)
	return ub
}

// InsertInto creates a new `InsertBuilder` with table name using verb INSERT INTO.
// By default, all exported fields of s are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// InsertInto never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) InsertInto(table string, value ...interface{}) *InsertBuilder {
	return s.InsertIntoForTag(table, "", value...)
}

// InsertIgnoreInto creates a new `InsertBuilder` with table name using verb INSERT IGNORE INTO.
// By default, all exported fields of s are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// InsertIgnoreInto never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) InsertIgnoreInto(table string, value ...interface{}) *InsertBuilder {
	return s.InsertIgnoreIntoForTag(table, "", value...)
}

// ReplaceInto creates a new `InsertBuilder` with table name using verb REPLACE INTO.
// By default, all exported fields of s are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// ReplaceInto never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) ReplaceInto(table string, value ...interface{}) *InsertBuilder {
	return s.ReplaceIntoForTag(table, "", value...)
}

// buildColsAndValuesForTag uses ib to set exported fields tagged with tag as columns
// and add value as a list of values.
func (s *Struct) buildColsAndValuesForTag(ib *InsertBuilder, tag string, value ...interface{}) {
	sf := s.structFieldsParser()

	if sf.taggedFields == nil {
		return
	}

	fields, ok := sf.taggedFields[tag]

	if !ok {
		return
	}

	vs := make([]reflect.Value, 0, len(value))

	for _, item := range value {
		v := reflect.ValueOf(item)
		v = dereferencedValue(v)

		if v.Type() == s.structType {
			vs = append(vs, v)
		}
	}

	if len(vs) == 0 {
		return
	}
	cols := make([]string, 0, len(fields))
	values := make([][]interface{}, len(vs))

	for _, f := range fields {
		cols = append(cols, f)
		name := sf.fieldAlias[f]

		for i, v := range vs {
			data := v.FieldByName(name).Interface()
			values[i] = append(values[i], data)
		}
	}

	cols = s.quoteFields(sf, cols)
	ib.Cols(cols...)

	for _, value := range values {
		ib.Values(value...)
	}
}

// InsertIntoForTag creates a new `InsertBuilder` with table name using verb INSERT INTO.
// By default, exported fields tagged with tag are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// InsertIntoForTag never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) InsertIntoForTag(table string, tag string, value ...interface{}) *InsertBuilder {
	ib := s.Flavor.NewInsertBuilder()
	ib.InsertInto(table)

	s.buildColsAndValuesForTag(ib, tag, value...)
	return ib
}

// InsertIgnoreIntoForTag creates a new `InsertBuilder` with table name using verb INSERT IGNORE INTO.
// By default, exported fields tagged with tag are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// InsertIgnoreIntoForTag never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) InsertIgnoreIntoForTag(table string, tag string, value ...interface{}) *InsertBuilder {
	ib := s.Flavor.NewInsertBuilder()
	ib.InsertIgnoreInto(table)

	s.buildColsAndValuesForTag(ib, tag, value...)
	return ib
}

// ReplaceIntoForTag creates a new `InsertBuilder` with table name using verb REPLACE INTO.
// By default, exported fields tagged with tag are set as columns by calling `InsertBuilder#Cols`,
// and value is added as a list of values by calling `InsertBuilder#Values`.
//
// ReplaceIntoForTag never returns any error.
// If the type of any item in value is not expected, it will be ignored.
// If value is an empty slice, `InsertBuilder#Values` will not be called.
func (s *Struct) ReplaceIntoForTag(table string, tag string, value ...interface{}) *InsertBuilder {
	ib := s.Flavor.NewInsertBuilder()
	ib.ReplaceInto(table)

	s.buildColsAndValuesForTag(ib, tag, value...)
	return ib
}

// DeleteFrom creates a new `DeleteBuilder` with table name.
//
// Caller is responsible to set WHERE condition to match right record.
func (s *Struct) DeleteFrom(table string) *DeleteBuilder {
	db := s.Flavor.NewDeleteBuilder()
	db.DeleteFrom(table)
	return db
}

// Addr takes address of all exported fields of the s from the value.
// The returned result can be used in `Row#Scan` directly.
func (s *Struct) Addr(value interface{}) []interface{} {
	return s.AddrForTag("", value)
}

// AddrForTag takes address of all fields of the s tagged with tag from the value.
// The returned result can be used in `Row#Scan` directly.
//
// If tag is not defined in s in advance,
func (s *Struct) AddrForTag(tag string, value interface{}) []interface{} {
	sf := s.structFieldsParser()
	fields, ok := sf.taggedFields[tag]

	if !ok {
		return nil
	}

	return s.AddrWithCols(fields, value)
}

// AddrWithCols takes address of all columns defined in cols from the value.
// The returned result can be used in `Row#Scan` directly.
func (s *Struct) AddrWithCols(cols []string, value interface{}) []interface{} {
	sf := s.structFieldsParser()
	v := reflect.ValueOf(value)
	v = dereferencedValue(v)

	if v.Type() != s.structType {
		return nil
	}

	for _, c := range cols {
		if _, ok := sf.fieldAlias[c]; !ok {
			return nil
		}
	}

	addrs := make([]interface{}, 0, len(cols))

	for _, c := range cols {
		name := sf.fieldAlias[c]
		data := v.FieldByName(name).Addr().Interface()
		addrs = append(addrs, data)
	}

	return addrs
}

func (s *Struct) quoteFields(sf *structFields, fields []string) []string {
	// Try best not to allocate new slice.
	if len(sf.quotedFields) == 0 {
		return fields
	}

	needQuote := false

	for _, field := range fields {
		if _, ok := sf.quotedFields[field]; ok {
			needQuote = true
			break
		}
	}

	if !needQuote {
		return fields
	}

	quoted := make([]string, 0, len(fields))

	for _, field := range fields {
		if _, ok := sf.quotedFields[field]; ok {
			quoted = append(quoted, s.Flavor.Quote(field))
		} else {
			quoted = append(quoted, field)
		}
	}

	return quoted
}

func getOptMatchedMap(opt string) (res map[string]string) {
	res = map[string]string{}
	sm := optRegex.FindStringSubmatch(opt)
	for i, name := range optRegex.SubexpNames() {
		if name != "" {
			res[name] = sm[i]
		}
	}
	return
}

func getTagsFromOptParams(opts string) (tags []string) {
	tags = splitTokens(opts)
	if len(tags) == 0 {
		tags = append(tags, "")
	}
	return
}

func splitTokens(fieldtag string) (res []string) {
	res = strings.Split(fieldtag, ",")
	for i, v := range res {
		res[i] = strings.TrimSpace(v)
	}
	return
}

func dereferencedType(t reflect.Type) reflect.Type {
	for k := t.Kind(); k == reflect.Ptr || k == reflect.Interface; k = t.Kind() {
		t = t.Elem()
	}

	return t
}

func dereferencedValue(v reflect.Value) reflect.Value {
	for k := v.Kind(); k == reflect.Ptr || k == reflect.Interface; k = v.Kind() {
		v = v.Elem()
	}

	return v
}

// isEmptyValue checks if v is zero.
// Following code is borrowed from `IsZero` method in `reflect.Value` since Go 1.13.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return math.Float64bits(v.Float()) == 0
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		return math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isEmptyValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isEmptyValue(v.Field(i)) {
				return false
			}
		}
		return true
	}

	return false
}
