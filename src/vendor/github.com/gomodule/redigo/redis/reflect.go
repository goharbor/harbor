package redis

import (
	"reflect"
	"runtime"
)

// methodName returns the name of the calling method,
// assumed to be two stack frames above.
func methodName() string {
	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown method"
	}
	return f.Name()
}

// mustBe panics if f's kind is not expected.
func mustBe(v reflect.Value, expected reflect.Kind) {
	if v.Kind() != expected {
		panic(&reflect.ValueError{Method: methodName(), Kind: v.Kind()})
	}
}

// fieldByIndexCreate returns the nested field corresponding
// to index creating elements that are nil when stepping through.
// It panics if v is not a struct.
func fieldByIndexCreate(v reflect.Value, index []int) reflect.Value {
	if len(index) == 1 {
		return v.Field(index[0])
	}

	mustBe(v, reflect.Struct)
	for i, x := range index {
		if i > 0 {
			if v.Kind() == reflect.Ptr && v.Type().Elem().Kind() == reflect.Struct {
				if v.IsNil() {
					v.Set(reflect.New(v.Type().Elem()))
				}
				v = v.Elem()
			}
		}
		v = v.Field(x)
	}

	return v
}
