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
	o "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/orm"
)

// DefaultContext provides a basic job context
type DefaultContext struct {
	// System context
	sysContext context.Context
	// Logger for job
	logger logger.Interface
	// Other required information
	properties map[string]interface{}
	// Track the job attached with the context
	tracker job.Tracker
}

// NewDefaultContext is constructor of building DefaultContext
func NewDefaultContext(sysCtx context.Context) job.Context {
	return &DefaultContext{
		sysContext: sysCtx,
		properties: make(map[string]interface{}),
	}
}

// Build implements the same method in env.Context interface
// This func will build the job execution context before running
func (dc *DefaultContext) Build(t job.Tracker) (job.Context, error) {
	if t == nil {
		return nil, errors.New("nil job tracker")
	}

	jContext := &DefaultContext{
		// TODO support DB transaction
		sysContext: orm.NewContext(dc.sysContext, o.NewOrm()),
		tracker:    t,
		properties: make(map[string]interface{}),
	}

	// Copy properties
	if len(dc.properties) > 0 {
		for k, v := range dc.properties {
			jContext.properties[k] = v
		}
	}

	// Set loggers for job
	lg, err := createLoggers(t.Job().Info.JobID)
	if err != nil {
		return nil, err
	}

	jContext.logger = lg

	return jContext, nil
}

// Get implements the same method in env.Context interface
func (dc *DefaultContext) Get(prop string) (interface{}, bool) {
	v, ok := dc.properties[prop]
	return v, ok
}

// SystemContext implements the same method in env.Context interface
func (dc *DefaultContext) SystemContext() context.Context {
	return dc.sysContext
}

// Checkin is bridge func for reporting detailed status
func (dc *DefaultContext) Checkin(status string) error {
	return dc.tracker.CheckIn(status)
}

// OPCommand return the control operational command like stop if have
func (dc *DefaultContext) OPCommand() (job.OPCommand, bool) {
	latest, err := dc.tracker.Status()
	if err != nil {
		return job.NilCommand, false
	}

	if job.StoppedStatus == latest {
		return job.StopCommand, true
	}

	return job.NilCommand, false
}

// GetLogger returns the logger
func (dc *DefaultContext) GetLogger() logger.Interface {
	return dc.logger
}

// Tracker returns the tracker tracking the job attached with the context
func (dc *DefaultContext) Tracker() job.Tracker {
	return dc.tracker
}
