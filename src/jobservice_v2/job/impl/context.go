// Copyright 2018 The Harbor Authors. All rights reserved.

package impl

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/vmware/harbor/src/jobservice_v2/config"
	"github.com/vmware/harbor/src/jobservice_v2/env"
	"github.com/vmware/harbor/src/jobservice_v2/job"
	jlogger "github.com/vmware/harbor/src/jobservice_v2/job/impl/logger"
	"github.com/vmware/harbor/src/jobservice_v2/logger"
)

//Context ...
type Context struct {
	//System context
	sysContext context.Context

	//Logger for job
	logger logger.Interface

	//op command func
	opCommandFunc job.CheckOPCmdFunc

	//checkin func
	checkInFunc job.CheckInFunc
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

	//Init logger here
	logPath := fmt.Sprintf("%s/%s.log", config.GetLogBasePath(), dep.ID)
	jContext.logger = jlogger.New(logPath, config.GetLogLevel())
	if jContext.logger == nil {
		return nil, errors.New("failed to initialize job logger")
	}

	if opCommandFunc, ok := dep.ExtraData["opCommandFunc"]; ok {
		if reflect.TypeOf(opCommandFunc).Kind() == reflect.Func {
			if funcRef, ok := opCommandFunc.(job.CheckOPCmdFunc); ok {
				jContext.opCommandFunc = funcRef
			}
		}
	}
	if jContext.opCommandFunc == nil {
		return nil, errors.New("failed to inject opCommandFunc")
	}

	if checkInFunc, ok := dep.ExtraData["checkInFunc"]; ok {
		if reflect.TypeOf(checkInFunc).Kind() == reflect.Func {
			if funcRef, ok := checkInFunc.(job.CheckInFunc); ok {
				jContext.checkInFunc = funcRef
			}
		}
	}

	if jContext.checkInFunc == nil {
		return nil, errors.New("failed to inject checkInFunc")
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
	if c.checkInFunc != nil {
		c.checkInFunc(status)
	} else {
		return errors.New("nil check in function")
	}

	return nil
}

//OPCommand return the control operational command like stop/cancel if have
func (c *Context) OPCommand() (string, bool) {
	if c.opCommandFunc != nil {
		return c.opCommandFunc()
	}

	return "", false
}

//GetLogger returns the logger
func (c *Context) GetLogger() logger.Interface {
	return c.logger
}
