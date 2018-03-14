package impl

import (
	"context"

	hlog "github.com/vmware/harbor/src/common/utils/log"
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

//Build implements the same method in env.JobContext interface
func (c *Context) Build() error {
	return nil
}

//Get implements the same method in env.JobContext interface
func (c *Context) Get(prop string) interface{} {
	return nil
}

//SystemContext implements the same method in env.JobContext interface
func (c *Context) SystemContext() context.Context {
	return c.sysContext
}
