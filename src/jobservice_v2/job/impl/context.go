// Copyright 2018 The Harbor Authors. All rights reserved.

package impl

import (
	"context"

	hlog "github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice_v2/env"
)

//Context ...
type Context struct {
	//System context
	sysContext context.Context

	//Logger for job
	logger *hlog.Logger
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
	return &Context{
		sysContext: c.sysContext,
	}, nil
}

//Get implements the same method in env.JobContext interface
func (c *Context) Get(prop string) interface{} {
	return nil
}

//SystemContext implements the same method in env.JobContext interface
func (c *Context) SystemContext() context.Context {
	return c.sysContext
}
