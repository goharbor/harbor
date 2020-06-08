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
	"time"

	"github.com/goharbor/harbor/src/lib/q"
)

// ExecutionManager manages executions.
// The execution and task managers provide an execution-task model to abstract the interactive with jobservice.
// All of the operations with jobservice should be delegated by them
type ExecutionManager interface {
	// Create an execution. The "vendorType" specifies the type of vendor (replication, scan, gc, retention, etc.),
	// and the "vendorID" specifies the ID of vendor if needed(policy ID for replication and retention).  The
	// "extraAttrs" can be used to set the customized attributes
	Create(ctx context.Context, vendorType string, vendorID int64, trigger string,
		extraAttrs ...map[string]interface{}) (id int64, err error)
	// UpdateStatus updates the status of the execution.
	// In most cases, the execution status can be calculated from the referenced tasks automatically.
	// When the execution contains no tasks or failed to create tasks, the status should be set manually
	UpdateStatus(ctx context.Context, id int64, status, message string, endTime time.Time) (err error)
	// Stop all linked tasks of the specified execution
	Stop(ctx context.Context, id int64) (err error)
	// Delete the specified execution and its tasks
	Delete(ctx context.Context, id int64) (err error)
	// Get the specified execution
	Get(ctx context.Context, id int64) (execution *Execution, err error)
	// List executions according to the query
	List(ctx context.Context, query *q.Query) (executions []*Execution, err error)
}
