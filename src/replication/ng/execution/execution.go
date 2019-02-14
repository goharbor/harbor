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
	"github.com/goharbor/harbor/src/replication/ng/model"
)

// Manager manages the executions
type Manager interface {
	// Create a new execution
	Create(*model.Execution) (int64, error)
	// List the summaries of executions
	List(*model.ExecutionQuery) (int64, []*model.Execution, error)
	// Get the specified execution
	Get(int64) (*model.Execution, error)
	// Update the data of the specified execution, the "props" are the
	// properties of execution that need to be updated
	Update(execution *model.Execution, props ...string) error
	// Remove the execution specified by the ID
	Remove(int64) error
	// Remove all executions of one policy specified by the policy ID
	RemoveAll(int64) error
	// Create a task
	CreateTask(*model.Task) (int64, error)
	// List the tasks according to the query
	ListTasks(*model.TaskQuery) (int64, []*model.Task, error)
	// Get one specified task
	GetTask(int64) (*model.Task, error)
	// Update the task, the "props" are the properties of task
	// that need to be updated
	UpdateTask(task *model.Task, props ...string) error
	// UpdateInitializedTask only updates the task whose status is pending,
	// the "props" are the properties of task that need to be updated
	UpdateInitializedTask(task *model.Task, props ...string) error
	// Remove one task specified by task ID
	RemoveTask(int64) error
	// Remove all tasks of one execution specified by the execution ID
	RemoveAllTasks(int64) error
	// Get the log of one specific task
	GetTaskLog(int64) ([]byte, error)
}
