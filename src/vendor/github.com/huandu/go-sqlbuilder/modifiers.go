// Copyright 2018 Huan Du. All rights reserved.
// Licensed under the MIT license that can be found in the LICENSE file.

package sqlbuilder

import (
	"reflect"
	"strings"
)

// Escape replaces `$` with `$$` in ident.
func Escape(ident string) string {
	return strings.Replace(ident, "$", "$$", -1)
}

// EscapeAll replaces `$` with `$$` in all strings of ident.
func EscapeAll(ident ...string) []string {
	escaped := make([]string, 0, len(ident))

	for _, i := range ident {
		escaped = append(escaped, Escape(i))
	}

	return escaped
}

// Flatten recursively extracts values in slices and returns
// a flattened []interface{} with all values.
// If slices is not a slice, return `[]interface{}{slices}`.
func Flatten(slices interface{}) (flattened []interface{}) {
	v := reflect.ValueOf(slices)
	slices, flattened = flatten(v)

	if slices != nil {
		return []interface{}{slices}
	}

	return flattened
}

func flatten(v reflect.Value) (elem interface{}, flattened []interface{}) {
	k := v.Kind()

	for k == reflect.Interface {
		v = v.Elem()
		k = v.Kind()
	}

	if k != reflect.Slice && k != reflect.Array {
		return v.Interface(), nil
	}

	for i, l := 0, v.Len(); i < l; i++ {
		e, f := flatten(v.Index(i))

		if e == nil {
			flattened = append(flattened, f...)
		} else {
			flattened = append(flattened, e)
		}
	}

	return
}

type rawArgs struct {
	expr string
}

// Raw marks the expr as a raw value which will not be added to args.
func Raw(expr string) interface{} {
	return rawArgs{expr}
}

type listArgs struct {
	args []interface{}
}

// List marks arg as a list of data.
// If arg is `[]int{1, 2, 3}`, it will be compiled to `?, ?, ?` with args `[1 2 3]`.
func List(arg interface{}) interface{} {
	return listArgs{Flatten(arg)}
}

type namedArgs struct {
	name string
	arg  interface{}
}

// Named creates a named argument.
// Unlike `sql.Named`, this named argument works only with `Build` or `BuildNamed` for convenience
// and will be replaced to a `?` after `Compile`.
func Named(name string, arg interface{}) interface{} {
	return namedArgs{
		name: name,
		arg:  arg,
	}
}
