//go:build !go1.18
// +build !go1.18

package redis

import (
	"errors"
	"reflect"
)

// fieldByIndexErr returns the nested field corresponding to index.
// It returns an error if evaluation requires stepping through a nil
// pointer, but panics if it must step through a field that
// is not a struct.
func fieldByIndexErr(v reflect.Value, index []int) (reflect.Value, error) {
	if len(index) == 1 {
		return v.Field(index[0]), nil
	}

	mustBe(v, reflect.Struct)
	for i, x := range index {
		if i > 0 {
			if v.Kind() == reflect.Ptr && v.Type().Elem().Kind() == reflect.Struct {
				if v.IsNil() {
					return reflect.Value{}, errors.New("reflect: indirection through nil pointer to embedded struct field " + v.Type().Elem().Name())
				}
				v = v.Elem()
			}
		}
		v = v.Field(x)
	}

	return v, nil
}
