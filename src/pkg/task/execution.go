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

package task

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

var (
	// ExecMgr is a global execution manager instance
	ExecMgr    = NewExecutionManager()
	ErrTimeOut = errors.New("stopping the execution timeout")
)

// ExecutionManager manages executions.
// The execution and task managers provide an execution-task model to abstract the interactive with jobservice.
// All of the operations with jobservice should be delegated by them
type ExecutionManager interface {
	// Create an execution. The "vendorType" specifies the type of vendor (e.g. replication, scan, gc, retention, etc.),
	// and the "vendorID" specifies the ID of vendor if needed(e.g. policy ID for replication and retention).
	// The "extraAttrs" can be used to set the customized attributes
	Create(ctx context.Context, vendorType string, vendorID int64, trigger string,
		extraAttrs ...map[string]any) (id int64, err error)
	// Update the extra attributes of the specified execution
	UpdateExtraAttrs(ctx context.Context, id int64, extraAttrs map[string]any) (err error)
	// MarkDone marks the status of the specified execution as success.
	// It must be called to update the execution status if the created execution contains no tasks.
	// In other cases, the execution status can be calculated from the referenced tasks automatically
	// and no need to update it explicitly
	MarkDone(ctx context.Context, id int64, message string) (err error)
	// MarkError marks the status of the specified execution as error.
	// It must be called to update the execution status when failed to create tasks.
	// In other cases, the execution status can be calculated from the referenced tasks automatically
	// and no need to update it explicitly
	MarkError(ctx context.Context, id int64, message string) (err error)
	// Stop all linked tasks of the specified execution
	Stop(ctx context.Context, id int64) (err error)
	// StopAndWait stops all linked tasks of the specified execution and waits until all tasks are stopped
	// or get an error
	StopAndWait(ctx context.Context, id int64, timeout time.Duration) (err error)
	// StopAndWaitWithError calls the StopAndWait first, if it doesn't return error, then it call MarkError if the origError is not empty
	StopAndWaitWithError(ctx context.Context, id int64, timeout time.Duration, origError error) (err error)
	// Delete the specified execution and its tasks
	Delete(ctx context.Context, id int64) (err error)
	// Delete all executions and tasks of the specific vendor. They can be deleted only when all the executions/tasks
	// of the vendor are in final status
	DeleteByVendor(ctx context.Context, vendorType string, vendorID int64) (err error)
	// Get the specified execution
	Get(ctx context.Context, id int64) (execution *Execution, err error)
	// List executions according to the query
	// Query the "ExtraAttrs" by setting 'query.Keywords["ExtraAttrs.key"]="value"'
	List(ctx context.Context, query *q.Query) (executions []*Execution, err error)
	// Count counts total of executions according to the query.
	// Query the "ExtraAttrs" by setting 'query.Keywords["ExtraAttrs.key"]="value"'
	Count(ctx context.Context, query *q.Query) (int64, error)
}

// NewExecutionManager return an instance of the default execution manager
func NewExecutionManager() ExecutionManager {
	return &executionManager{
		executionDAO: dao.NewExecutionDAO(),
		taskMgr:      Mgr,
		taskDAO:      dao.NewTaskDAO(),
	}
}

type executionManager struct {
	executionDAO dao.ExecutionDAO
	taskMgr      Manager
	taskDAO      dao.TaskDAO
}

func (e *executionManager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return e.executionDAO.Count(ctx, query)
}

func (e *executionManager) Create(ctx context.Context, vendorType string, vendorID int64, trigger string,
	extraAttrs ...map[string]any) (int64, error) {
	extras := map[string]any{}
	if len(extraAttrs) > 0 && extraAttrs[0] != nil {
		extras = extraAttrs[0]
	}
	data, err := json.Marshal(extras)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	execution := &dao.Execution{
		VendorType: vendorType,
		VendorID:   vendorID,
		Status:     job.RunningStatus.String(),
		Trigger:    trigger,
		ExtraAttrs: string(data),
		StartTime:  now,
		UpdateTime: now,
	}
	id, err := e.executionDAO.Create(ctx, execution)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (e *executionManager) UpdateExtraAttrs(ctx context.Context, id int64, extraAttrs map[string]any) error {
	data, err := json.Marshal(extraAttrs)
	if err != nil {
		return err
	}

	execution := &dao.Execution{
		ID:         id,
		ExtraAttrs: string(data),
		UpdateTime: time.Now(),
	}

	return e.executionDAO.Update(ctx, execution, "ExtraAttrs", "UpdateTime")
}

func (e *executionManager) MarkDone(ctx context.Context, id int64, message string) error {
	now := time.Now()
	return e.executionDAO.Update(ctx, &dao.Execution{
		ID:            id,
		Status:        job.SuccessStatus.String(),
		StatusMessage: message,
		UpdateTime:    now,
		EndTime:       now,
	}, "Status", "StatusMessage", "UpdateTime", "EndTime")
}

func (e *executionManager) MarkError(ctx context.Context, id int64, message string) error {
	now := time.Now()
	return e.executionDAO.Update(ctx, &dao.Execution{
		ID:            id,
		Status:        job.ErrorStatus.String(),
		StatusMessage: message,
		UpdateTime:    now,
		EndTime:       now,
	}, "Status", "StatusMessage", "UpdateTime", "EndTime")
}

func (e *executionManager) Stop(ctx context.Context, id int64) error {
	execution, err := e.executionDAO.Get(ctx, id)
	if err != nil {
		return err
	}

	// when an execution is in final status, if it contains task that is a periodic or retrying job it will
	// run again in the near future, so we must operate the stop action no matter the status is final or not
	tasks, err := e.taskDAO.List(ctx, &q.Query{
		Keywords: map[string]any{
			"ExecutionID": id,
		},
	})
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		// in final status, return directly
		if job.Status(execution.Status).Final() {
			return nil
		}
		// isn't in final status, update directly.
		// as this is used for the corner case(the case that the execution exists but all tasks are disappeared. In normal
		// cases, if the execution contains no tasks, it is already set as "success" by the upper level caller directly),
		// no need to handle concurrency
		now := time.Now()
		return e.executionDAO.Update(ctx, &dao.Execution{
			ID:         id,
			Status:     job.StoppedStatus.String(),
			Revision:   execution.Revision + 1,
			UpdateTime: now,
			EndTime:    now,
		}, "Status", "Revision", "UpdateTime", "EndTime")
	}

	for _, task := range tasks {
		if err = e.taskMgr.Stop(ctx, task.ID); err != nil {
			log.Errorf("failed to stop task %d: %v", task.ID, err)
			continue
		}
	}
	return err
}

func (e *executionManager) StopAndWait(ctx context.Context, id int64, timeout time.Duration) error {
	var (
		overtime bool
		errChan  = make(chan error, 1)
		lock     = sync.RWMutex{}
	)
	go func() {
		// stop the execution
		if err := e.Stop(ctx, id); err != nil {
			errChan <- err
			return
		}
		// check the status of the execution
		interval := 100 * time.Millisecond
		stop := false
		for !stop {
			execution, err := e.executionDAO.Get(ctx, id)
			if err != nil {
				errChan <- err
				return
			}
			// if the status is final, return
			if job.Status(execution.Status).Final() {
				errChan <- nil
				return
			}
			time.Sleep(interval)
			if interval < 1*time.Second {
				interval = interval * 2
			}
			lock.RLock()
			stop = overtime
			lock.RUnlock()
		}
	}()

	select {
	case <-time.After(timeout):
		lock.Lock()
		overtime = true
		lock.Unlock()
		return ErrTimeOut
	case err := <-errChan:
		return err
	}
}

func (e *executionManager) StopAndWaitWithError(ctx context.Context, id int64, timeout time.Duration, origError error) error {
	if err := e.StopAndWait(ctx, id, timeout); err != nil {
		return err
	}
	if origError != nil {
		return e.MarkError(ctx, id, origError.Error())
	}
	return nil
}

func (e *executionManager) Delete(ctx context.Context, id int64) error {
	tasks, err := e.taskDAO.List(ctx, &q.Query{
		Keywords: map[string]any{
			"ExecutionID": id,
		},
	})
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if !job.Status(task.Status).Final() {
			return errors.New(nil).WithCode(errors.PreconditionCode).
				WithMessagef("the execution %d has tasks that aren't in final status, stop the tasks first", id)
		}

		log.Debugf("delete task %d as execution %d has been deleted", task.ID, task.ExecutionID)
		if err = e.taskDAO.Delete(ctx, task.ID); err != nil {
			// the tasks may be deleted by the other execution deletion operation in the same time(e.g. execution sweeper),
			// ignore the not found error for the tasks
			if errors.IsNotFoundErr(err) {
				continue
			}
			return err
		}
	}

	return e.executionDAO.Delete(ctx, id)
}

func (e *executionManager) DeleteByVendor(ctx context.Context, vendorType string, vendorID int64) error {
	executions, err := e.executionDAO.List(ctx, &q.Query{
		Keywords: map[string]any{
			"VendorType": vendorType,
			"VendorID":   vendorID,
		}})
	if err != nil {
		return err
	}
	// check the status
	for _, execution := range executions {
		if !job.Status(execution.Status).Final() {
			return errors.New(nil).WithCode(errors.PreconditionCode).
				WithMessagef("contains executions that aren't in final status, stop the execution first")
		}
	}
	// delete the executions
	for _, execution := range executions {
		if err = e.Delete(ctx, execution.ID); err != nil {
			if errors.IsNotFoundErr(err) {
				continue
			}
			return err
		}
	}
	return nil
}

func (e *executionManager) Get(ctx context.Context, id int64) (*Execution, error) {
	execution, err := e.executionDAO.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return e.populateExecution(ctx, execution), nil
}

func (e *executionManager) List(ctx context.Context, query *q.Query) ([]*Execution, error) {
	executions, err := e.executionDAO.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var execs []*Execution
	for _, execution := range executions {
		execs = append(execs, e.populateExecution(ctx, execution))
	}
	return execs, nil
}

func (e *executionManager) populateExecution(ctx context.Context, execution *dao.Execution) *Execution {
	exec := &Execution{
		ID:            execution.ID,
		VendorType:    execution.VendorType,
		VendorID:      execution.VendorID,
		Status:        execution.Status,
		StatusMessage: execution.StatusMessage,
		Metrics:       nil,
		Trigger:       execution.Trigger,
		StartTime:     execution.StartTime,
		UpdateTime:    execution.UpdateTime,
		EndTime:       execution.EndTime,
	}

	if len(execution.ExtraAttrs) > 0 {
		extras := map[string]any{}
		if err := json.Unmarshal([]byte(execution.ExtraAttrs), &extras); err != nil {
			log.Errorf("failed to unmarshal the extra attributes of execution %d: %v", execution.ID, err)
		} else {
			exec.ExtraAttrs = extras
		}
	}

	// populate task metrics
	metrics, err := e.executionDAO.GetMetrics(ctx, execution.ID)
	if err != nil {
		log.Errorf("failed to get metrics of the execution %d: %v", execution.ID, err)
	} else {
		exec.Metrics = metrics
	}

	return exec
}
