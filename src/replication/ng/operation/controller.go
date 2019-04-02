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

	"github.com/goharbor/harbor/src/common/job"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/ng/dao/models"
	"github.com/goharbor/harbor/src/replication/ng/model"
	"github.com/goharbor/harbor/src/replication/ng/operation/execution"
	"github.com/goharbor/harbor/src/replication/ng/operation/flow"
	"github.com/goharbor/harbor/src/replication/ng/operation/scheduler"
)

// Controller handles the replication-related operations: start,
// stop, query, etc.
type Controller interface {
	// trigger is used to specified that what this replication is triggered by
	StartReplication(policy *model.Policy, resource *model.Resource, trigger string) (int64, error)
	StopReplication(int64) error
	ListExecutions(...*models.ExecutionQuery) (int64, []*models.Execution, error)
	GetExecution(int64) (*models.Execution, error)
	ListTasks(...*models.TaskQuery) (int64, []*models.Task, error)
	GetTask(int64) (*models.Task, error)
	UpdateTaskStatus(id int64, status string, statusCondition ...string) error
	GetTaskLog(int64) ([]byte, error)
}

// NewController returns a controller implementation
func NewController(js job.Client) Controller {
	return &controller{
		executionMgr: execution.NewDefaultManager(),
		scheduler:    scheduler.NewScheduler(js),
		flowCtl:      flow.NewController(),
	}
}

type controller struct {
	flowCtl      flow.Controller
	executionMgr execution.Manager
	scheduler    scheduler.Scheduler
}

func (c *controller) StartReplication(policy *model.Policy, resource *model.Resource, trigger string) (int64, error) {
	if resource != nil && len(resource.Metadata.Vtags) != 1 {
		return 0, fmt.Errorf("the length of Vtags must be 1: %v", resource.Metadata.Vtags)
	}
	if len(trigger) == 0 {
		trigger = model.TriggerTypeManual
	}
	id, err := createExecution(c.executionMgr, policy.ID, trigger)
	if err != nil {
		return 0, err
	}

	flow := c.createFlow(id, policy, resource)
	if n, err := c.flowCtl.Start(flow); err != nil {
		// just update the status text, the status will be updated automatically
		// when listing the execution records
		if e := c.executionMgr.Update(&models.Execution{
			ID:         id,
			Status:     models.ExecutionStatusFailed,
			StatusText: err.Error(),
			Total:      n,
			Failed:     n,
		}, "Status", "StatusText", "Total", "Failed"); e != nil {
			log.Errorf("failed to update the execution %d: %v", id, e)
		}
		log.Errorf("the execution %d failed: %v", id, err)
	}

	return id, nil
}

// create different replication flows according to the input parameters
func (c *controller) createFlow(executionID int64, policy *model.Policy, resource *model.Resource) flow.Flow {
	// replicate the deletion operation, so create a deletion flow
	if resource != nil && resource.Deleted {
		return flow.NewDeletionFlow(c.executionMgr, c.scheduler, executionID, policy, []*model.Resource{resource})
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
				Type: model.FilterTypeTag,
				// only support replicate one tag
				Value: resource.Metadata.Vtags[0],
			},
		}
		filters = append(filters, policy.Filters...)
		policy.Filters = filters
	}
	return flow.NewCopyFlow(c.executionMgr, c.scheduler, executionID, policy)
}

func (c *controller) StopReplication(executionID int64) error {
	// TODO implement the function
	return nil
}
func (c *controller) ListExecutions(query ...*models.ExecutionQuery) (int64, []*models.Execution, error) {
	return c.executionMgr.List(query...)
}
func (c *controller) GetExecution(executionID int64) (*models.Execution, error) {
	return c.executionMgr.Get(executionID)
}
func (c *controller) ListTasks(query ...*models.TaskQuery) (int64, []*models.Task, error) {
	return c.executionMgr.ListTasks(query...)
}
func (c *controller) GetTask(id int64) (*models.Task, error) {
	return c.executionMgr.GetTask(id)
}
func (c *controller) UpdateTaskStatus(id int64, status string, statusCondition ...string) error {
	return c.executionMgr.UpdateTaskStatus(id, status, statusCondition...)
}
func (c *controller) GetTaskLog(taskID int64) ([]byte, error) {
	return c.executionMgr.GetTaskLog(taskID)
}

// create the execution record in database
func createExecution(mgr execution.Manager, policyID int64, trigger string) (int64, error) {
	id, err := mgr.Create(&models.Execution{
		PolicyID:  policyID,
		Trigger:   trigger,
		Status:    models.ExecutionStatusInProgress,
		StartTime: time.Now(),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create the execution record for replication based on policy %d: %v", policyID, err)
	}
	log.Debugf("an execution record for replication based on the policy %d created: %d", policyID, id)
	return id, nil
}
