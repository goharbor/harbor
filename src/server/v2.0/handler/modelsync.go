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
	"strconv"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/modelsync"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	pkgmodel "github.com/goharbor/harbor/src/pkg/modelsync/policy/model"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/model_sync"
)

func newModelSyncAPI() *modelSyncAPI {
	return &modelSyncAPI{
		ctl: modelsync.Ctl,
	}
}

type modelSyncAPI struct {
	BaseAPI
	ctl modelsync.Controller
}

func (m *modelSyncAPI) Prepare(_ context.Context, _ string, _ any) middleware.Responder {
	return nil
}

func (m *modelSyncAPI) toPolicy(in *models.ModelSyncPolicy) *modelsync.Policy {
	p := &modelsync.Policy{
		ID:             in.ID,
		Name:           in.Name,
		Description:    in.Description,
		Enabled:        in.Enabled,
		Registry:       &model.Registry{ID: in.RegistryID},
		SrcRepository:  in.SrcRepository,
		SrcRevision:    in.SrcRevision,
		FileFilter:     toFileFilter(in.FileFilters, in.FileExcludes),
		DestProjectID:  in.DestProjectID,
		DestRepository: in.DestRepository,
		TriggerType:    pkgmodel.TriggerTypeManual,
	}
	// accept the nested registry object as an alternative to registry_id
	if in.RegistryID == 0 && in.Registry != nil {
		p.Registry.ID = in.Registry.ID
	}
	if in.Trigger != nil {
		if in.Trigger.Type != "" {
			p.TriggerType = in.Trigger.Type
		}
		if in.Trigger.TriggerSettings != nil {
			p.Cron = in.Trigger.TriggerSettings.Cron
		}
	}
	return p
}

func (m *modelSyncAPI) CreateModelSyncPolicy(ctx context.Context, params operation.CreateModelSyncPolicyParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceModelSyncPolicy); err != nil {
		return m.SendError(ctx, err)
	}
	sc, err := m.GetSecurityContext(ctx)
	if err != nil {
		return m.SendError(ctx, err)
	}
	policy := m.toPolicy(params.Policy)
	policy.ID = 0
	policy.Creator = sc.GetUsername()
	id, err := m.ctl.CreatePolicy(ctx, policy)
	if err != nil {
		return m.SendError(ctx, err)
	}
	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), id)
	return operation.NewCreateModelSyncPolicyCreated().WithLocation(location)
}

func (m *modelSyncAPI) UpdateModelSyncPolicy(ctx context.Context, params operation.UpdateModelSyncPolicyParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceModelSyncPolicy); err != nil {
		return m.SendError(ctx, err)
	}
	existing, err := m.ctl.GetPolicy(ctx, params.ID)
	if err != nil {
		return m.SendError(ctx, err)
	}
	policy := m.toPolicy(params.Policy)
	policy.ID = params.ID
	policy.Creator = existing.Creator
	if err := m.ctl.UpdatePolicy(ctx, policy); err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewUpdateModelSyncPolicyOK()
}

func (m *modelSyncAPI) ListModelSyncPolicies(ctx context.Context, params operation.ListModelSyncPoliciesParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceModelSyncPolicy); err != nil {
		return m.SendError(ctx, err)
	}
	query, err := m.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return m.SendError(ctx, err)
	}
	total, err := m.ctl.PolicyCount(ctx, query)
	if err != nil {
		return m.SendError(ctx, err)
	}
	policies, err := m.ctl.ListPolicies(ctx, query)
	if err != nil {
		return m.SendError(ctx, err)
	}
	result := make([]*models.ModelSyncPolicy, 0, len(policies))
	for _, p := range policies {
		result = append(result, convertModelSyncPolicy(p))
	}
	return operation.NewListModelSyncPoliciesOK().
		WithXTotalCount(total).
		WithLink(m.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(result)
}

func (m *modelSyncAPI) GetModelSyncPolicy(ctx context.Context, params operation.GetModelSyncPolicyParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceModelSyncPolicy); err != nil {
		return m.SendError(ctx, err)
	}
	policy, err := m.ctl.GetPolicy(ctx, params.ID)
	if err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewGetModelSyncPolicyOK().WithPayload(convertModelSyncPolicy(policy))
}

func (m *modelSyncAPI) DeleteModelSyncPolicy(ctx context.Context, params operation.DeleteModelSyncPolicyParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionDelete, rbac.ResourceModelSyncPolicy); err != nil {
		return m.SendError(ctx, err)
	}
	if err := m.ctl.DeletePolicy(ctx, params.ID); err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewDeleteModelSyncPolicyOK()
}

func (m *modelSyncAPI) StartModelSync(ctx context.Context, params operation.StartModelSyncParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceModelSync); err != nil {
		return m.SendError(ctx, err)
	}
	policy, err := m.ctl.GetPolicy(ctx, params.ID)
	if err != nil {
		return m.SendError(ctx, err)
	}
	executionID, err := m.ctl.Start(ctx, policy, task.ExecutionTriggerManual)
	if err != nil {
		return m.SendError(ctx, err)
	}
	location := strings.TrimSuffix(params.HTTPRequest.URL.Path, "/") + "/" + strconv.FormatInt(executionID, 10)
	return operation.NewStartModelSyncCreated().WithLocation(location)
}

func (m *modelSyncAPI) StopModelSync(ctx context.Context, params operation.StopModelSyncParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionStop, rbac.ResourceModelSync); err != nil {
		return m.SendError(ctx, err)
	}
	if err := m.ctl.Stop(ctx, params.ExecutionID); err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewStopModelSyncOK()
}

func (m *modelSyncAPI) ListModelSyncExecutions(ctx context.Context, params operation.ListModelSyncExecutionsParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceModelSync); err != nil {
		return m.SendError(ctx, err)
	}
	if _, err := m.ctl.GetPolicy(ctx, params.ID); err != nil {
		return m.SendError(ctx, err)
	}
	query, err := m.BuildQuery(ctx, nil, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return m.SendError(ctx, err)
	}
	query.Keywords["PolicyID"] = params.ID
	if params.Status != nil {
		query.Keywords["Status"] = parseExecutionStatus(*params.Status)
	}
	if params.Trigger != nil {
		query.Keywords["Trigger"] = parseTrigger(*params.Trigger)
	}
	total, err := m.ctl.ExecutionCount(ctx, query)
	if err != nil {
		return m.SendError(ctx, err)
	}
	executions, err := m.ctl.ListExecutions(ctx, query)
	if err != nil {
		return m.SendError(ctx, err)
	}
	result := make([]*models.ModelSyncExecution, 0, len(executions))
	for _, e := range executions {
		result = append(result, convertModelSyncExecution(e))
	}
	return operation.NewListModelSyncExecutionsOK().
		WithXTotalCount(total).
		WithLink(m.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(result)
}

func (m *modelSyncAPI) GetModelSyncExecution(ctx context.Context, params operation.GetModelSyncExecutionParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceModelSync); err != nil {
		return m.SendError(ctx, err)
	}
	execution, err := m.ctl.GetExecution(ctx, params.ExecutionID)
	if err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewGetModelSyncExecutionOK().WithPayload(convertModelSyncExecution(execution))
}

func (m *modelSyncAPI) ListModelSyncTasks(ctx context.Context, params operation.ListModelSyncTasksParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceModelSync); err != nil {
		return m.SendError(ctx, err)
	}
	if _, err := m.ctl.GetExecution(ctx, params.ExecutionID); err != nil {
		return m.SendError(ctx, err)
	}
	query, err := m.BuildQuery(ctx, nil, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return m.SendError(ctx, err)
	}
	query.Keywords["ExecutionID"] = params.ExecutionID
	if params.Status != nil {
		query.Keywords["Status"] = parseTaskStatus(*params.Status)
	}
	total, err := m.ctl.TaskCount(ctx, query)
	if err != nil {
		return m.SendError(ctx, err)
	}
	tasks, err := m.ctl.ListTasks(ctx, query)
	if err != nil {
		return m.SendError(ctx, err)
	}
	result := make([]*models.ModelSyncTask, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, convertModelSyncTask(t))
	}
	return operation.NewListModelSyncTasksOK().
		WithXTotalCount(total).
		WithLink(m.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(result)
}

func (m *modelSyncAPI) GetModelSyncLog(ctx context.Context, params operation.GetModelSyncLogParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceModelSync); err != nil {
		return m.SendError(ctx, err)
	}
	execution, err := m.ctl.GetExecution(ctx, params.ExecutionID)
	if err != nil {
		return m.SendError(ctx, err)
	}
	t, err := m.ctl.GetTask(ctx, params.TaskID)
	if err != nil {
		return m.SendError(ctx, err)
	}
	if execution.ID != t.ExecutionID {
		return m.SendError(ctx, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessagef("execution %d contains no task with ID %d", params.ExecutionID, params.TaskID))
	}
	log, err := m.ctl.GetTaskLog(ctx, params.TaskID)
	if err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewGetModelSyncLogOK().WithContentType("text/plain").WithPayload(string(log))
}

func (m *modelSyncAPI) PreviewModelSync(ctx context.Context, params operation.PreviewModelSyncParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceModelSyncPolicy); err != nil {
		return m.SendError(ctx, err)
	}
	req := &modelsync.PreviewRequest{
		SrcRevision: params.Preview.SrcRevision,
		FileFilter:  toFileFilter(params.Preview.FileFilters, params.Preview.FileExcludes),
	}
	if params.Preview.RegistryID != nil {
		req.RegistryID = *params.Preview.RegistryID
	}
	if params.Preview.SrcRepository != nil {
		req.SrcRepository = *params.Preview.SrcRepository
	}
	preview, err := m.ctl.Preview(ctx, req)
	if err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewPreviewModelSyncOK().WithPayload(convertModelSyncPreview(preview))
}

func (m *modelSyncAPI) ListModelSyncAdapters(ctx context.Context, _ operation.ListModelSyncAdaptersParams) middleware.Responder {
	if err := m.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceModelSyncPolicy); err != nil {
		return m.SendError(ctx, err)
	}
	return operation.NewListModelSyncAdaptersOK().WithPayload(m.ctl.ListAdapterTypes(ctx))
}

func parseExecutionStatus(status string) string {
	switch status {
	case "InProgress":
		return job.RunningStatus.String()
	case "Succeed":
		return job.SuccessStatus.String()
	case "Stopped":
		return job.StoppedStatus.String()
	case "Failed":
		return job.ErrorStatus.String()
	}
	return status
}

func parseTaskStatus(status string) any {
	switch status {
	case "InProgress":
		return &q.OrList{Values: []any{job.ScheduledStatus.String(), job.RunningStatus.String()}}
	case "Succeed":
		return job.SuccessStatus.String()
	case "Failed":
		return job.ErrorStatus.String()
	}
	return status
}

func parseTrigger(trigger string) string {
	switch trigger {
	case "manual":
		return task.ExecutionTriggerManual
	case "scheduled":
		return task.ExecutionTriggerSchedule
	}
	return trigger
}

func convertModelSyncPolicy(p *modelsync.Policy) *models.ModelSyncPolicy {
	out := &models.ModelSyncPolicy{
		ID:                 p.ID,
		Name:               p.Name,
		Description:        p.Description,
		Creator:            p.Creator,
		Enabled:            p.Enabled,
		SrcRepository:      p.SrcRepository,
		SrcRevision:        p.SrcRevision,
		FileFilters:        []string{},
		FileExcludes:       []string{},
		DestProjectID:      p.DestProjectID,
		DestProjectName:    p.DestProjectName,
		DestRepository:     p.EffectiveDestRepository(),
		LastSyncedRevision: p.LastSyncedRevision,
		CreationTime:       strfmt.DateTime(p.CreationTime),
		UpdateTime:         strfmt.DateTime(p.UpdateTime),
		Trigger:            &models.ModelSyncTrigger{Type: p.TriggerType},
	}
	if p.FileFilter != nil {
		out.FileFilters = append(out.FileFilters, p.FileFilter.Include...)
		out.FileExcludes = append(out.FileExcludes, p.FileFilter.Exclude...)
	}
	if p.Cron != "" {
		out.Trigger.TriggerSettings = &models.ModelSyncTriggerTriggerSettings{Cron: p.Cron}
	}
	if p.Registry != nil {
		out.RegistryID = p.Registry.ID
		if p.Registry.Name != "" || p.Registry.Type != "" {
			out.Registry = convertRegistry(p.Registry)
		}
	}
	return out
}

func convertModelSyncExecution(e *modelsync.Execution) *models.ModelSyncExecution {
	out := &models.ModelSyncExecution{
		ID:         e.ID,
		PolicyID:   e.PolicyID,
		StatusText: e.StatusMessage,
		Operator:   e.Operator,
		StartTime:  strfmt.DateTime(e.StartTime),
		EndTime:    strfmt.DateTime(e.EndTime),
	}
	if e.Metrics != nil {
		out.Total = e.Metrics.TaskCount
		out.Succeed = e.Metrics.SuccessTaskCount
		out.Failed = e.Metrics.ErrorTaskCount
		out.InProgress = e.Metrics.PendingTaskCount + e.Metrics.ScheduledTaskCount + e.Metrics.RunningTaskCount
		out.Stopped = e.Metrics.StoppedTaskCount
	}
	switch e.Trigger {
	case task.ExecutionTriggerManual:
		out.Trigger = "manual"
	case task.ExecutionTriggerSchedule:
		out.Trigger = "scheduled"
	default:
		out.Trigger = e.Trigger
	}
	switch e.Status {
	case job.RunningStatus.String():
		out.Status = "InProgress"
	case job.SuccessStatus.String():
		out.Status = "Succeed"
	case job.StoppedStatus.String():
		out.Status = "Stopped"
	case job.ErrorStatus.String():
		out.Status = "Failed"
	default:
		out.Status = e.Status
	}
	return out
}

func convertModelSyncTask(t *modelsync.Task) *models.ModelSyncTask {
	out := &models.ModelSyncTask{
		ID:             t.ID,
		ExecutionID:    t.ExecutionID,
		StatusMessage:  t.StatusMessage,
		JobID:          t.JobID,
		SrcRepository:  t.SrcRepository,
		SrcRevision:    t.SrcRevision,
		DestRepository: t.DestRepository,
		Revision:       t.Revision,
		Digest:         t.Digest,
		Tags:           t.Tags,
		Files:          t.Files,
		Size:           t.Size,
		Skipped:        t.Skipped,
		CreationTime:   strfmt.DateTime(t.CreationTime),
		StartTime:      strfmt.DateTime(t.StartTime),
		EndTime:        strfmt.DateTime(t.EndTime),
	}
	if out.Tags == nil {
		out.Tags = []string{}
	}
	switch t.Status {
	case job.PendingStatus.String():
		out.Status = "Pending"
	case job.ScheduledStatus.String(), job.RunningStatus.String():
		out.Status = "InProgress"
	case job.SuccessStatus.String():
		out.Status = "Succeed"
	case job.StoppedStatus.String():
		out.Status = "Stopped"
	case job.ErrorStatus.String():
		out.Status = "Failed"
	default:
		out.Status = t.Status
	}
	return out
}

func convertModelSyncPreview(p *modelsync.Preview) *models.ModelSyncPreview {
	out := &models.ModelSyncPreview{
		Repository:   p.Repository,
		Ref:          p.Ref,
		Revision:     p.Revision,
		SourceURL:    p.SourceURL,
		RevisionTag:  p.RevisionTag,
		Digest:       p.Digest,
		Metadata:     p.Metadata,
		Tags:         p.Tags,
		TotalFiles:   p.TotalFiles,
		TotalSize:    p.TotalSize,
		MatchedFiles: p.MatchedFiles,
		MatchedSize:  p.MatchedSize,
		Files:        make([]*models.ModelSyncPreviewFile, 0, len(p.Files)),
	}
	if out.Tags == nil {
		out.Tags = []string{}
	}
	for _, f := range p.Files {
		out.Files = append(out.Files, &models.ModelSyncPreviewFile{
			Path:    f.Path,
			Size:    f.Size,
			Sha256:  f.SHA256,
			Type:    f.Type,
			Matched: f.Matched,
		})
	}
	return out
}

// toFileFilter builds the file filter from the include and exclude pattern lists, nil when both are empty.
func toFileFilter(include, exclude []string) *model.FileFilter {
	clean := func(in []string) []string {
		var out []string
		for _, p := range in {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	f := &model.FileFilter{Include: clean(include), Exclude: clean(exclude)}
	if f.IsEmpty() {
		return nil
	}
	return f
}
