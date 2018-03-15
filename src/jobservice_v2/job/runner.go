// Copyright 2018 The Harbor Authors. All rights reserved.

package job

import "reflect"

//Wrap returns a new (job.)Interface based on the wrapped job handler reference.
func Wrap(j interface{}) Interface {
	theType := reflect.TypeOf(j)

	if theType.Kind() == reflect.Ptr {
		theType = theType.Elem()
	}

	//Crate new
	v := reflect.New(theType).Elem()
	return v.Addr().Interface().(Interface)
}
