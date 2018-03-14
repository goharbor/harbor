// Copyright 2018 The Harbor Authors. All rights reserved.

package env

import "context"

//JobContext is combination of BaseContext and other job specified resources.
//JobContext will be the real execution context for one job.
type JobContext interface {
	//Build the context
	//
	//dep JobData : Dependencies for building the context, just in case that the build
	//function need some external info
	//
	//Returns:
	// error if meet any problems
	Build(dep JobData) error

	//Get property from the context
	//
	//prop string : key of the context property
	//
	//Returns:
	//  The data of the specified context property
	Get(prop string) interface{}

	//SystemContext returns the system context
	//
	//Returns:
	//  context.Context
	SystemContext() context.Context
}

//JobData defines job context dependencies.
type JobData struct {
	ID   string
	Name string
	Args map[string]interface{}
}
