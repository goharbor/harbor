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

package gc

import (
	"context"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

var (
	// Ctl is a global garbage collection controller instance
	Ctl = NewController()
)

// Controller manages the tags
type Controller interface {
	// Start start a manual gc job
	Start(ctx context.Context, policy Policy, trigger string) (int64, error)
	// Stop stop a gc job
	Stop(ctx context.Context, id int64) error

	// ExecutionCount returns the total count of executions according to the query
	ExecutionCount(ctx context.Context, query *q.Query) (count int64, err error)
	// ListExecutions lists the executions according to the query
	ListExecutions(ctx context.Context, query *q.Query) (executions []*Execution, err error)
	// GetExecution gets the specific execution
	GetExecution(ctx context.Context, executionID int64) (execution *Execution, err error)

	// GetTask gets the specific task
	GetTask(ctx context.Context, id int64) (*Task, error)
	// ListTasks lists the tasks according to the query
	ListTasks(ctx context.Context, query *q.Query) (tasks []*Task, err error)
	// GetTaskLog gets log of the specific task
	GetTaskLog(ctx context.Context, id int64) ([]byte, error)

	// GetSchedule get the current gc schedule
	GetSchedule(ctx context.Context) (*scheduler.Schedule, error)
	// CreateSchedule create the gc schedule with cron type & string
	CreateSchedule(ctx context.Context, cronType, cron string, policy Policy) (int64, error)
	// DeleteSchedule remove the gc schedule
	DeleteSchedule(ctx context.Context) error
}

// NewController creates an instance of the default repository controller
func NewController() Controller {
	return &controller{
		taskMgr:      task.NewManager(),
		exeMgr:       task.NewExecutionManager(),
		schedulerMgr: scheduler.New(),
	}
}

type controller struct {
	taskMgr      task.Manager
	exeMgr       task.ExecutionManager
	schedulerMgr scheduler.Scheduler
}

// Start starts the manual GC
func (c *controller) Start(ctx context.Context, policy Policy, trigger string) (int64, error) {
	para := make(map[string]any)
	para["delete_untagged"] = policy.DeleteUntagged
	para["dry_run"] = policy.DryRun
	para["workers"] = policy.Workers
	para["redis_url_reg"] = policy.ExtraAttrs["redis_url_reg"]
	para["time_window"] = policy.ExtraAttrs["time_window"]

	execID, err := c.exeMgr.Create(ctx, job.GarbageCollectionVendorType, -1, trigger, para)
	if err != nil {
		return -1, err
	}
	_, err = c.taskMgr.Create(ctx, execID, &task.Job{
		Name: job.GarbageCollectionVendorType,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: para,
	})
	if err != nil {
		return -1, err
	}
	return execID, nil
}

// Stop ...
func (c *controller) Stop(ctx context.Context, id int64) error {
	return c.exeMgr.Stop(ctx, id)
}

// ExecutionCount ...
func (c *controller) ExecutionCount(ctx context.Context, query *q.Query) (int64, error) {
	query.Keywords["VendorType"] = job.GarbageCollectionVendorType
	return c.exeMgr.Count(ctx, query)
}

// ListExecutions ...
func (c *controller) ListExecutions(ctx context.Context, query *q.Query) ([]*Execution, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.GarbageCollectionVendorType

	execs, err := c.exeMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var executions []*Execution
	for _, exec := range execs {
		executions = append(executions, convertExecution(exec))
	}
	return executions, nil
}

// GetExecution ...
func (c *controller) GetExecution(ctx context.Context, id int64) (*Execution, error) {
	execs, err := c.exeMgr.List(ctx, &q.Query{
		Keywords: map[string]any{
			"ID":         id,
			"VendorType": job.GarbageCollectionVendorType,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(execs) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessagef("garbage collection execution %d not found", id)
	}
	return convertExecution(execs[0]), nil
}

// GetTask ...
func (c *controller) GetTask(ctx context.Context, id int64) (*Task, error) {
	tasks, err := c.taskMgr.List(ctx, &q.Query{
		Keywords: map[string]any{
			"ID":         id,
			"VendorType": job.GarbageCollectionVendorType,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessagef("garbage collection task %d not found", id)
	}
	return convertTask(tasks[0]), nil
}

// ListTasks ...
func (c *controller) ListTasks(ctx context.Context, query *q.Query) ([]*Task, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.GarbageCollectionVendorType
	tks, err := c.taskMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var tasks []*Task
	for _, tk := range tks {
		tasks = append(tasks, convertTask(tk))
	}
	return tasks, nil
}

// GetTaskLog ...
func (c *controller) GetTaskLog(ctx context.Context, id int64) ([]byte, error) {
	_, err := c.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	return c.taskMgr.GetLog(ctx, id)
}

// GetSchedule ...
func (c *controller) GetSchedule(ctx context.Context) (*scheduler.Schedule, error) {
	sch, err := c.schedulerMgr.ListSchedules(ctx, q.New(q.KeyWords{"VendorType": job.GarbageCollectionVendorType}))
	if err != nil {
		return nil, err
	}
	if len(sch) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).WithMessage("no gc schedule is found")
	}
	if sch[0] == nil {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).WithMessage("no gc schedule is found")
	}
	return sch[0], nil
}

// CreateSchedule ...
func (c *controller) CreateSchedule(ctx context.Context, cronType, cron string, policy Policy) (int64, error) {
	extras := make(map[string]any)
	extras["delete_untagged"] = policy.DeleteUntagged
	extras["workers"] = policy.Workers
	return c.schedulerMgr.Schedule(ctx, job.GarbageCollectionVendorType, -1, cronType, cron, job.GarbageCollectionVendorType, policy, extras)
}

// DeleteSchedule ...
func (c *controller) DeleteSchedule(ctx context.Context) error {
	return c.schedulerMgr.UnScheduleByVendor(ctx, job.GarbageCollectionVendorType, -1)
}

func convertExecution(exec *task.Execution) *Execution {
	return &Execution{
		ID:            exec.ID,
		Status:        exec.Status,
		StatusMessage: exec.StatusMessage,
		Trigger:       exec.Trigger,
		ExtraAttrs:    exec.ExtraAttrs,
		StartTime:     exec.StartTime,
		UpdateTime:    exec.UpdateTime,
	}
}

func convertTask(task *task.Task) *Task {
	return &Task{
		ID:             task.ID,
		ExecutionID:    task.ExecutionID,
		Status:         task.Status,
		StatusMessage:  task.StatusMessage,
		RunCount:       task.RunCount,
		DeleteUntagged: task.GetBoolFromExtraAttrs("delete_untagged"),
		DryRun:         task.GetBoolFromExtraAttrs("dry_run"),
		Workers:        int(task.GetNumFromExtraAttrs("workers")),
		JobID:          task.JobID,
		CreationTime:   task.CreationTime,
		StartTime:      task.StartTime,
		UpdateTime:     task.UpdateTime,
		EndTime:        task.EndTime,
	}
}
