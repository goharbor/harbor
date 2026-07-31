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
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/common/rbac"
	modelimportctl "github.com/goharbor/harbor/src/controller/modelimport"
	projectctl "github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	modelimportpkg "github.com/goharbor/harbor/src/pkg/modelimport"
	"github.com/goharbor/harbor/src/pkg/modelimport/huggingface"
	importmodel "github.com/goharbor/harbor/src/pkg/modelimport/model"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/model_import"
)

func newModelImportAPI() *modelImportAPI {
	return &modelImportAPI{
		modelImportCtl: modelimportctl.Ctl,
		projectCtl:     projectctl.Ctl,
		hfClient:       huggingface.NewClient(),
	}
}

var _ restapi.ModelImportAPI = (*modelImportAPI)(nil)

type modelImportAPI struct {
	BaseAPI
	modelImportCtl modelimportctl.Controller
	projectCtl     projectctl.Controller
	hfClient       huggingface.Client
}

func (api *modelImportAPI) CreateModelImportPolicy(ctx context.Context, params operation.CreateModelImportPolicyParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	policy, err := api.convertPolicy(ctx, params.ProjectName, params.Policy)
	if err != nil {
		return api.SendError(ctx, err)
	}
	if _, err = api.modelImportCtl.CreatePolicy(ctx, policy); err != nil {
		return api.SendError(ctx, err)
	}
	location := fmt.Sprintf("%s/%s", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), policy.Name)
	return operation.NewCreateModelImportPolicyCreated().WithLocation(location)
}

func (api *modelImportAPI) ListModelImportPolicies(ctx context.Context, params operation.ListModelImportPoliciesParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	project, err := api.projectCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return api.SendError(ctx, err)
	}
	query, err := api.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return api.SendError(ctx, err)
	}
	query.Keywords["project_id"] = project.ProjectID
	total, err := api.modelImportCtl.CountPolicies(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}
	policies, err := api.modelImportCtl.ListPolicies(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}
	payload := make([]*models.ModelImportPolicy, 0, len(policies))
	for _, policy := range policies {
		payload = append(payload, modelImportPolicyToPayload(policy))
	}
	return operation.NewListModelImportPoliciesOK().WithPayload(payload).WithXTotalCount(total).
		WithLink(api.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String())
}

func (api *modelImportAPI) GetModelImportPolicy(ctx context.Context, params operation.GetModelImportPolicyParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	policy, err := api.policyByName(ctx, params.ProjectName, params.ModelImportPolicyName)
	if err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewGetModelImportPolicyOK().WithPayload(modelImportPolicyToPayload(policy))
}

func (api *modelImportAPI) UpdateModelImportPolicy(ctx context.Context, params operation.UpdateModelImportPolicyParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionUpdate, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	existing, err := api.policyByName(ctx, params.ProjectName, params.ModelImportPolicyName)
	if err != nil {
		return api.SendError(ctx, err)
	}
	policy, err := api.convertPolicy(ctx, params.ProjectName, params.Policy)
	if err != nil {
		return api.SendError(ctx, err)
	}
	policy.ID = existing.ID
	if err = api.modelImportCtl.UpdatePolicy(ctx, policy); err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewUpdateModelImportPolicyOK()
}

func (api *modelImportAPI) DeleteModelImportPolicy(ctx context.Context, params operation.DeleteModelImportPolicyParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionDelete, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	policy, err := api.policyByName(ctx, params.ProjectName, params.ModelImportPolicyName)
	if err != nil {
		return api.SendError(ctx, err)
	}
	if err = api.modelImportCtl.DeletePolicy(ctx, policy.ID); err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewDeleteModelImportPolicyOK()
}

func (api *modelImportAPI) RunModelImportPolicy(ctx context.Context, params operation.RunModelImportPolicyParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionPush, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	policy, err := api.policyByName(ctx, params.ProjectName, params.ModelImportPolicyName)
	if err != nil {
		return api.SendError(ctx, err)
	}
	executionID, err := api.modelImportCtl.Start(ctx, policy.ID, task.ExecutionTriggerManual)
	if err != nil {
		return api.SendError(ctx, err)
	}
	location := fmt.Sprintf("/projects/%s/model-import/policies/%s/executions/%d", params.ProjectName, params.ModelImportPolicyName, executionID)
	return operation.NewRunModelImportPolicyCreated().WithLocation(location)
}

func (api *modelImportAPI) RunModelImportExecution(ctx context.Context, params operation.RunModelImportExecutionParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionPush, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	req, err := api.convertExecutionRequest(ctx, params.ProjectName, params.Request)
	if err != nil {
		return api.SendError(ctx, err)
	}
	executionID, err := api.modelImportCtl.StartImport(ctx, req, task.ExecutionTriggerManual)
	if err != nil {
		return api.SendError(ctx, err)
	}
	location := fmt.Sprintf("/projects/%s/model-import/executions/%d", params.ProjectName, executionID)
	return operation.NewRunModelImportExecutionCreated().WithLocation(location)
}

func (api *modelImportAPI) ListModelImportExecutions(ctx context.Context, params operation.ListModelImportExecutionsParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	policy, err := api.policyByName(ctx, params.ProjectName, params.ModelImportPolicyName)
	if err != nil {
		return api.SendError(ctx, err)
	}
	query, err := api.BuildQuery(ctx, nil, nil, params.Page, params.PageSize)
	if err != nil {
		return api.SendError(ctx, err)
	}
	query.Keywords = map[string]any{
		"vendor_type": job.ModelImportVendorType,
		"vendor_id":   policy.ID,
	}
	total, err := api.modelImportCtl.ExecutionCount(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}
	executions, err := api.modelImportCtl.ListExecutions(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}
	payload := make([]*models.Execution, 0, len(executions))
	for _, exec := range executions {
		payload = append(payload, executionToPayload(exec))
	}
	return operation.NewListModelImportExecutionsOK().WithPayload(payload).WithXTotalCount(total).
		WithLink(api.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String())
}

func (api *modelImportAPI) GetModelImportExecution(ctx context.Context, params operation.GetModelImportExecutionParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	_, exec, err := api.authorizeExecution(ctx, params.ProjectName, params.ModelImportPolicyName, params.ExecutionID)
	if err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewGetModelImportExecutionOK().WithPayload(executionToPayload(exec))
}

func (api *modelImportAPI) ListModelImportTasks(ctx context.Context, params operation.ListModelImportTasksParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	if _, _, err := api.authorizeExecution(ctx, params.ProjectName, params.ModelImportPolicyName, params.ExecutionID); err != nil {
		return api.SendError(ctx, err)
	}
	query, err := api.BuildQuery(ctx, nil, nil, params.Page, params.PageSize)
	if err != nil {
		return api.SendError(ctx, err)
	}
	query.Keywords = map[string]any{
		"vendor_type":  job.ModelImportVendorType,
		"execution_id": params.ExecutionID,
	}
	tasks, err := api.modelImportCtl.ListTasks(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}
	payload := make([]*models.Task, 0, len(tasks))
	for _, t := range tasks {
		payload = append(payload, taskToPayload(t))
	}
	return operation.NewListModelImportTasksOK().WithPayload(payload)
}

func (api *modelImportAPI) GetModelImportLog(ctx context.Context, params operation.GetModelImportLogParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	if _, _, err := api.authorizeExecution(ctx, params.ProjectName, params.ModelImportPolicyName, params.ExecutionID); err != nil {
		return api.SendError(ctx, err)
	}
	t, err := api.modelImportCtl.GetTask(ctx, params.TaskID)
	if err != nil {
		return api.SendError(ctx, err)
	}
	if t.ExecutionID != params.ExecutionID {
		return api.SendError(ctx, errors.NotFoundError(nil).WithMessagef("model import task %d not found", params.TaskID))
	}
	content, err := api.modelImportCtl.GetTaskLog(ctx, params.TaskID)
	if err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewGetModelImportLogOK().WithPayload(string(content))
}

func (api *modelImportAPI) ValidateModelImportSource(ctx context.Context, params operation.ValidateModelImportSourceParams) middleware.Responder {
	if err := api.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourceRepository); err != nil {
		return api.SendError(ctx, err)
	}
	if params.Source == nil {
		return api.SendError(ctx, errors.BadRequestError(nil).WithMessage("source is required"))
	}
	revision := params.Source.HfRevision
	if revision == "" {
		revision = "main"
	}
	snapshot, err := api.hfClient.Snapshot(ctx, params.Source.HfRepoID, revision, params.Source.HfSubpath, params.Source.HfToken)
	if err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewValidateModelImportSourceOK().WithPayload(&models.ModelImportValidation{
		HfRepoID:   snapshot.RepoID,
		HfRevision: snapshot.Revision,
		CommitSha:  snapshot.CommitSHA,
		SourceURL:  snapshot.SourceURL,
		FileCount:  int64(len(snapshot.Files)),
		CardData:   snapshot.CardData,
	})
}

func (api *modelImportAPI) convertPolicy(ctx context.Context, projectName string, payload *models.ModelImportPolicy) (*importmodel.Policy, error) {
	if payload == nil {
		return nil, errors.BadRequestError(nil).WithMessage("policy is required")
	}
	if ok, err := regexp.MatchString(nameRegex, payload.Name); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("name %s is invalid", payload.Name)
	}
	project, err := api.projectCtl.GetByName(ctx, projectName)
	if err != nil {
		return nil, err
	}
	securityCtx, _ := api.GetSecurityContext(ctx)
	creator := payload.Creator
	if creator == "" && securityCtx != nil {
		creator = securityCtx.GetUsername()
	}
	return &importmodel.Policy{
		ID:                  payload.ID,
		Name:                payload.Name,
		Description:         payload.Description,
		ProjectID:           project.ProjectID,
		TargetRepository:    payload.TargetRepository,
		TargetTag:           payload.TargetTag,
		HuggingFaceRepoID:   payload.HfRepoID,
		HuggingFaceRevision: payload.HfRevision,
		HuggingFaceSubpath:  payload.HfSubpath,
		PlainHFToken:        payload.HfToken,
		LastSyncedCommit:    payload.LastSyncedCommit,
		LastArtifactDigest:  payload.LastArtifactDigest,
		Creator:             creator,
		Enabled:             payload.Enabled,
		TriggerStr:          payload.Trigger,
		ExtraAttrsStr:       payload.ExtraAttrs,
		CreationTime:        time.Time(payload.CreationTime),
		UpdateTime:          time.Time(payload.UpdateTime),
	}, nil
}

func (api *modelImportAPI) convertExecutionRequest(ctx context.Context, projectName string, payload *models.ModelImportExecutionRequest) (*modelimportpkg.ImportRequest, error) {
	if payload == nil {
		return nil, errors.BadRequestError(nil).WithMessage("model import request is required")
	}
	project, err := api.projectCtl.GetByName(ctx, projectName)
	if err != nil {
		return nil, err
	}
	return &modelimportpkg.ImportRequest{
		ProjectID:          project.ProjectID,
		TargetRepository:   payload.TargetRepository,
		TargetTag:          payload.TargetTag,
		HuggingFaceRepoID:  payload.HfRepoID,
		Revision:           payload.HfRevision,
		Subpath:            payload.HfSubpath,
		Token:              payload.HfToken,
		LastSyncedCommit:   payload.LastSyncedCommit,
		LastArtifactDigest: payload.LastArtifactDigest,
	}, nil
}

func (api *modelImportAPI) policyByName(ctx context.Context, projectName, policyName string) (*importmodel.Policy, error) {
	project, err := api.projectCtl.GetByName(ctx, projectName)
	if err != nil {
		return nil, err
	}
	return api.modelImportCtl.GetPolicyByName(ctx, project.ProjectID, policyName)
}

// authorizeExecution resolves the named policy and verifies that the
// caller-supplied execution ID actually belongs to it. Without this check an
// execution/task/log endpoint would leak resources across policies and
// projects, since the execution and task IDs are client-supplied and only
// scoped to the MODEL_IMPORT vendor type by the controller.
func (api *modelImportAPI) authorizeExecution(ctx context.Context, projectName, policyName string, executionID int64) (*importmodel.Policy, *task.Execution, error) {
	policy, err := api.policyByName(ctx, projectName, policyName)
	if err != nil {
		return nil, nil, err
	}
	exec, err := api.modelImportCtl.GetExecution(ctx, executionID)
	if err != nil {
		return nil, nil, err
	}
	if exec.VendorID != policy.ID {
		return nil, nil, errors.NotFoundError(nil).WithMessagef("model import execution %d not found", executionID)
	}
	return policy, exec, nil
}

func modelImportPolicyToPayload(policy *importmodel.Policy) *models.ModelImportPolicy {
	if policy == nil {
		return nil
	}
	return &models.ModelImportPolicy{
		ID:                 policy.ID,
		Name:               policy.Name,
		Description:        policy.Description,
		ProjectID:          policy.ProjectID,
		TargetRepository:   policy.TargetRepository,
		TargetTag:          policy.TargetTag,
		HfRepoID:           policy.HuggingFaceRepoID,
		HfRevision:         policy.HuggingFaceRevision,
		HfSubpath:          policy.HuggingFaceSubpath,
		LastSyncedCommit:   policy.LastSyncedCommit,
		LastArtifactDigest: policy.LastArtifactDigest,
		Creator:            policy.Creator,
		Enabled:            policy.Enabled,
		Trigger:            policy.TriggerStr,
		ExtraAttrs:         policy.ExtraAttrsStr,
		CreationTime:       strfmt.DateTime(policy.CreationTime),
		UpdateTime:         strfmt.DateTime(policy.UpdateTime),
	}
}

func executionToPayload(exec *task.Execution) *models.Execution {
	if exec == nil {
		return nil
	}
	payload := &models.Execution{
		ID:            exec.ID,
		VendorType:    exec.VendorType,
		VendorID:      exec.VendorID,
		Status:        exec.Status,
		StatusMessage: exec.StatusMessage,
		Trigger:       exec.Trigger,
		ExtraAttrs:    exec.ExtraAttrs,
		StartTime:     exec.StartTime.Format(time.RFC3339),
		EndTime:       exec.EndTime.Format(time.RFC3339),
	}
	if exec.Metrics != nil {
		payload.Metrics = &models.Metrics{
			TaskCount:          exec.Metrics.TaskCount,
			SuccessTaskCount:   exec.Metrics.SuccessTaskCount,
			ErrorTaskCount:     exec.Metrics.ErrorTaskCount,
			PendingTaskCount:   exec.Metrics.PendingTaskCount,
			RunningTaskCount:   exec.Metrics.RunningTaskCount,
			ScheduledTaskCount: exec.Metrics.ScheduledTaskCount,
			StoppedTaskCount:   exec.Metrics.StoppedTaskCount,
		}
	}
	return payload
}

func taskToPayload(t *task.Task) *models.Task {
	if t == nil {
		return nil
	}
	return &models.Task{
		ID:            t.ID,
		ExecutionID:   t.ExecutionID,
		Status:        t.Status,
		StatusMessage: t.StatusMessage,
		RunCount:      t.RunCount,
		ExtraAttrs:    t.ExtraAttrs,
		CreationTime:  t.CreationTime.Format(time.RFC3339),
		UpdateTime:    t.UpdateTime.Format(time.RFC3339),
		StartTime:     t.StartTime.Format(time.RFC3339),
		EndTime:       t.EndTime.Format(time.RFC3339),
	}
}
