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

package handler

import (
	"context"
	"github.com/goharbor/harbor/src/common/rbac"
	"strconv"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/controller/replication"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task"
	replica "github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/event"
	"github.com/goharbor/harbor/src/replication/policy"
	"github.com/goharbor/harbor/src/replication/policy/manager"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/replication"
)

func newReplicationAPI() *replicationAPI {
	return &replicationAPI{
		ctl:       replication.Ctl,
		policyMgr: manager.NewDefaultManager(),
	}
}

type replicationAPI struct {
	BaseAPI
	ctl       replication.Controller
	policyMgr policy.Controller
}

func (r *replicationAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	return nil
}

func (r *replicationAPI) StartReplication(ctx context.Context, params operation.StartReplicationParams) middleware.Responder {
	// TODO move the following logic to the replication controller after refactoring the policy management part with the new programming model
	if err := r.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceReplication); err != nil {
		return r.SendError(ctx, err)
	}
	policy, err := r.policyMgr.Get(params.Execution.PolicyID)
	if err != nil {
		return r.SendError(ctx, err)
	}
	if policy == nil {
		return r.SendError(ctx, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessage("the replication policy %d not found", params.Execution.PolicyID))
	}
	if err = event.PopulateRegistries(replica.RegistryMgr, policy); err != nil {
		return r.SendError(ctx, err)
	}

	// the legacy replication scheduler job("src/jobservice/job/impl/replication/scheduler.go") calls the start replication API
	// to trigger the scheduled replication, a query string "trigger" is added when sending the request
	// here is the logic to cover this part
	trigger := task.ExecutionTriggerManual
	if params.HTTPRequest.URL.Query().Get("trigger") == "scheduled" {
		trigger = task.ExecutionTriggerSchedule
	}

	executionID, err := r.ctl.Start(ctx, policy, nil, trigger)
	if err != nil {
		return r.SendError(ctx, err)
	}
	location := strings.TrimSuffix(params.HTTPRequest.URL.Path, "/") + "/" + strconv.FormatInt(executionID, 10)
	return operation.NewStartReplicationCreated().WithLocation(location)
}

func (r *replicationAPI) StopReplication(ctx context.Context, params operation.StopReplicationParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceReplication); err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.ctl.Stop(ctx, params.ID); err != nil {
		return r.SendError(ctx, err)
	}
	return nil
}

func (r *replicationAPI) ListReplicationExecutions(ctx context.Context, params operation.ListReplicationExecutionsParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceReplication); err != nil {
		return r.SendError(ctx, err)
	}
	query, err := r.BuildQuery(ctx, nil, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return r.SendError(ctx, err)
	}
	if params.PolicyID != nil {
		query.Keywords["PolicyID"] = *params.PolicyID
	}
	if params.Status != nil {
		status := *params.Status
		// as we convert the status when responding requests to keep the backward compatibility,
		// here we need to reverse-convert the status
		switch status {
		case "InProgress":
			status = job.RunningStatus.String()
		case "Succeed":
			status = job.SuccessStatus.String()
		case "Stopped":
			status = job.StoppedStatus.String()
		case "Failed":
			status = job.ErrorStatus.String()
		}
		query.Keywords["Status"] = status
	}
	if params.Trigger != nil {
		trigger := *params.Trigger
		// as we convert the trigger when responding requests to keep the backward compatibility,
		// here we need to reverse-convert the trigger
		switch trigger {
		case "manual":
			trigger = task.ExecutionTriggerManual
		case "scheduled":
			trigger = task.ExecutionTriggerSchedule
		case "event_based":
			trigger = task.ExecutionTriggerEvent
		}
		query.Keywords["Trigger"] = trigger
	}

	total, err := r.ctl.ExecutionCount(ctx, query)
	if err != nil {
		return r.SendError(ctx, err)
	}
	executions, err := r.ctl.ListExecutions(ctx, query)
	if err != nil {
		return r.SendError(ctx, err)
	}

	var execs []*models.ReplicationExecution
	for _, execution := range executions {
		execs = append(execs, convertExecution(execution))
	}

	return operation.NewListReplicationExecutionsOK().
		WithXTotalCount(total).
		WithLink(r.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(execs)
}

func (r *replicationAPI) GetReplicationExecution(ctx context.Context, params operation.GetReplicationExecutionParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceReplication); err != nil {
		return r.SendError(ctx, err)
	}
	execution, err := r.ctl.GetExecution(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewGetReplicationExecutionOK().WithPayload(convertExecution(execution))
}

func (r *replicationAPI) ListReplicationTasks(ctx context.Context, params operation.ListReplicationTasksParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceReplication); err != nil {
		return r.SendError(ctx, err)
	}
	query, err := r.BuildQuery(ctx, nil, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return r.SendError(ctx, err)
	}
	query.Keywords["ExecutionID"] = params.ID
	if params.Status != nil {
		var status interface{} = *params.Status
		// as we convert the status when responding requests to keep the backward compatibility,
		// here we need to reverse-convert the status
		// the status "pending" and "stopped" is same with jobservice, no need to convert
		switch status {
		case "InProgress":
			status = &q.OrList{
				Values: []interface{}{
					job.ScheduledStatus.String(),
					job.RunningStatus.String(),
				},
			}
		case "Succeed":
			status = job.SuccessStatus.String()
		case "Failed":
			status = job.ErrorStatus.String()
		}
		query.Keywords["Status"] = status
	}
	if params.ResourceType != nil {
		query.Keywords["ExtraAttrs.resource_type"] = *params.ResourceType
	}

	total, err := r.ctl.TaskCount(ctx, query)
	if err != nil {
		return r.SendError(ctx, err)
	}

	tasks, err := r.ctl.ListTasks(ctx, query)
	if err != nil {
		return r.SendError(ctx, err)
	}

	var tks []*models.ReplicationTask
	for _, task := range tasks {
		tks = append(tks, convertTask(task))
	}

	return operation.NewListReplicationTasksOK().
		WithXTotalCount(total).
		WithLink(r.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(tks)
}

func (r *replicationAPI) GetReplicationLog(ctx context.Context, params operation.GetReplicationLogParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceReplication); err != nil {
		return r.SendError(ctx, err)
	}
	execution, err := r.ctl.GetExecution(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, err)
	}
	task, err := r.ctl.GetTask(ctx, params.TaskID)
	if err != nil {
		return r.SendError(ctx, err)
	}
	if execution.ID != task.ExecutionID {
		return r.SendError(ctx, errors.New(nil).
			WithCode(errors.NotFoundCode).
			WithMessage("execution %d contains no task with ID %d", params.ID, params.TaskID))
	}
	log, err := r.ctl.GetTaskLog(ctx, params.TaskID)
	if err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewGetReplicationLogOK().WithContentType("text/plain").WithPayload(string(log))
}
func convertExecution(execution *replication.Execution) *models.ReplicationExecution {
	exec := &models.ReplicationExecution{
		ID:         execution.ID,
		PolicyID:   execution.PolicyID,
		StatusText: execution.StatusMessage,
		StartTime:  strfmt.DateTime(execution.StartTime),
		EndTime:    strfmt.DateTime(execution.EndTime),
	}
	// keep backward compatibility
	if execution.Metrics != nil {
		exec.Total = execution.Metrics.TaskCount
		exec.Succeed = execution.Metrics.SuccessTaskCount
		exec.Failed = execution.Metrics.ErrorTaskCount
		exec.InProgress = execution.Metrics.PendingTaskCount +
			execution.Metrics.ScheduledTaskCount + execution.Metrics.RunningTaskCount
		exec.Stopped = execution.Metrics.StoppedTaskCount
	}
	switch execution.Trigger {
	case task.ExecutionTriggerManual:
		exec.Trigger = "manual"
	case task.ExecutionTriggerSchedule:
		exec.Trigger = "scheduled"
	case task.ExecutionTriggerEvent:
		exec.Trigger = "event_based"
	}
	switch execution.Status {
	case job.RunningStatus.String():
		exec.Status = "InProgress"
	case job.SuccessStatus.String():
		exec.Status = "Succeed"
	case job.StoppedStatus.String():
		exec.Status = "Stopped"
	case job.ErrorStatus.String():
		exec.Status = "Failed"
	}

	return exec
}

func convertTask(task *replication.Task) *models.ReplicationTask {
	tk := &models.ReplicationTask{
		ID:           task.ID,
		ExecutionID:  task.ExecutionID,
		JobID:        task.JobID,
		Operation:    task.Operation,
		ResourceType: task.ResourceType,
		SrcResource:  task.SourceResource,
		DstResource:  task.DestinationResource,
		StartTime:    strfmt.DateTime(task.StartTime),
		EndTime:      strfmt.DateTime(task.EndTime),
	}
	// keep backward compatibility
	switch task.Status {
	case job.ScheduledStatus.String(), job.RunningStatus.String():
		tk.Status = "InProgress"
	case job.SuccessStatus.String():
		tk.Status = "Succeed"
	case job.ErrorStatus.String():
		tk.Status = "Failed"
	// the status "pending" and "stopped" is same with jobservice, no need to convert
	default:
		tk.Status = task.Status
	}
	return tk
}
