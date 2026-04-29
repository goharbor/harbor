//go:build go1.18
// +build go1.18

package redis

import (
	"reflect"
)

// fieldByIndexErr returns the nested field corresponding to index.
// It returns an error if evaluation requires stepping through a nil
// pointer, but panics if it must step through a field that
// is not a struct.
func fieldByIndexErr(v reflect.Value, index []int) (reflect.Value, error) {
	return v.FieldByIndexErr(index)
}
