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

	"github.com/goharbor/harbor/src/lib/q"
)

// Manager manages tasks.
// The execution and task managers provide an execution-task model to abstract the interactive with jobservice.
// All of the operations with jobservice should be delegated by them
type Manager interface {
	// Create submits the job to jobservice and creates a corresponding task record.
	// An execution must be created first and the task will be linked to it.
	// The "extraAttrs" can be used to set the customized attributes
	Create(ctx context.Context, executionID int64, job *Job, extraAttrs ...map[string]interface{}) (id int64, err error)
	// Stop the specified task
	Stop(ctx context.Context, id int64) (err error)
	// Get the specified task
	Get(ctx context.Context, id int64) (task *Task, err error)
	// List the tasks according to the query
	List(ctx context.Context, query *q.Query) (tasks []*Task, err error)
	// Get the log of the specified task
	GetLog(ctx context.Context, id int64) (log []byte, err error)
}
