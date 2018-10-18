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
	"reflect"

	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	jlogger "github.com/goharbor/harbor/src/jobservice/job/impl/logger"
	"github.com/goharbor/harbor/src/jobservice/logger"
	jmodel "github.com/goharbor/harbor/src/jobservice/models"
)

// DefaultContext provides a basic job context
type DefaultContext struct {
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
}

// NewDefaultContext is constructor of building DefaultContext
func NewDefaultContext(sysCtx context.Context) env.JobContext {
	return &DefaultContext{
		sysContext: sysCtx,
		properties: make(map[string]interface{}),
	}
}

// Build implements the same method in env.JobContext interface
// This func will build the job execution context before running
func (c *DefaultContext) Build(dep env.JobData) (env.JobContext, error) {
	jContext := &DefaultContext{
		sysContext: c.sysContext,
		properties: make(map[string]interface{}),
	}

	// Copy properties
	if len(c.properties) > 0 {
		for k, v := range c.properties {
			jContext.properties[k] = v
		}
	}

	// Init logger here
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
func (c *DefaultContext) Get(prop string) (interface{}, bool) {
	v, ok := c.properties[prop]
	return v, ok
}

// SystemContext implements the same method in env.JobContext interface
func (c *DefaultContext) SystemContext() context.Context {
	return c.sysContext
}

// Checkin is bridge func for reporting detailed status
func (c *DefaultContext) Checkin(status string) error {
	if c.checkInFunc != nil {
		c.checkInFunc(status)
	} else {
		return errors.New("nil check in function")
	}

	return nil
}

// OPCommand return the control operational command like stop/cancel if have
func (c *DefaultContext) OPCommand() (string, bool) {
	if c.opCommandFunc != nil {
		return c.opCommandFunc()
	}

	return "", false
}

// GetLogger returns the logger
func (c *DefaultContext) GetLogger() logger.Interface {
	return c.logger
}

// LaunchJob launches sub jobs
func (c *DefaultContext) LaunchJob(req jmodel.JobRequest) (jmodel.JobStats, error) {
	if c.launchJobFunc == nil {
		return jmodel.JobStats{}, errors.New("nil launch job function")
	}

	return c.launchJobFunc(req)
}
