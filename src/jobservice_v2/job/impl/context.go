// Copyright 2018 The Harbor Authors. All rights reserved.

package impl

import (
	"context"
	"reflect"

	hlog "github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice_v2/env"
	"github.com/vmware/harbor/src/jobservice_v2/job"
)

//Context ...
type Context struct {
	//System context
	sysContext context.Context

	//Logger for job
	logger *hlog.Logger

	//op command func
	opCommandFunc job.CheckOPCmdFunc
}

//NewContext ...
func NewContext(sysCtx context.Context) *Context {
	return &Context{
		sysContext: sysCtx,
	}
}

//InitDao ...
func (c *Context) InitDao() error {
	return nil
}

//Build implements the same method in env.JobContext interface
//This func will build the job execution context before running
func (c *Context) Build(dep env.JobData) (env.JobContext, error) {
	jContext := &Context{
		sysContext: c.sysContext,
	}

	//TODO:Init logger here
	if opCommandFunc, ok := dep.ExtraData["opCommandFunc"]; ok {
		if reflect.TypeOf(opCommandFunc).Kind() == reflect.Func {
			if funcRef, ok := opCommandFunc.(job.CheckOPCmdFunc); ok {
				jContext.opCommandFunc = funcRef
			}
		}
	}

	return jContext, nil
}

//Get implements the same method in env.JobContext interface
func (c *Context) Get(prop string) interface{} {
	return nil
}

//SystemContext implements the same method in env.JobContext interface
func (c *Context) SystemContext() context.Context {
	return c.sysContext
}

//Checkin is bridge func for reporting detailed status
func (c *Context) Checkin(status string) error {
	return nil
}

//OPCommand return the control operational command like stop/cancel if have
func (c *Context) OPCommand() (string, bool) {
	if c.opCommandFunc != nil {
		return c.opCommandFunc()
	}

	return "", false
}
