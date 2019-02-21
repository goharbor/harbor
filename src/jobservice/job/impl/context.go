// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package impl

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/goharbor/harbor/src/common"
	comcfg "github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/logger/sweeper"
	jmodel "github.com/goharbor/harbor/src/jobservice/models"
)

const (
	maxRetryTimes = 5
)

// Context ...
type Context struct {
	// System context
	sysContext context.Context

	// Logger for job
	logger logger.Interface

	// op command func
	opCommandFunc job.CheckOPCmdFunc

	// checkin func
	checkInFunc job.CheckInFunc

	// launch job
	launchJobFunc job.LaunchJobFunc

	// other required information
	properties map[string]interface{}

	// admin server client
	cfgMgr comcfg.CfgManager
}

// NewContext ...
func NewContext(sysCtx context.Context, cfgMgr *comcfg.CfgManager) *Context {
	return &Context{
		sysContext: sysCtx,
		cfgMgr:     *cfgMgr,
		properties: make(map[string]interface{}),
	}
}

// Init ...
func (c *Context) Init() error {
	var (
		counter = 0
		err     error
	)

	for counter == 0 || err != nil {
		counter++
		err = c.cfgMgr.Load()
		if err != nil {
			logger.Errorf("Job context initialization error: %s\n", err.Error())
			if counter < maxRetryTimes {
				backoff := (int)(math.Pow(2, (float64)(counter))) + 2*counter + 5
				logger.Infof("Retry in %d seconds", backoff)
				time.Sleep(time.Duration(backoff) * time.Second)
			} else {
				return fmt.Errorf("job context initialization error: %s (%d times tried)", err.Error(), counter)
			}
		}
	}

	db := c.cfgMgr.GetDatabaseCfg()

	err = dao.InitDatabase(db)
	if err != nil {
		return err
	}

	// Initialize DB finished
	initDBCompleted()
	return nil
}

// Build implements the same method in env.JobContext interface
// This func will build the job execution context before running
func (c *Context) Build(dep env.JobData) (env.JobContext, error) {
	jContext := &Context{
		sysContext: c.sysContext,
		cfgMgr:     c.cfgMgr,
		properties: make(map[string]interface{}),
	}

	// Copy properties
	if len(c.properties) > 0 {
		for k, v := range c.properties {
			jContext.properties[k] = v
		}
	}

	// Refresh config properties
	err := c.cfgMgr.Load()
	props := c.cfgMgr.GetAll()
	if err != nil {
		return nil, err
	}
	for k, v := range props {
		jContext.properties[k] = v
	}

	// Set loggers for job
	if err := setLoggers(func(lg logger.Interface) {
		jContext.logger = lg
	}, dep.ID); err != nil {
		return nil, err
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

	if launchJobFunc, ok := dep.ExtraData["launchJobFunc"]; ok {
		if reflect.TypeOf(launchJobFunc).Kind() == reflect.Func {
			if funcRef, ok := launchJobFunc.(job.LaunchJobFunc); ok {
				jContext.launchJobFunc = funcRef
			}
		}
	}

	if jContext.launchJobFunc == nil {
		return nil, errors.New("failed to inject launchJobFunc")
	}

	return jContext, nil
}

// Get implements the same method in env.JobContext interface
func (c *Context) Get(prop string) (interface{}, bool) {
	v, ok := c.properties[prop]
	return v, ok
}

// SystemContext implements the same method in env.JobContext interface
func (c *Context) SystemContext() context.Context {
	return c.sysContext
}

// Checkin is bridge func for reporting detailed status
func (c *Context) Checkin(status string) error {
	if c.checkInFunc != nil {
		c.checkInFunc(status)
	} else {
		return errors.New("nil check in function")
	}

	return nil
}

// OPCommand return the control operational command like stop/cancel if have
func (c *Context) OPCommand() (string, bool) {
	if c.opCommandFunc != nil {
		return c.opCommandFunc()
	}

	return "", false
}

// GetLogger returns the logger
func (c *Context) GetLogger() logger.Interface {
	return c.logger
}

// LaunchJob launches sub jobs
func (c *Context) LaunchJob(req jmodel.JobRequest) (jmodel.JobStats, error) {
	if c.launchJobFunc == nil {
		return jmodel.JobStats{}, errors.New("nil launch job function")
	}

	return c.launchJobFunc(req)
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
	postgresql.SSLMode = cfg[common.PostGreSQLSSLMode].(string)
	database.PostGreSQL = postgresql

	return database
}

// create loggers based on the configurations and set it to the job executing context.
func setLoggers(setter func(lg logger.Interface), jobID string) error {
	if setter == nil {
		return errors.New("missing setter func")
	}

	// Init job loggers here
	lOptions := []logger.Option{}
	for _, lc := range config.DefaultConfig.JobLoggerConfigs {
		// For running job, the depth should be 5
		if lc.Name == logger.LoggerNameFile || lc.Name == logger.LoggerNameStdOutput || lc.Name == logger.LoggerNameDB {
			if lc.Settings == nil {
				lc.Settings = map[string]interface{}{}
			}
			lc.Settings["depth"] = 5
		}
		if lc.Name == logger.LoggerNameFile || lc.Name == logger.LoggerNameDB {
			// Need extra param
			fSettings := map[string]interface{}{}
			for k, v := range lc.Settings {
				// Copy settings
				fSettings[k] = v
			}
			if lc.Name == logger.LoggerNameFile {
				// Append file name param
				fSettings["filename"] = fmt.Sprintf("%s.log", jobID)
				lOptions = append(lOptions, logger.BackendOption(lc.Name, lc.Level, fSettings))
			} else { // DB Logger
				// Append DB key
				fSettings["key"] = jobID
				lOptions = append(lOptions, logger.BackendOption(lc.Name, lc.Level, fSettings))
			}
		} else {
			lOptions = append(lOptions, logger.BackendOption(lc.Name, lc.Level, lc.Settings))
		}
	}
	// Get logger for the job
	lg, err := logger.GetLogger(lOptions...)
	if err != nil {
		return fmt.Errorf("initialize job logger error: %s", err)
	}

	setter(lg)

	return nil
}

func initDBCompleted() error {
	sweeper.PrepareDBSweep()
	return nil
}
