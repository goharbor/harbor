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
	"github.com/goharbor/harbor/src/pkg/task"
)

// ExecutionController manages the execution.
type ExecutionController interface {
	// Create an execution. The "vendorType" specifies the type of vendor (e.g. replication, scan, gc, retention, etc.),
	// and the "vendorID" specifies the ID of vendor if needed(e.g. policy ID for replication and retention).
	// The "extraAttrs" can be used to set the customized attributes.
	Create(ctx context.Context, vendorType string, vendorID int64, trigger string,
		extraAttrs ...map[string]interface{}) (id int64, err error)
	// MarkDone marks the status of the specified execution as success.
	// It must be called to update the execution status if the created execution contains no tasks.
	// In other cases, the execution status can be calculated from the referenced tasks automatically
	// and no need to update it explicitly.
	MarkDone(ctx context.Context, id int64, message string) (err error)
	// MarkError marks the status of the specified execution as error.
	// It must be called to update the execution status when failed to create tasks.
	// In other cases, the execution status can be calculated from the referenced tasks automatically
	// and no need to update it explicitly.
	MarkError(ctx context.Context, id int64, message string) (err error)
	// Stop all linked tasks of the specified execution.
	Stop(ctx context.Context, id int64) (err error)
	// Delete the specified execution and its tasks.
	Delete(ctx context.Context, id int64) (err error)
	// Get the specified execution.
	Get(ctx context.Context, id int64) (execution *task.Execution, err error)
	// List executions according to the query.
	List(ctx context.Context, query *q.Query) (executions []*task.Execution, err error)
}

// executionController defines the execution controller.
type executionController struct {
	mgr task.ExecutionManager
}

// NewController creates an instance of the default execution controller.
func NewExecutionController() ExecutionController {
	return &executionController{
		mgr: task.ExecMgr,
	}
}

// Create an execution. The "vendorType" specifies the type of vendor (e.g. replication, scan, gc, retention, etc.),
// and the "vendorID" specifies the ID of vendor if needed(e.g. policy ID for replication and retention).
// The "extraAttrs" can be used to set the customized attributes.
func (ec *executionController) Create(ctx context.Context, vendorType string, vendorID int64, trigger string,
	extraAttrs ...map[string]interface{}) (id int64, err error) {
	return ec.mgr.Create(ctx, vendorType, vendorID, trigger, extraAttrs...)
}

// MarkDone marks the status of the specified execution as success.
// It must be called to update the execution status if the created execution contains no tasks.
// In other cases, the execution status can be calculated from the referenced tasks automatically
// and no need to update it explicitly.
func (ec *executionController) MarkDone(ctx context.Context, id int64, message string) (err error) {
	return ec.mgr.MarkDone(ctx, id, message)
}

// MarkError marks the status of the specified execution as error.
// It must be called to update the execution status when failed to create tasks.
// In other cases, the execution status can be calculated from the referenced tasks automatically
// and no need to update it explicitly.
func (ec *executionController) MarkError(ctx context.Context, id int64, message string) (err error) {
	return ec.mgr.MarkError(ctx, id, message)
}

// Stop all linked tasks of the specified execution.
func (ec *executionController) Stop(ctx context.Context, id int64) (err error) {
	return ec.mgr.Stop(ctx, id)
}

// Delete the specified execution and its tasks.
func (ec *executionController) Delete(ctx context.Context, id int64) (err error) {
	return ec.mgr.Delete(ctx, id)
}

// Get the specified execution.
func (ec *executionController) Get(ctx context.Context, id int64) (execution *task.Execution, err error) {
	return ec.mgr.Get(ctx, id)
}

// List executions according to the query.
func (ec *executionController) List(ctx context.Context, query *q.Query) (executions []*task.Execution, err error) {
	return ec.mgr.List(ctx, query)
}
