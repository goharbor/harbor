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

package retention

import (
	"fmt"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/q"
	"strconv"
	"time"
)

// APIController to handle the requests related with retention
type APIController interface {
	// Handle the related hooks from the job service and launch the corresponding actions if needed
	//
	//  Arguments:
	//    policyID string         : uuid of the retention policy
	//    event *job.StatusChange : event object sent by job service
	//
	//  Returns:
	//    common error object if any errors occurred
	HandleHook(policyID string, event *job.StatusChange) error

	GetRetention(id int64) (*policy.Metadata, error)

	CreateRetention(p *policy.Metadata) error

	UpdateRetention(p *policy.Metadata) error

	DeleteRetention(id int64) error

	TriggerRetentionExec(policyID int64, trigger string) error

	StopRetentionExec(eid int64) error

	ListRetentionExec(policyID int64, query *q.Query) ([]*Execution, error)

	ListRetentionExecHistory(executionID int64, query *q.Query) ([]*Task, error)
}

// DefaultAPIController ...
type DefaultAPIController struct {
	manager           Manager
	projectManager    promgr.ProjectManager
	projectManagerNew project.Manager
	repositoryMgr     repository.Manager
	scheduler         Scheduler
}

// GetRetention Get Retention
func (r *DefaultAPIController) GetRetention(id int64) (*policy.Metadata, error) {
	return r.manager.GetPolicy(id)
}

// CreateRetention Create Retention
func (r *DefaultAPIController) CreateRetention(p *policy.Metadata) error {
	if p.Scope.Level != "project" {
		return fmt.Errorf("scope %s is not support", p.Scope.Level)
	}
	if p.Scope.Reference <= 0 {
		return fmt.Errorf("Invalid Project id %d", p.Scope.Reference)
	}
	proj, err := r.projectManager.Get(p.Scope.Reference)
	if err != nil {
		return (err)
	}
	if proj == nil {
		return fmt.Errorf("Invalid Project id %d", p.Scope.Reference)
	}

	if p.Trigger.Kind == "Schedule" {
		jobid, err := r.scheduler.Schedule(strconv.FormatInt(p.ID, 10), p.Trigger.Settings["cron"].(string))
		if err != nil {
			return err
		}
		p.Trigger.References["jobid"] = jobid
	}
	if _, err = r.manager.CreatePolicy(p); err != nil {
		return err
	}
	if err = r.projectManager.GetMetadataManager().Add(p.Scope.Reference,
		map[string]string{"retention_id": strconv.FormatInt(p.Scope.Reference, 10)}); err != nil {
		return err
	}
	return err
}

// UpdateRetention Update Retention
func (r *DefaultAPIController) UpdateRetention(p *policy.Metadata) error {
	p0, err := r.manager.GetPolicy(p.ID)
	if err != nil {
		return err
	}
	if p0.Trigger.Kind == "Schedule" {
		if p.Trigger.Kind == "Schedule" {
			if p0.Trigger.Settings["cron"] != p.Trigger.Settings["cron"] {
				err = r.scheduler.UnSchedule(p0.Trigger.References["jobid"].(string))
				if err != nil {
					return err
				}
				jobid, err := r.scheduler.Schedule(strconv.FormatInt(p.ID, 10), p.Trigger.Settings["cron"].(string))
				if err != nil {
					return err
				}
				p.Trigger.References["jobid"] = jobid
			} else {
				p.Trigger.References["jobid"] = p0.Trigger.References["jobid"]
			}
		} else {
			err = r.scheduler.UnSchedule(p0.Trigger.References["jobid"].(string))
			if err != nil {
				return err
			}
		}
	} else {
		if p.Trigger.Kind == "Schedule" {
			jobid, err := r.scheduler.Schedule(strconv.FormatInt(p.ID, 10), p.Trigger.Settings["cron"].(string))
			if err != nil {
				return err
			}
			p.Trigger.References["jobid"] = jobid
		}
	}
	return r.manager.UpdatePolicy(p)
}

// DeleteRetention Delete Retention
func (r *DefaultAPIController) DeleteRetention(id int64) error {
	p, err := r.manager.GetPolicy(id)
	if err != nil {
		return err
	}
	if p.Trigger.Kind == "Schedule" {
		err = r.scheduler.UnSchedule(p.Trigger.References["jobid"].(string))
		if err != nil {
			return err
		}
	}

	execs, err := r.manager.ListExecutions(id, nil)
	if err != nil {
		return err
	}
	for _, e := range execs {
		err = r.manager.DeleteTasks(e.ID, nil)
		if err != nil {
			return err
		}
		err = r.manager.DeleteExecutions(e.ID, nil)
		return err
	}

	return r.manager.DeletePolicy(id)
}

// TriggerRetentionExec Trigger Retention Execution
func (r *DefaultAPIController) TriggerRetentionExec(policyID int64, trigger string) error {
	launcher := NewLauncher(r.projectManagerNew, r.repositoryMgr, r.manager, dep.DefaultClient)
	p, err := r.manager.GetPolicy(policyID)
	if err != nil {
		return err
	}

	exec := &Execution{
		PolicyID:  policyID,
		StartTime: time.Now(),
		Status:    "Running",
		Trigger:   trigger,
	}
	id, err := r.manager.CreateExecution(exec)
	num, err := launcher.Launch(p, id)
	if err != nil {
		return err
	}
	if num == 0 {
		exec := &Execution{
			ID:      id,
			EndTime: time.Now(),
			Status:  "Success",
		}
		err = r.manager.UpdateExecution(exec)
		if err != nil {
			return err
		}
	}
	return err

}

// StopRetentionExec Stop Retention Execution
func (r *DefaultAPIController) StopRetentionExec(eid int64) error {
	e, err := r.manager.GetExecution(eid)
	if err != nil {
		return err
	}
	if e.Status != "Running" {
		return fmt.Errorf("Can't abort, current status is %s", e.Status)
	}
	exec := &Execution{}
	exec.ID = eid
	exec.Status = "Abort"
	exec.EndTime = time.Now()
	// TODO stop the execution
	return r.manager.UpdateExecution(exec)
}

// ListRetentionExec List Retention Executions
func (r *DefaultAPIController) ListRetentionExec(policyID int64, query *q.Query) ([]*Execution, error) {
	return r.manager.ListExecutions(policyID, query)
}

// ListRetentionExecHistory List Retention Execution Histories
func (r *DefaultAPIController) ListRetentionExecHistory(executionID int64, query *q.Query) ([]*Task, error) {
	q1 := &q.TaskQuery{
		ExecutionID: executionID,
		PageNumber:  query.PageNumber,
		PageSize:    query.PageSize,
	}
	return r.manager.ListTasks(q1)
}

// HandleHook HandleHook
func (r *DefaultAPIController) HandleHook(policyID string, event *job.StatusChange) error {
	panic("implement me")
}

// NewAPIController ...
func NewAPIController(projectManager promgr.ProjectManager, projectManagerNew project.Manager, repositoryMgr repository.Manager, scheduler Scheduler) APIController {
	return &DefaultAPIController{
		manager:           NewManager(),
		projectManager:    projectManager,
		projectManagerNew: projectManagerNew,
		repositoryMgr:     repositoryMgr,
		scheduler:         scheduler,
	}
}
