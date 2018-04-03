// Copyright 2018 The Harbor Authors. All rights reserved.

package pool

import (
	"reflect"

	"github.com/vmware/harbor/src/jobservice/job"
)

//Wrap returns a new job.Interface based on the wrapped job handler reference.
func Wrap(j interface{}) job.Interface {
	theType := reflect.TypeOf(j)

	if theType.Kind() == reflect.Ptr {
		theType = theType.Elem()
	}

	//Crate new
	v := reflect.New(theType).Elem()
	return v.Addr().Interface().(job.Interface)
}
