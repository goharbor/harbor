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
	"sync"
	"time"

	comcfg "github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/logger/sweeper"
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
	// other required information
	properties map[string]interface{}
	// admin server client
	cfgMgr comcfg.CfgManager
	// job life cycle tracker
	tracker job.Tracker
	// job logger configs settings map lock
	lock sync.Mutex
}

// NewContext ...
func NewContext(sysCtx context.Context, cfgMgr *comcfg.CfgManager) *Context {
	return &Context{
		sysContext: comcfg.NewContext(sysCtx, cfgMgr),
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
func (c *Context) Build(tracker job.Tracker) (job.Context, error) {
	if tracker == nil || tracker.Job() == nil {
		return nil, errors.New("nil job tracker")
	}

	jContext := &Context{
		sysContext: c.sysContext,
		cfgMgr:     c.cfgMgr,
		properties: make(map[string]interface{}),
		tracker:    tracker,
	}

	// Copy properties
	if len(c.properties) > 0 {
		for k, v := range c.properties {
			jContext.properties[k] = v
		}
	}

	// Refresh config properties
	err := c.cfgMgr.Load()
	if err != nil {
		return nil, err
	}

	props := c.cfgMgr.GetAll()
	for k, v := range props {
		jContext.properties[k] = v
	}

	// Set loggers for job
	c.lock.Lock()
	defer c.lock.Unlock()
	lg, err := createLoggers(tracker.Job().Info.JobID)
	if err != nil {
		return nil, err
	}
	jContext.logger = lg

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
	return c.tracker.CheckIn(status)
}

// OPCommand return the control operational command like stop/cancel if have
func (c *Context) OPCommand() (job.OPCommand, bool) {
	latest, err := c.tracker.Status()
	if err != nil {
		return job.NilCommand, false
	}

	if job.StoppedStatus == latest {
		return job.StopCommand, true
	}

	return job.NilCommand, false
}

// GetLogger returns the logger
func (c *Context) GetLogger() logger.Interface {
	return c.logger
}

// Tracker returns the job tracker attached with the context
func (c *Context) Tracker() job.Tracker {
	return c.tracker
}

// create loggers based on the configurations.
func createLoggers(jobID string) (logger.Interface, error) {
	// Init job loggers here
	lOptions := make([]logger.Option, 0)
	for _, lc := range config.DefaultConfig.JobLoggerConfigs {
		// For running job, the depth should be 5
		if lc.Name == logger.NameFile || lc.Name == logger.NameStdOutput || lc.Name == logger.NameDB {
			if lc.Settings == nil {
				lc.Settings = map[string]interface{}{}
			}
			lc.Settings["depth"] = 5
		}
		if lc.Name == logger.NameFile || lc.Name == logger.NameDB {
			// Need extra param
			fSettings := map[string]interface{}{}
			for k, v := range lc.Settings {
				// Copy settings
				fSettings[k] = v
			}
			if lc.Name == logger.NameFile {
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
	return logger.GetLogger(lOptions...)
}

func initDBCompleted() error {
	sweeper.PrepareDBSweep()
	return nil
}
