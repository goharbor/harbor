// Copyright 2018 The Harbor Authors. All rights reserved.

package impl

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/vmware/harbor/src/adminserver/client"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/jobservice/config"
	"github.com/vmware/harbor/src/jobservice/env"
	"github.com/vmware/harbor/src/jobservice/job"
	jlogger "github.com/vmware/harbor/src/jobservice/job/impl/logger"
	"github.com/vmware/harbor/src/jobservice/logger"
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

	//other required information
	properties map[string]interface{}

	//admin server client
	adminClient client.Client
}

//NewContext ...
func NewContext(sysCtx context.Context, adminClient client.Client) *Context {
	return &Context{
		sysContext:  sysCtx,
		adminClient: adminClient,
		properties:  make(map[string]interface{}),
	}
}

//Init ...
func (c *Context) Init() error {
	configs, err := c.adminClient.GetCfgs()
	if err != nil {
		return err
	}

	db := getDBFromConfig(configs)

	return dao.InitDatabase(db)
}

//Build implements the same method in env.JobContext interface
//This func will build the job execution context before running
func (c *Context) Build(dep env.JobData) (env.JobContext, error) {
	jContext := &Context{
		sysContext:  c.sysContext,
		adminClient: c.adminClient,
		properties:  make(map[string]interface{}),
	}

	//Copy properties
	if len(c.properties) > 0 {
		for k, v := range c.properties {
			jContext.properties[k] = v
		}
	}

	//Refresh admin server properties
	props, err := c.adminClient.GetCfgs()
	if err != nil {
		return nil, err
	}
	for k, v := range props {
		jContext.properties[k] = v
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
func (c *Context) Get(prop string) (interface{}, bool) {
	v, ok := c.properties[prop]
	return v, ok
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

func getDBFromConfig(cfg map[string]interface{}) *models.Database {
	database := &models.Database{}
	database.Type = cfg[common.DatabaseType].(string)
	postgresql := &models.PostGreSQL{}
	postgresql.Host = cfg[common.PostGreSQLHOST].(string)
	postgresql.Port = int(cfg[common.PostGreSQLPort].(float64))
	postgresql.Username = cfg[common.PostGreSQLUsername].(string)
	postgresql.Password = cfg[common.PostGreSQLPassword].(string)
	postgresql.Database = cfg[common.PostGreSQLDatabase].(string)
	database.PostGreSQL = postgresql

	return database
}
