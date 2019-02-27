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

package operation

import (
	"github.com/goharbor/harbor/src/replication/ng/execution"
	"github.com/goharbor/harbor/src/replication/ng/flow"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

// Controller handles the replication-related operations: start,
// stop, query, etc.
type Controller interface {
	StartReplication(policy *model.Policy) (int64, error)
	StopReplication(int64) error
	ListExecutions(...*model.ExecutionQuery) (int64, []*model.Execution, error)
	GetExecution(int64) (*model.Execution, error)
	ListTasks(...*model.TaskQuery) (int64, []*model.Task, error)
	GetTaskLog(int64) ([]byte, error)
}

// NewController returns a controller implementation
func NewController(flowCtl flow.Controller, executionMgr execution.Manager) Controller {
	return &defaultController{
		flowCtl:      flowCtl,
		executionMgr: executionMgr,
	}
}

type defaultController struct {
	flowCtl      flow.Controller
	executionMgr execution.Manager
}

func (d *defaultController) StartReplication(policy *model.Policy) (int64, error) {
	return d.flowCtl.StartReplication(policy)
}
func (d *defaultController) StopReplication(executionID int64) error {
	return d.flowCtl.StopReplication(executionID)
}
func (d *defaultController) ListExecutions(query ...*model.ExecutionQuery) (int64, []*model.Execution, error) {
	return d.executionMgr.List(query...)
}
func (d *defaultController) GetExecution(executionID int64) (*model.Execution, error) {
	return d.executionMgr.Get(executionID)
}
func (d *defaultController) ListTasks(query ...*model.TaskQuery) (int64, []*model.Task, error) {
	return d.executionMgr.ListTasks(query...)
}
func (d *defaultController) GetTaskLog(taskID int64) ([]byte, error) {
	return d.executionMgr.GetTaskLog(taskID)
}
