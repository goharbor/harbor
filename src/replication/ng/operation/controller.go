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
	"strings"
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
	StartReplication(policy *model.Policy, resource *model.Resource, trigger model.TriggerType) (int64, error)
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

func (c *controller) StartReplication(policy *model.Policy, resource *model.Resource, trigger model.TriggerType) (int64, error) {
	// only support one tag if the resource is specified as we append the tag name as a filter
	// when creating the flow in function "createFlow"
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
		// only update the execution when got error.
		// if got no error, it will be updated automatically
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
	// copy only one resource, add extra filters to the policy to make sure
	// only the specified resource will be filtered out
	if resource != nil {
		filters := []*model.Filter{
			{
				Type:  model.FilterTypeResource,
				Value: resource.Type,
			},
			{
				Type: model.FilterTypeName,
				// TODO only filter the repo part?
				Value: resource.Metadata.GetResourceName(),
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
	_, tasks, err := c.ListTasks(&models.TaskQuery{
		ExecutionID: executionID,
	})
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if !isTaskRunning(task) {
			log.Debugf("the task %d(job ID: %s) isn't running, its status is %s, skip", task.ID, task.JobID, task.Status)
			continue
		}
		if err = c.scheduler.Stop(task.JobID); err != nil {
			if isNotRunningJobError(err) {
				log.Warningf("got not running job error when trying stop the task %d(job ID: %s): %v, skip", task.ID, task.JobID, err)
				continue
			}
			return err
		}
		log.Debugf("the stop request for task %d(job ID: %s) sent", task.ID, task.JobID)
	}
	return nil
}

func isTaskRunning(task *models.Task) bool {
	if task == nil {
		return false
	}
	switch task.Status {
	case models.TaskStatusSucceed,
		models.TaskStatusStopped,
		models.TaskStatusFailed:
		return false
	}
	return true
}

// when trying to stop a job which isn't running in jobservice,
// an error whose message contains "xxx is not a running job"
// will be returned
func isNotRunningJobError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "is not a running job")
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
func createExecution(mgr execution.Manager, policyID int64, trigger model.TriggerType) (int64, error) {
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
