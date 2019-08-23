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

package execution

import (
	"fmt"

	"github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/replication/dao"
	"github.com/goharbor/harbor/src/replication/dao/models"
)

// Manager manages the executions
type Manager interface {
	// Create a new execution
	Create(*models.Execution) (int64, error)
	// List the summaries of executions
	List(...*models.ExecutionQuery) (int64, []*models.Execution, error)
	// Get the specified execution
	Get(int64) (*models.Execution, error)
	// Update the data of the specified execution, the "props" are the
	// properties of execution that need to be updated
	Update(execution *models.Execution, props ...string) error
	// Remove the execution specified by the ID
	Remove(int64) error
	// Remove all executions of one policy specified by the policy ID
	RemoveAll(int64) error
	// Create a task
	CreateTask(*models.Task) (int64, error)
	// List the tasks according to the query
	ListTasks(...*models.TaskQuery) (int64, []*models.Task, error)
	// Get one specified task
	GetTask(int64) (*models.Task, error)
	// Update the task, the "props" are the properties of task
	// that need to be updated, it cannot include "status". If
	// you want to update the status, use "UpdateTaskStatus" instead
	UpdateTask(task *models.Task, props ...string) error
	// UpdateTaskStatus only updates the task status. If "statusCondition"
	// presents, only the tasks whose status equal to "statusCondition"
	// will be updated
	UpdateTaskStatus(taskID int64, status string, statusCondition ...string) error
	// Remove one task specified by task ID
	RemoveTask(int64) error
	// Remove all tasks of one execution specified by the execution ID
	RemoveAllTasks(int64) error
	// Get the log of one specific task
	GetTaskLog(int64) ([]byte, error)
}

// DefaultManager ..
type DefaultManager struct {
}

// NewDefaultManager ...
func NewDefaultManager() Manager {
	return &DefaultManager{}
}

// Create a new execution
func (dm *DefaultManager) Create(execution *models.Execution) (int64, error) {
	return dao.AddExecution(execution)
}

// List the summaries of executions
func (dm *DefaultManager) List(queries ...*models.ExecutionQuery) (int64, []*models.Execution, error) {
	total, err := dao.GetTotalOfExecutions(queries...)
	if err != nil {
		return 0, nil, err
	}

	executions, err := dao.GetExecutions(queries...)
	if err != nil {
		return 0, nil, err
	}
	return total, executions, nil
}

// Get the specified execution
func (dm *DefaultManager) Get(id int64) (*models.Execution, error) {
	return dao.GetExecution(id)
}

// Update ...
func (dm *DefaultManager) Update(execution *models.Execution, props ...string) error {
	n, err := dao.UpdateExecution(execution, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("Execution not found error: %d ", execution.ID)
	}
	return nil
}

// Remove the execution specified by the ID
func (dm *DefaultManager) Remove(id int64) error {
	return dao.DeleteExecution(id)
}

// RemoveAll executions of one policy specified by the policy ID
func (dm *DefaultManager) RemoveAll(policyID int64) error {
	return dao.DeleteAllExecutions(policyID)
}

// CreateTask used to create a task
func (dm *DefaultManager) CreateTask(task *models.Task) (int64, error) {
	return dao.AddTask(task)
}

// ListTasks list the tasks according to the query
func (dm *DefaultManager) ListTasks(queries ...*models.TaskQuery) (int64, []*models.Task, error) {
	total, err := dao.GetTotalOfTasks(queries...)
	if err != nil {
		return 0, nil, err
	}

	tasks, err := dao.GetTasks(queries...)
	if err != nil {
		return 0, nil, err
	}
	return total, tasks, nil
}

// GetTask get one specified task
func (dm *DefaultManager) GetTask(id int64) (*models.Task, error) {
	return dao.GetTask(id)
}

// UpdateTask ...
func (dm *DefaultManager) UpdateTask(task *models.Task, props ...string) error {
	n, err := dao.UpdateTask(task, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("Task not found error: %d ", task.ID)
	}
	return nil
}

// UpdateTaskStatus ...
func (dm *DefaultManager) UpdateTaskStatus(taskID int64, status string, statusCondition ...string) error {
	if _, err := dao.UpdateTaskStatus(taskID, status, statusCondition...); err != nil {
		return err
	}
	return nil
}

// RemoveTask remove one task specified by task ID
func (dm *DefaultManager) RemoveTask(id int64) error {
	return dao.DeleteTask(id)
}

// RemoveAllTasks of one execution specified by the execution ID
func (dm *DefaultManager) RemoveAllTasks(executionID int64) error {
	return dao.DeleteAllTasks(executionID)
}

// GetTaskLog get the log of one specific task
func (dm *DefaultManager) GetTaskLog(taskID int64) ([]byte, error) {
	task, err := dao.GetTask(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("Task not found %d ", taskID)
	}

	return utils.GetJobServiceClient().GetJobLog(task.JobID)
}
