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

package replication

import (
	"context"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/controller/replication/flow"
	replicationmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/reg"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/replication"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

func init() {
	// keep only the latest created 50 replication execution records
	task.SetExecutionSweeperCount(job.Replication, 50)
}

// Ctl is a global replication controller instance
var Ctl = NewController()

// Controller defines the operations related with replication
type Controller interface {
	// PolicyCount returns the total count of policies according to the query
	PolicyCount(ctx context.Context, query *q.Query) (count int64, err error)
	// ListPolicies lists the policies according to the query
	ListPolicies(ctx context.Context, query *q.Query) (policies []*replicationmodel.Policy, err error)
	// GetPolicy gets the specific policy
	GetPolicy(ctx context.Context, id int64) (policy *replicationmodel.Policy, err error)
	// CreatePolicy creates a policy
	CreatePolicy(ctx context.Context, policy *replicationmodel.Policy) (id int64, err error)
	// UpdatePolicy updates the specific policy
	UpdatePolicy(ctx context.Context, policy *replicationmodel.Policy, props ...string) (err error)
	// DeletePolicy deletes the specific policy
	DeletePolicy(ctx context.Context, id int64) (err error)
	// Start the replication according to the policy
	Start(ctx context.Context, policy *replicationmodel.Policy, resource *model.Resource, trigger string) (executionID int64, err error)
	// Stop the replication specified by the execution ID
	Stop(ctx context.Context, executionID int64) (err error)
	// ExecutionCount returns the total count of executions according to the query
	ExecutionCount(ctx context.Context, query *q.Query) (count int64, err error)
	// ListExecutions lists the executions according to the query
	ListExecutions(ctx context.Context, query *q.Query) (executions []*Execution, err error)
	// GetExecution gets the specific execution
	GetExecution(ctx context.Context, executionID int64) (execution *Execution, err error)
	// TaskCount returns the total count of tasks according to the query
	TaskCount(ctx context.Context, query *q.Query) (count int64, err error)
	// ListTasks lists the tasks according to the query
	ListTasks(ctx context.Context, query *q.Query) (tasks []*Task, err error)
	// GetTask gets the specific task
	GetTask(ctx context.Context, taskID int64) (task *Task, err error)
	// GetTaskLog gets the log of the specific task
	GetTaskLog(ctx context.Context, taskID int64) (log []byte, err error)
}

// NewController creates a new instance of the replication controller
func NewController() Controller {
	return &controller{
		repMgr:     replication.Mgr,
		execMgr:    task.ExecMgr,
		taskMgr:    task.Mgr,
		regMgr:     reg.Mgr,
		scheduler:  scheduler.Sched,
		flowCtl:    flow.NewController(),
		ormCreator: orm.Crt,
		wp:         lib.NewWorkerPool(10),
	}
}

type controller struct {
	repMgr     replication.Manager
	execMgr    task.ExecutionManager
	taskMgr    task.Manager
	regMgr     reg.Manager
	scheduler  scheduler.Scheduler
	flowCtl    flow.Controller
	ormCreator orm.Creator
	wp         *lib.WorkerPool
}

func (c *controller) Start(ctx context.Context, policy *replicationmodel.Policy, resource *model.Resource, trigger string) (int64, error) {
	logger := log.GetLogger(ctx)
	if !policy.Enabled {
		return 0, errors.New(nil).WithCode(errors.PreconditionCode).
			WithMessage("the policy %d is disabled", policy.ID)
	}
	// create an execution record
	id, err := c.execMgr.Create(ctx, job.Replication, policy.ID, trigger)
	if err != nil {
		return 0, err
	}
	// start the replication flow in background
	// as the process runs inside a goroutine, the transaction in the outer ctx
	// may be submitted already when the process starts, so create an new context
	// with orm populated to the goroutine
	go func() {
		c.wp.GetWorker()
		defer c.wp.ReleaseWorker()

		ctx := orm.NewContext(context.Background(), c.ormCreator.Create())
		// recover in case panic during the adapter process
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("recovered from the panic: %v", err)
				c.markError(ctx, id, fmt.Errorf("panic during the process"))
			}
		}()

		// as we start a new transaction in the goroutine, the execution record may not
		// be inserted yet, wait until it is ready before continue
		if err := retry.Retry(func() error {
			_, err := c.execMgr.Get(ctx, id)
			return err
		}); err != nil {
			c.markError(ctx, id, fmt.Errorf(
				"failed to wait the execution record to be inserted: %v", err))
			return
		}

		err := c.flowCtl.Start(ctx, id, policy, resource)
		if err == nil {
			// no err, return directly
			return
		}
		c.markError(ctx, id, err)
	}()
	return id, nil
}

func (c *controller) markError(ctx context.Context, executionID int64, err error) {
	logger := log.GetLogger(ctx)
	// try to stop the execution first in case that some tasks are already created
	if err := c.execMgr.StopAndWait(ctx, executionID, 10*time.Second); err != nil {
		logger.Errorf("failed to stop the execution %d: %v", executionID, err)
	}
	if err := c.execMgr.MarkError(ctx, executionID, err.Error()); err != nil {
		logger.Errorf("failed to mark error for the execution %d: %v", executionID, err)
	}
}

func (c *controller) Stop(ctx context.Context, id int64) error {
	return c.execMgr.Stop(ctx, id)
}

func (c *controller) ExecutionCount(ctx context.Context, query *q.Query) (int64, error) {
	return c.execMgr.Count(ctx, c.buildExecutionQuery(query))
}

func (c *controller) ListExecutions(ctx context.Context, query *q.Query) ([]*Execution, error) {
	execs, err := c.execMgr.List(ctx, c.buildExecutionQuery(query))
	if err != nil {
		return nil, err
	}
	var executions []*Execution
	for _, exec := range execs {
		executions = append(executions, convertExecution(exec))
	}
	return executions, nil
}

func (c *controller) buildExecutionQuery(query *q.Query) *q.Query {
	// as the following logic may change the content of the query, clone it first
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.Replication
	// convert the query keyword "PolicyID" or "policy_id" to the "VendorID"
	if value, exist := query.Keywords["PolicyID"]; exist {
		query.Keywords["VendorID"] = value
		delete(query.Keywords, "PolicyID")
	}
	if value, exist := query.Keywords["policy_id"]; exist {
		query.Keywords["VendorID"] = value
		delete(query.Keywords, "policy_id")
	}
	return query
}

func (c *controller) GetExecution(ctx context.Context, id int64) (*Execution, error) {
	execs, err := c.execMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ID":         id,
			"VendorType": job.Replication,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(execs) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessage("replication execution %d not found", id)
	}
	return convertExecution(execs[0]), nil
}

func (c *controller) TaskCount(ctx context.Context, query *q.Query) (int64, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.Replication
	return c.taskMgr.Count(ctx, query)
}

func (c *controller) ListTasks(ctx context.Context, query *q.Query) ([]*Task, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.Replication
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

func (c *controller) GetTask(ctx context.Context, id int64) (*Task, error) {
	tasks, err := c.taskMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ID":         id,
			"VendorType": job.Replication,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessage("replication task %d not found", id)
	}
	return convertTask(tasks[0]), nil
}

func (c *controller) GetTaskLog(ctx context.Context, id int64) ([]byte, error) {
	// make sure the task specified by ID is replication task
	_, err := c.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	return c.taskMgr.GetLog(ctx, id)
}

func convertExecution(exec *task.Execution) *Execution {
	return &Execution{
		ID:            exec.ID,
		PolicyID:      exec.VendorID,
		Status:        exec.Status,
		StatusMessage: exec.StatusMessage,
		Metrics:       exec.Metrics,
		Trigger:       exec.Trigger,
		StartTime:     exec.StartTime,
		EndTime:       exec.EndTime,
	}
}

func convertTask(task *task.Task) *Task {
	return &Task{
		ID:                  task.ID,
		ExecutionID:         task.ExecutionID,
		Status:              task.Status,
		StatusMessage:       task.StatusMessage,
		RunCount:            task.RunCount,
		ResourceType:        task.GetStringFromExtraAttrs("resource_type"),
		SourceResource:      task.GetStringFromExtraAttrs("source_resource"),
		DestinationResource: task.GetStringFromExtraAttrs("destination_resource"),
		Operation:           task.GetStringFromExtraAttrs("operation"),
		JobID:               task.JobID,
		CreationTime:        task.CreationTime,
		StartTime:           task.StartTime,
		UpdateTime:          task.UpdateTime,
		EndTime:             task.EndTime,
	}
}
