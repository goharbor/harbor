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

package webhook

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	"github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/pkg/task"
)

var (
	// Ctl is a global webhook controller instance
	Ctl = NewController()

	// webhookJobVendors represents webhook(http) or slack.
	webhookJobVendors = q.NewOrList([]interface{}{job.WebhookJobVendorType, job.SlackJobVendorType, job.TeamsJobVendorType})
)

type Controller interface {
	// CreatePolicy creates webhook policy
	CreatePolicy(ctx context.Context, policy *model.Policy) (int64, error)
	// ListPolicies lists webhook policies filter by query
	ListPolicies(ctx context.Context, query *q.Query) ([]*model.Policy, error)
	// CountPolicies counts webhook policies filter by query
	CountPolicies(ctx context.Context, query *q.Query) (int64, error)
	// GetPolicy gets webhook policy by specified ID
	GetPolicy(ctx context.Context, id int64) (*model.Policy, error)
	// UpdatePolicy updates webhook policy
	UpdatePolicy(ctx context.Context, policy *model.Policy) error
	// DeletePolicy deletes webhook policy by specified ID
	DeletePolicy(ctx context.Context, policyID int64) error
	// GetRelatedPolices gets related policies by the input project id and event type
	GetRelatedPolices(ctx context.Context, projectID int64, eventType string) ([]*model.Policy, error)

	// CountExecutions counts executions under the webhook policy
	CountExecutions(ctx context.Context, policyID int64, query *q.Query) (int64, error)
	// ListExecutions lists executions under the webhook policy
	ListExecutions(ctx context.Context, policyID int64, query *q.Query) ([]*task.Execution, error)
	// CountTasks counts tasks under the webhook execution
	CountTasks(ctx context.Context, execID int64, query *q.Query) (int64, error)
	// ListTasks lists tasks under the webhook execution
	ListTasks(ctx context.Context, execID int64, query *q.Query) ([]*task.Task, error)
	// GetTask gets the webhook task by the specified ID
	GetTask(ctx context.Context, taskID int64) (*task.Task, error)
	// GetTaskLog gets task log
	GetTaskLog(ctx context.Context, taskID int64) ([]byte, error)

	// GetLastTriggerTime gets policy last trigger time group by event type
	GetLastTriggerTime(ctx context.Context, eventType string, policyID int64) (time.Time, error)
}

type controller struct {
	policyMgr policy.Manager
	execMgr   task.ExecutionManager
	taskMgr   task.Manager
}

func NewController() Controller {
	return &controller{
		policyMgr: policy.Mgr,
		execMgr:   task.ExecMgr,
		taskMgr:   task.Mgr,
	}
}

func (c *controller) CreatePolicy(ctx context.Context, policy *model.Policy) (int64, error) {
	return c.policyMgr.Create(ctx, policy)
}

func (c *controller) ListPolicies(ctx context.Context, query *q.Query) ([]*model.Policy, error) {
	return c.policyMgr.List(ctx, query)
}

func (c *controller) CountPolicies(ctx context.Context, query *q.Query) (int64, error) {
	return c.policyMgr.Count(ctx, query)
}

func (c *controller) GetPolicy(ctx context.Context, id int64) (*model.Policy, error) {
	return c.policyMgr.Get(ctx, id)
}

func (c *controller) UpdatePolicy(ctx context.Context, policy *model.Policy) error {
	return c.policyMgr.Update(ctx, policy)
}

func (c *controller) DeletePolicy(ctx context.Context, policyID int64) error {
	// delete executions under the webhook policy,
	// there are three vendor types(webhook, slack & teams) needs to be deleted.
	if err := c.execMgr.DeleteByVendor(ctx, job.WebhookJobVendorType, policyID); err != nil {
		return errors.Wrapf(err, "failed to delete executions for webhook of policy %d", policyID)
	}
	if err := c.execMgr.DeleteByVendor(ctx, job.SlackJobVendorType, policyID); err != nil {
		return errors.Wrapf(err, "failed to delete executions for slack of policy %d", policyID)
	}
	if err := c.execMgr.DeleteByVendor(ctx, job.TeamsJobVendorType, policyID); err != nil {
		return errors.Wrapf(err, "failed to delete executions for teams of policy %d", policyID)
	}

	return c.policyMgr.Delete(ctx, policyID)
}

func (c *controller) GetRelatedPolices(ctx context.Context, projectID int64, eventType string) ([]*model.Policy, error) {
	return c.policyMgr.GetRelatedPolices(ctx, projectID, eventType)
}

func (c *controller) CountExecutions(ctx context.Context, policyID int64, query *q.Query) (int64, error) {
	return c.execMgr.Count(ctx, buildExecutionQuery(policyID, query))
}

func (c *controller) ListExecutions(ctx context.Context, policyID int64, query *q.Query) ([]*task.Execution, error) {
	return c.execMgr.List(ctx, buildExecutionQuery(policyID, query))
}

func (c *controller) CountTasks(ctx context.Context, execID int64, query *q.Query) (int64, error) {
	return c.taskMgr.Count(ctx, buildTaskQuery(execID, query))
}

func (c *controller) ListTasks(ctx context.Context, execID int64, query *q.Query) ([]*task.Task, error) {
	return c.taskMgr.List(ctx, buildTaskQuery(execID, query))
}

func (c *controller) GetTask(ctx context.Context, taskID int64) (*task.Task, error) {
	query := q.New(q.KeyWords{
		"id":          taskID,
		"vendor_type": webhookJobVendors,
	})
	tasks, err := c.taskMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessagef("webhook task %d not found", taskID)
	}
	return tasks[0], nil
}

func (c *controller) GetTaskLog(ctx context.Context, taskID int64) ([]byte, error) {
	// ensure the webhook task exist
	_, err := c.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return c.taskMgr.GetLog(ctx, taskID)
}

func buildExecutionQuery(policyID int64, query *q.Query) *q.Query {
	query = q.MustClone(query)
	query.Keywords["vendor_type"] = webhookJobVendors
	query.Keywords["vendor_id"] = policyID
	return query
}

func buildTaskQuery(execID int64, query *q.Query) *q.Query {
	query = q.MustClone(query)
	query.Keywords["vendor_type"] = webhookJobVendors
	query.Keywords["execution_id"] = execID
	return query
}

func (c *controller) GetLastTriggerTime(ctx context.Context, eventType string, policyID int64) (time.Time, error) {
	query := q.New(q.KeyWords{
		"vendor_type":           webhookJobVendors,
		"vendor_id":             policyID,
		"ExtraAttrs.event_type": eventType,
	})
	// fetch the latest execution sort by start_time
	execs, err := c.execMgr.List(ctx, query.First(q.NewSort("start_time", true)))
	if err != nil {
		return time.Time{}, err
	}

	if len(execs) > 0 {
		return execs[0].StartTime, nil
	}

	return time.Time{}, nil
}
