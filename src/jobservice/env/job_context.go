// Copyright 2018 The Harbor Authors. All rights reserved.

package env

import (
	"context"

	"github.com/vmware/harbor/src/jobservice/logger"
)

//JobContext is combination of BaseContext and other job specified resources.
//JobContext will be the real execution context for one job.
type JobContext interface {
	//Build the context based on the parent context
	//
	//dep JobData : Dependencies for building the context, just in case that the build
	//function need some external info
	//
	//Returns:
	// new JobContext based on the parent one
	// error if meet any problems
	Build(dep JobData) (JobContext, error)

	//Get property from the context
	//
	//prop string : key of the context property
	//
	//Returns:
	//  The data of the specified context property if have
	//  bool to indicate if the property existing
	Get(prop string) (interface{}, bool)

	//SystemContext returns the system context
	//
	//Returns:
	//  context.Context
	SystemContext() context.Context

	//Checkin is bridge func for reporting detailed status
	//
	//status string : detailed status
	//
	//Returns:
	//  error if meet any problems
	Checkin(status string) error

	//OPCommand return the control operational command like stop/cancel if have
	//
	//Returns:
	//  op command if have
	//  flag to indicate if have command
	OPCommand() (string, bool)

	//Return the logger
	GetLogger() logger.Interface
}

//JobData defines job context dependencies.
type JobData struct {
	ID        string
	Name      string
	Args      map[string]interface{}
	ExtraData map[string]interface{}
}

//JobContextInitializer is a func to initialize the concrete job context
type JobContextInitializer func(ctx *Context) (JobContext, error)
