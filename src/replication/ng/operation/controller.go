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
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/ng/dao/models"
	"github.com/goharbor/harbor/src/replication/ng/execution"
	"github.com/goharbor/harbor/src/replication/ng/flow"
	"github.com/goharbor/harbor/src/replication/ng/model"
	"github.com/goharbor/harbor/src/replication/ng/registry"
	"github.com/goharbor/harbor/src/replication/ng/scheduler"
)

// Controller handles the replication-related operations: start,
// stop, query, etc.
type Controller interface {
	StartReplication(policy *model.Policy, resource *model.Resource) (int64, error)
	StopReplication(int64) error
	ListExecutions(...*models.ExecutionQuery) (int64, []*models.Execution, error)
	GetExecution(int64) (*models.Execution, error)
	ListTasks(...*models.TaskQuery) (int64, []*models.Task, error)
	GetTask(int64) (*models.Task, error)
	UpdateTaskStatus(id int64, status string, statusCondition ...string) error
	GetTaskLog(int64) ([]byte, error)
}

// NewController returns a controller implementation
func NewController(executionMgr execution.Manager, registrgMgr registry.Manager,
	scheduler scheduler.Scheduler) Controller {
	return &defaultController{
		executionMgr: executionMgr,
		registryMgr:  registrgMgr,
		scheduler:    scheduler,
		flowCtl:      flow.NewController(),
	}
}

type defaultController struct {
	flowCtl      flow.Controller
	executionMgr execution.Manager
	registryMgr  registry.Manager
	scheduler    scheduler.Scheduler
}

func (d *defaultController) StartReplication(policy *model.Policy, resource *model.Resource) (int64, error) {
	if resource != nil && len(resource.Metadata.Vtags) != 1 {
		return 0, fmt.Errorf("the length of Vtags must be 1: %v", resource.Metadata.Vtags)
	}

	id, err := createExecution(d.executionMgr, policy.ID)
	if err != nil {
		return 0, err
	}

	flow := d.createFlow(id, policy, resource)
	if err = d.flowCtl.Start(flow); err != nil {
		// mark the execution as failure and log the error message
		// no error will be returned as the execution is created successfully
		markExecutionFailure(d.executionMgr, id, err.Error())
	}

	return id, nil
}

// create different replication flows according to the input parameters
func (d *defaultController) createFlow(executionID int64, policy *model.Policy, resource *model.Resource) flow.Flow {
	// replicate the deletion operation, so create a deletion flow
	if resource != nil && resource.Deleted {
		return flow.NewDeletionFlow(d.executionMgr, d.registryMgr, d.scheduler, executionID, policy, []*model.Resource{resource})

	}
	// copy only one resource, add extra filters to the  policy to make sure
	// only the resource will be filtered out
	if resource != nil {
		filters := []*model.Filter{
			{
				Type:  model.FilterTypeResource,
				Value: resource.Type,
			},
			{
				Type:  model.FilterTypeName,
				Value: resource.Metadata.Name,
			},
			{
				Type: model.FilterTypeVersion,
				// only support replicate one tag
				Value: resource.Metadata.Vtags[0],
			},
		}
		filters = append(filters, policy.Filters...)
		policy.Filters = filters
	}
	return flow.NewCopyFlow(d.executionMgr, d.registryMgr, d.scheduler, executionID, policy)
}

func (d *defaultController) StopReplication(executionID int64) error {
	// TODO implement the function
	return nil
}
func (d *defaultController) ListExecutions(query ...*models.ExecutionQuery) (int64, []*models.Execution, error) {
	return d.executionMgr.List(query...)
}
func (d *defaultController) GetExecution(executionID int64) (*models.Execution, error) {
	return d.executionMgr.Get(executionID)
}
func (d *defaultController) ListTasks(query ...*models.TaskQuery) (int64, []*models.Task, error) {
	return d.executionMgr.ListTasks(query...)
}
func (d *defaultController) GetTask(id int64) (*models.Task, error) {
	return d.executionMgr.GetTask(id)
}
func (d *defaultController) UpdateTaskStatus(id int64, status string, statusCondition ...string) error {
	return d.executionMgr.UpdateTaskStatus(id, status, statusCondition...)
}
func (d *defaultController) GetTaskLog(taskID int64) ([]byte, error) {
	return d.executionMgr.GetTaskLog(taskID)
}

// create the execution record in database
func createExecution(mgr execution.Manager, policyID int64) (int64, error) {
	id, err := mgr.Create(&models.Execution{
		PolicyID:  policyID,
		Status:    models.ExecutionStatusInProgress,
		StartTime: time.Now(),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create the execution record for replication based on policy %d: %v", policyID, err)
	}
	log.Debugf("an execution record for replication based on the policy %d created: %d", policyID, id)
	return id, nil
}

// mark the execution as failure in database
func markExecutionFailure(mgr execution.Manager, id int64, message string) {
	err := mgr.Update(
		&models.Execution{
			ID:         id,
			Status:     models.ExecutionStatusFailed,
			StatusText: message,
			EndTime:    time.Now(),
		}, "Status", "StatusText", "EndTime")
	if err != nil {
		log.Errorf("failed to update the execution %d: %v", id, err)
		return
	}
	log.Debugf("the execution %d is marked as failure: %s", id, message)
}
