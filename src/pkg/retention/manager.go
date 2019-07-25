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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/pkg/retention/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/q"
)

// Manager defines operations of managing policy
type Manager interface {
	// Create new policy and return ID
	CreatePolicy(p *policy.Metadata) (int64, error)
	// Update the existing policy
	// Full update
	UpdatePolicy(p *policy.Metadata) error
	// Delete the specified policy
	// No actual use so far
	DeletePolicyAndExec(ID int64) error
	// Get the specified policy
	GetPolicy(ID int64) (*policy.Metadata, error)
	// Create a new retention execution
	CreateExecution(execution *Execution) (int64, error)
	// Update the specified execution
	UpdateExecution(execution *Execution) error
	// Get the specified execution
	GetExecution(eid int64) (*Execution, error)
	// List execution histories
	ListExecutions(policyID int64, query *q.Query) ([]*Execution, error)
	// List tasks histories
	ListTasks(query ...*q.TaskQuery) ([]*Task, error)
	// Create a new retention task
	CreateTask(task *Task) (int64, error)
	// Update the specified task
	UpdateTask(task *Task, cols ...string) error
	// Get the log of the specified task
	GetTaskLog(taskID int64) ([]byte, error)
}

// DefaultManager ...
type DefaultManager struct {
}

// CreatePolicy Create Policy
func (d *DefaultManager) CreatePolicy(p *policy.Metadata) (int64, error) {
	p1 := &models.RetentionPolicy{}
	p1.ScopeLevel = p.Scope.Level
	p1.TriggerKind = p.Trigger.Kind
	data, _ := json.Marshal(p)
	p1.Data = string(data)
	p1.CreateTime = time.Now()
	p1.UpdateTime = p1.CreateTime
	return dao.CreatePolicy(p1)
}

// UpdatePolicy Update Policy
func (d *DefaultManager) UpdatePolicy(p *policy.Metadata) error {
	p1 := &models.RetentionPolicy{}
	p1.ID = p.ID
	p1.ScopeLevel = p.Scope.Level
	p1.TriggerKind = p.Trigger.Kind
	p.ID = 0
	data, _ := json.Marshal(p)
	p.ID = p1.ID
	p1.Data = string(data)
	p1.UpdateTime = time.Now()
	return dao.UpdatePolicy(p1, "scope_level", "trigger_kind", "data", "update_time")
}

// DeletePolicyAndExec Delete Policy
func (d *DefaultManager) DeletePolicyAndExec(id int64) error {
	return dao.DeletePolicyAndExec(id)
}

// GetPolicy Get Policy
func (d *DefaultManager) GetPolicy(id int64) (*policy.Metadata, error) {
	p1, err := dao.GetPolicy(id)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	p := &policy.Metadata{}
	if err = json.Unmarshal([]byte(p1.Data), p); err != nil {
		return nil, err
	}
	p.ID = id
	return p, nil
}

// CreateExecution Create Execution
func (d *DefaultManager) CreateExecution(execution *Execution) (int64, error) {
	exec := &models.RetentionExecution{}
	exec.PolicyID = execution.PolicyID
	exec.StartTime = time.Now()
	exec.DryRun = execution.DryRun
	exec.Status = "Running"
	exec.Trigger = "manual"
	return dao.CreateExecution(exec)
}

// UpdateExecution Update Execution
func (d *DefaultManager) UpdateExecution(execution *Execution) error {
	exec := &models.RetentionExecution{}
	exec.ID = execution.ID
	exec.EndTime = execution.EndTime
	exec.Status = execution.Status
	return dao.UpdateExecution(exec, "end_time", "status")
}

// ListExecutions List Executions
func (d *DefaultManager) ListExecutions(policyID int64, query *q.Query) ([]*Execution, error) {
	execs, err := dao.ListExecutions(policyID, query)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var execs1 []*Execution
	for _, e := range execs {
		e1 := &Execution{}
		e1.ID = e.ID
		e1.PolicyID = e.PolicyID
		e1.Status = e.Status
		e1.StartTime = e.StartTime
		e1.EndTime = e.EndTime
		execs1 = append(execs1, e1)
	}
	return execs1, nil
}

// GetExecution Get Execution
func (d *DefaultManager) GetExecution(eid int64) (*Execution, error) {
	e, err := dao.GetExecution(eid)
	if err != nil {
		return nil, err
	}
	e1 := &Execution{}
	e1.ID = e.ID
	e1.PolicyID = e.PolicyID
	e1.Status = e.Status
	e1.StartTime = e.StartTime
	e1.EndTime = e.EndTime
	return e1, nil
}

// CreateTask creates task record
func (d *DefaultManager) CreateTask(task *Task) (int64, error) {
	if task == nil {
		return 0, errors.New("nil task")
	}
	t := &models.RetentionTask{
		ExecutionID: task.ExecutionID,
		JobID:       task.JobID,
		Status:      task.Status,
		StartTime:   task.StartTime,
		EndTime:     task.EndTime,
	}
	return dao.CreateTask(t)
}

// ListTasks lists tasks according to the query
func (d *DefaultManager) ListTasks(query ...*q.TaskQuery) ([]*Task, error) {
	ts, err := dao.ListTask(query...)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	tasks := []*Task{}
	for _, t := range ts {
		tasks = append(tasks, &Task{
			ID:          t.ID,
			ExecutionID: t.ExecutionID,
			JobID:       t.JobID,
			Status:      t.Status,
			StartTime:   t.StartTime,
			EndTime:     t.EndTime,
		})
	}
	return tasks, nil
}

// UpdateTask updates the task
func (d *DefaultManager) UpdateTask(task *Task, cols ...string) error {
	if task == nil {
		return errors.New("nil task")
	}
	if task.ID <= 0 {
		return fmt.Errorf("invalid task ID: %d", task.ID)
	}
	return dao.UpdateTask(&models.RetentionTask{
		ID:          task.ID,
		ExecutionID: task.ExecutionID,
		JobID:       task.JobID,
		Status:      task.Status,
		StartTime:   task.StartTime,
		EndTime:     task.EndTime,
	}, cols...)
}

// GetTaskLog gets the logs of task
func (d *DefaultManager) GetTaskLog(taskID int64) ([]byte, error) {
	panic("implement me")
}

// NewManager ...
func NewManager() Manager {
	return &DefaultManager{}
}
