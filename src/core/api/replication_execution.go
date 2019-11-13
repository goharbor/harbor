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

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/event"
	"github.com/goharbor/harbor/src/replication/model"
)

// ReplicationOperationAPI handles the replication operation requests
type ReplicationOperationAPI struct {
	BaseController
	execution *models.Execution
	task      *models.Task
}

// Prepare ...
func (r *ReplicationOperationAPI) Prepare() {
	r.BaseController.Prepare()
	// As we delegate the jobservice to trigger the scheduled replication,
	// we need to allow the jobservice to call the API
	if !(r.SecurityCtx.IsSysAdmin() || r.SecurityCtx.IsSolutionUser()) {
		if !r.SecurityCtx.IsAuthenticated() {
			r.SendUnAuthorizedError(errors.New("UnAuthorized"))
			return
		}
		r.SendForbiddenError(errors.New(r.SecurityCtx.GetUsername()))
		return
	}

	// check the existence of execution if execution ID is provided in the request path
	executionIDStr := r.GetStringFromPath(":id")
	if len(executionIDStr) > 0 {
		executionID, err := strconv.ParseInt(executionIDStr, 10, 64)
		if err != nil || executionID <= 0 {
			r.SendBadRequestError(fmt.Errorf("invalid execution ID: %s", executionIDStr))
			return
		}
		execution, err := replication.OperationCtl.GetExecution(executionID)
		if err != nil {
			r.SendInternalServerError(fmt.Errorf("failed to get execution %d: %v", executionID, err))
			return
		}
		if execution == nil {
			r.SendNotFoundError(fmt.Errorf("execution %d not found", executionID))
			return
		}
		r.execution = execution

		// check the existence of task if task ID is provided in the request path
		taskIDStr := r.GetStringFromPath(":tid")
		if len(taskIDStr) > 0 {
			taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
			if err != nil || taskID <= 0 {
				r.SendBadRequestError(fmt.Errorf("invalid task ID: %s", taskIDStr))
				return
			}
			task, err := replication.OperationCtl.GetTask(taskID)
			if err != nil {
				r.SendInternalServerError(fmt.Errorf("failed to get task %d: %v", taskID, err))
				return
			}
			if task == nil || task.ExecutionID != executionID {
				r.SendNotFoundError(fmt.Errorf("task %d not found", taskID))
				return
			}
			r.task = task
		}
	}
}

// The API is open only for system admin currently, we can use
// the code commentted below to make the API available to the
// users who have permission for all projects that the policy
// refers
/*
func (r *ReplicationOperationAPI) authorized(policy *model.Policy, resource rbac.Resource, action rbac.Action) bool {

	projects := []string{}
	// pull mode
	if policy.SrcRegistryID != 0 {
		projects = append(projects, policy.DestNamespace)
	} else {
		// push mode
		projects = append(projects, policy.SrcNamespaces...)
	}

	for _, project := range projects {
		resource := rbac.NewProjectNamespace(project).Resource(resource)
		if !r.SecurityCtx.Can(action, resource) {
			r.HandleForbidden(r.SecurityCtx.GetUsername())
			return false
		}
	}

	return true
}
*/

// ListExecutions ...
func (r *ReplicationOperationAPI) ListExecutions() {
	query := &models.ExecutionQuery{
		Trigger: r.GetString("trigger"),
	}

	if len(r.GetString("status")) > 0 {
		query.Statuses = []string{r.GetString("status")}
	}
	if len(r.GetString("policy_id")) > 0 {
		policyID, err := r.GetInt64("policy_id")
		if err != nil || policyID <= 0 {
			r.SendBadRequestError(fmt.Errorf("invalid policy_id %s", r.GetString("policy_id")))
			return
		}
		query.PolicyID = policyID
	}
	page, size, err := r.GetPaginationParams()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	query.Page = page
	query.Size = size

	total, executions, err := replication.OperationCtl.ListExecutions(query)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to list executions: %v", err))
		return
	}
	r.SetPaginationHeader(total, query.Page, query.Size)
	r.WriteJSONData(executions)
}

// CreateExecution starts a replication
func (r *ReplicationOperationAPI) CreateExecution() {
	execution := &models.Execution{}
	if err := r.DecodeJSONReq(execution); err != nil {
		r.SendBadRequestError(err)
		return
	}

	policy, err := replication.PolicyCtl.Get(execution.PolicyID)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to get policy %d: %v", execution.PolicyID, err))
		return
	}
	if policy == nil {
		r.SendNotFoundError(fmt.Errorf("policy %d not found", execution.PolicyID))
		return
	}
	if !policy.Enabled {
		r.SendBadRequestError(fmt.Errorf("the policy %d is disabled", execution.PolicyID))
		return
	}
	if err = event.PopulateRegistries(replication.RegistryMgr, policy); err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to populate registries for policy %d: %v", execution.PolicyID, err))
		return
	}

	trigger := r.GetString("trigger", string(model.TriggerTypeManual))
	executionID, err := replication.OperationCtl.StartReplication(policy, nil, model.TriggerType(trigger))
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to start replication for policy %d: %v", execution.PolicyID, err))
		return
	}
	r.Redirect(http.StatusCreated, strconv.FormatInt(executionID, 10))
}

// GetExecution gets one execution of the replication
func (r *ReplicationOperationAPI) GetExecution() {
	r.WriteJSONData(r.execution)
}

// StopExecution stops one execution of the replication
func (r *ReplicationOperationAPI) StopExecution() {
	if err := replication.OperationCtl.StopReplication(r.execution.ID); err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to stop execution %d: %v", r.execution.ID, err))
		return
	}
}

// ListTasks ...
func (r *ReplicationOperationAPI) ListTasks() {
	query := &models.TaskQuery{
		ExecutionID:  r.execution.ID,
		ResourceType: r.GetString("resource_type"),
	}
	status := r.GetString("status")
	if len(status) > 0 {
		query.Statuses = []string{status}
	}
	page, size, err := r.GetPaginationParams()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	query.Page = page
	query.Size = size
	total, tasks, err := replication.OperationCtl.ListTasks(query)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to list tasks: %v", err))
		return
	}
	r.SetPaginationHeader(total, query.Page, query.Size)
	r.WriteJSONData(tasks)
}

// GetTaskLog ...
func (r *ReplicationOperationAPI) GetTaskLog() {
	logBytes, err := replication.OperationCtl.GetTaskLog(r.task.ID)
	if err != nil {
		if httpErr, ok := err.(*common_http.Error); ok {
			if ok && httpErr.Code == http.StatusNotFound {
				r.SendNotFoundError(fmt.Errorf("the log of task %d not found", r.task.ID))
				return
			}
		}
		r.SendInternalServerError(fmt.Errorf("failed to get log of task %d: %v", r.task.ID, err))
		return
	}
	r.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(len(logBytes)))
	r.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")
	_, err = r.Ctx.ResponseWriter.Write(logBytes)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to write log of task %d: %v", r.task.ID, err))
		return
	}
}
