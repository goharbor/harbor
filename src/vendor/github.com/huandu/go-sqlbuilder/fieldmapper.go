package sqlbuilder

import (
	"github.com/huandu/xstrings"
)

var (
	// DefaultFieldMapper is the default field name to table column name mapper func.
	// It's nil by default which means field name will be kept as it is.
	//
	// If a Struct has its own mapper func, the DefaultFieldMapper is ignored in this Struct.
	// Field tag has precedence over all kinds of field mapper functions.
	//
	// Field mapper is called only once on a Struct when the Struct is used to create builder for the first time.
	DefaultFieldMapper FieldMapperFunc
)

// FieldMapperFunc is a func to map struct field names to column names,
// which will be used in query as columns.
type FieldMapperFunc func(name string) string

// SnakeCaseMapper is a field mapper which can convert field name from CamelCase to snake_case.
//
// For instance, it will convert "MyField" to "my_field".
//
// SnakeCaseMapper uses package "xstrings" to do the conversion.
// See https://pkg.go.dev/github.com/huandu/xstrings#ToSnakeCase for conversion rules.
func SnakeCaseMapper(field string) string {
	return xstrings.ToSnakeCase(field)
}
