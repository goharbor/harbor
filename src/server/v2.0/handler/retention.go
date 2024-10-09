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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-openapi/runtime/middleware"

	"github.com/goharbor/harbor/src/common/rbac"
	projectCtl "github.com/goharbor/harbor/src/controller/project"
	retentionCtl "github.com/goharbor/harbor/src/controller/retention"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/retention"
)

func newRetentionAPI() *retentionAPI {
	return &retentionAPI{
		projectCtl:   projectCtl.Ctl,
		retentionCtl: retentionCtl.Ctl,
		proMetaMgr:   pkg.ProjectMetaMgr,
	}
}

type retentionAPI struct {
	BaseAPI
	proMetaMgr   metadata.Manager
	retentionCtl retentionCtl.Controller
	projectCtl   projectCtl.Controller
}

var (
	rentenitionMetadataPayload = &models.RetentionMetadata{
		Templates: []*models.RetentionRuleMetadata{
			{
				Action:       "retain",
				DisplayText:  "the most recently pushed # artifacts",
				RuleTemplate: "latestPushedK",
				Params: []*models.RetentionRuleParamMetadata{
					{
						Required: true,
						Type:     "int",
						Unit:     "COUNT",
					},
				},
			},
			{
				RuleTemplate: "latestPulledN",
				DisplayText:  "the most recently pulled # artifacts",
				Action:       "retain",
				Params: []*models.RetentionRuleParamMetadata{
					{
						Type:     "int",
						Unit:     "COUNT",
						Required: true,
					},
				},
			},
			{
				RuleTemplate: "nDaysSinceLastPush",
				DisplayText:  "pushed within the last # days",
				Action:       "retain",
				Params: []*models.RetentionRuleParamMetadata{
					{
						Type:     "int",
						Unit:     "DAYS",
						Required: true,
					},
				},
			},
			{
				RuleTemplate: "nDaysSinceLastPull",
				DisplayText:  "pulled within the last # days",
				Action:       "retain",
				Params: []*models.RetentionRuleParamMetadata{
					{
						Type:     "int",
						Unit:     "DAYS",
						Required: true,
					},
				},
			},
			{
				RuleTemplate: "always",
				DisplayText:  "always",
				Action:       "retain",
				Params:       []*models.RetentionRuleParamMetadata{},
			},
		},
		ScopeSelectors: []*models.RetentionSelectorMetadata{
			{
				DisplayText: "Repositories",
				Kind:        "doublestar",
				Decorations: []string{
					"repoMatches",
					"repoExcludes",
				},
			},
		},
		TagSelectors: []*models.RetentionSelectorMetadata{
			{
				DisplayText: "Tags",
				Kind:        "doublestar",
				Decorations: []string{
					"matches",
					"excludes",
				},
			},
		},
	}
)

func (r *retentionAPI) Prepare(ctx context.Context, _ string, _ interface{}) middleware.Responder {
	if err := r.RequireAuthenticated(ctx); err != nil {
		return r.SendError(ctx, err)
	}

	return nil
}

func (r *retentionAPI) GetRentenitionMetadata(_ context.Context, _ operation.GetRentenitionMetadataParams) middleware.Responder {
	return operation.NewGetRentenitionMetadataOK().WithPayload(rentenitionMetadataPayload)
}

func (r *retentionAPI) GetRetention(ctx context.Context, params operation.GetRetentionParams) middleware.Responder {
	id := params.ID
	p, err := r.retentionCtl.GetRetention(ctx, id)
	if err != nil {
		return r.SendError(ctx, err)
	}
	err = r.requireAccess(ctx, p, rbac.ActionRead)
	if err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewGetRetentionOK().WithPayload(model.NewRetentionPolicy(p).ToSwagger())
}

func (r *retentionAPI) CreateRetention(ctx context.Context, params operation.CreateRetentionParams) middleware.Responder {
	p := model.NewRetentionPolicyFromSwagger(params.Policy).Metadata
	if len(p.Rules) > 15 {
		return r.SendError(ctx, errors.BadRequestError(fmt.Errorf("only 15 rules are allowed at most")))
	}
	if err := r.checkRuleConflict(p); err != nil {
		return r.SendError(ctx, errors.ConflictError(err))
	}
	err := r.requireAccess(ctx, p, rbac.ActionCreate)
	if err != nil {
		return r.SendError(ctx, err)
	}

	switch p.Scope.Level {
	case policy.ScopeLevelProject:
		if p.Scope.Reference <= 0 {
			return r.SendError(ctx, errors.BadRequestError(fmt.Errorf("invalid Project id %d", p.Scope.Reference)))
		}

		if _, err := r.projectCtl.Get(ctx, p.Scope.Reference); err != nil {
			if errors.IsNotFoundErr(err) {
				return r.SendError(ctx, errors.BadRequestError(fmt.Errorf("invalid Project id %d", p.Scope.Reference)))
			}
			return r.SendError(ctx, errors.BadRequestError(err))
		}
	default:
		return r.SendError(ctx, errors.BadRequestError(fmt.Errorf("scope %s is not support", p.Scope.Level)))
	}

	old, err := r.proMetaMgr.Get(ctx, p.Scope.Reference, "retention_id")
	if err != nil {
		return r.SendError(ctx, err)
	}
	if len(old) > 0 {
		return r.SendError(ctx, errors.BadRequestError(fmt.Errorf("project %v already has retention policy %v", p.Scope.Reference, old["retention_id"])))
	}

	id, err := r.retentionCtl.CreateRetention(ctx, p)
	if err != nil {
		return r.SendError(ctx, err)
	}

	if err := r.proMetaMgr.Add(ctx, p.Scope.Reference,
		map[string]string{"retention_id": strconv.FormatInt(id, 10)}); err != nil {
		return r.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), id)
	return operation.NewCreateRetentionCreated().WithLocation(location)
}

func (r *retentionAPI) UpdateRetention(ctx context.Context, params operation.UpdateRetentionParams) middleware.Responder {
	p := model.NewRetentionPolicyFromSwagger(params.Policy).Metadata
	p.ID = params.ID
	if len(p.Rules) > 15 {
		return r.SendError(ctx, errors.BadRequestError(fmt.Errorf("only 15 rules are allowed at most")))
	}
	if err := r.checkRuleConflict(p); err != nil {
		return r.SendError(ctx, errors.ConflictError(err))
	}

	if err := r.requireAccess(ctx, p, rbac.ActionUpdate); err != nil {
		return r.SendError(ctx, err)
	}

	if err := r.requirePolicyAccess(ctx, p); err != nil {
		return r.SendError(ctx, err)
	}

	if err := r.retentionCtl.UpdateRetention(ctx, p); err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewUpdateRetentionOK()
}

func (r *retentionAPI) checkRuleConflict(p *policy.Metadata) error {
	temp := make(map[string]int)
	for n, rule := range p.Rules {
		rule.ID = 0
		bs, _ := json.Marshal(rule)
		if old, exists := temp[string(bs)]; exists {
			return fmt.Errorf("rule %d is conflict with rule %d", n, old)
		}
		temp[string(bs)] = n
		rule.ID = n
	}
	return nil
}

func (r *retentionAPI) DeleteRetention(ctx context.Context, params operation.DeleteRetentionParams) middleware.Responder {
	p, err := r.retentionCtl.GetRetention(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, errors.BadRequestError(err))
	}
	err = r.requireAccess(ctx, p, rbac.ActionDelete)
	if err != nil {
		return r.SendError(ctx, err)
	}

	if err = r.retentionCtl.DeleteRetention(ctx, params.ID); err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewDeleteRetentionOK()
}

func (r *retentionAPI) TriggerRetentionExecution(ctx context.Context, params operation.TriggerRetentionExecutionParams) middleware.Responder {
	p, err := r.retentionCtl.GetRetention(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, errors.BadRequestError(err))
	}
	err = r.requireAccess(ctx, p, rbac.ActionUpdate)
	if err != nil {
		return r.SendError(ctx, err)
	}

	eid, err := r.retentionCtl.TriggerRetentionExec(ctx, params.ID, task.ExecutionTriggerManual, params.Body.DryRun)
	if err != nil {
		return r.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), eid)
	return operation.NewTriggerRetentionExecutionCreated().WithLocation(location)
}

func (r *retentionAPI) OperateRetentionExecution(ctx context.Context, params operation.OperateRetentionExecutionParams) middleware.Responder {
	if params.Body.Action != "stop" {
		return r.SendError(ctx, errors.BadRequestError((fmt.Errorf("action should be 'stop'"))))
	}
	p, err := r.retentionCtl.GetRetention(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, errors.BadRequestError(err))
	}
	if p == nil {
		return r.SendError(ctx, errors.New("retention policy is not found").WithCode(errors.NotFoundCode))
	}
	if err := r.requireAccess(ctx, p, rbac.ActionUpdate); err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.requirePolicyAccess(ctx, p); err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.requireExecutionInProject(ctx, p, params.Eid); err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.retentionCtl.OperateRetentionExec(ctx, params.Eid, params.Body.Action); err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewOperateRetentionExecutionOK()
}

func (r *retentionAPI) ListRetentionExecutions(ctx context.Context, params operation.ListRetentionExecutionsParams) middleware.Responder {
	query, err := r.BuildQuery(ctx, nil, nil, params.Page, params.PageSize)
	if err != nil {
		return r.SendError(ctx, err)
	}
	p, err := r.retentionCtl.GetRetention(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, errors.BadRequestError(err))
	}
	err = r.requireAccess(ctx, p, rbac.ActionList)
	if err != nil {
		return r.SendError(ctx, err)
	}
	execs, err := r.retentionCtl.ListRetentionExecs(ctx, params.ID, query)
	if err != nil {
		return r.SendError(ctx, err)
	}
	total, err := r.retentionCtl.GetTotalOfRetentionExecs(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, err)
	}
	var payload []*models.RetentionExecution
	for _, e := range execs {
		payload = append(payload, model.NewRetentionExec(e).ToSwagger())
	}
	return operation.NewListRetentionExecutionsOK().WithXTotalCount(total).
		WithLink(r.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(payload)
}

func (r *retentionAPI) ListRetentionTasks(ctx context.Context, params operation.ListRetentionTasksParams) middleware.Responder {
	query, err := r.BuildQuery(ctx, nil, nil, params.Page, params.PageSize)
	if err != nil {
		return r.SendError(ctx, err)
	}
	p, err := r.retentionCtl.GetRetention(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, errors.BadRequestError(err))
	}
	if p == nil {
		return r.SendError(ctx, errors.New("retention policy is not found").WithCode(errors.NotFoundCode))
	}
	if err := r.requireAccess(ctx, p, rbac.ActionList); err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.requirePolicyAccess(ctx, p); err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.requireExecutionInProject(ctx, p, params.Eid); err != nil {
		return r.SendError(ctx, err)
	}
	tasks, err := r.retentionCtl.ListRetentionExecTasks(ctx, params.Eid, query)
	if err != nil {
		return r.SendError(ctx, err)
	}
	total, err := r.retentionCtl.GetTotalOfRetentionExecTasks(ctx, params.Eid)
	if err != nil {
		return r.SendError(ctx, err)
	}
	var payload []*models.RetentionExecutionTask
	for _, t := range tasks {
		payload = append(payload, model.NewRetentionTask(t).ToSwagger())
	}
	return operation.NewListRetentionTasksOK().WithXTotalCount(total).
		WithLink(r.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(payload)
}

func (r *retentionAPI) GetRetentionTaskLog(ctx context.Context, params operation.GetRetentionTaskLogParams) middleware.Responder {
	p, err := r.retentionCtl.GetRetention(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, errors.BadRequestError(err))
	}
	if p == nil {
		return r.SendError(ctx, errors.New("retention policy is not found").WithCode(errors.NotFoundCode))
	}
	if err := r.requireAccess(ctx, p, rbac.ActionRead); err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.requirePolicyAccess(ctx, p); err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.requireTaskInProject(ctx, p, params.Eid, params.Tid); err != nil {
		return r.SendError(ctx, err)
	}

	log, err := r.retentionCtl.GetRetentionExecTaskLog(ctx, params.Tid)
	if err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewGetRetentionTaskLogOK().WithPayload(string(log))
}

func (r *retentionAPI) requireAccess(ctx context.Context, p *policy.Metadata, action rbac.Action, subresources ...rbac.Resource) error {
	switch p.Scope.Level {
	case "project":
		if len(subresources) == 0 {
			subresources = append(subresources, rbac.ResourceTagRetention)
		}
		err := r.RequireProjectAccess(ctx, p.Scope.Reference, action, subresources...)
		return err
	}
	return r.RequireSystemAccess(ctx, action, rbac.ResourceTagRetention)
}

// requirePolicyAccess checks the scope reference whether has the permission to
// the retention policy.
func (r *retentionAPI) requirePolicyAccess(ctx context.Context, p *policy.Metadata) error {
	// the id of policy should be consistent with project metadata
	meta, err := r.proMetaMgr.Get(ctx, p.Scope.Reference, "retention_id")
	if err != nil {
		return err
	}
	// validate
	if len(meta["retention_id"]) > 0 {
		// return err if retention id does not match
		if meta["retention_id"] == fmt.Sprintf("%d", p.ID) {
			return nil
		}
	}

	return errors.NotFoundError(errors.Errorf("the retention policy id %d does not match", p.ID))
}

func (r *retentionAPI) requireExecutionInProject(ctx context.Context, p *policy.Metadata, executionID int64) error {
	exec, err := r.retentionCtl.GetRetentionExec(ctx, executionID)
	if err != nil {
		return err
	}
	if exec == nil {
		return errors.New(nil).WithMessagef("project: %d, execution id %d not found", p.Scope.Reference, executionID).WithCode(errors.NotFoundCode)
	}
	if exec.PolicyID != p.ID {
		return errors.New(nil).WithMessagef("project: %d, execution id %d not found", p.Scope.Reference, executionID).WithCode(errors.NotFoundCode)
	}
	if exec.Type != job.RetentionVendorType {
		return errors.New(nil).WithMessagef("project: %d, execution id %d not found", p.Scope.Reference, executionID).WithCode(errors.NotFoundCode)
	}
	return nil
}

func (r *retentionAPI) requireTaskInProject(ctx context.Context, p *policy.Metadata, executionID, taskID int64) error {
	if err := r.requireExecutionInProject(ctx, p, executionID); err != nil {
		return err
	}
	task, err := r.retentionCtl.GetRetentionExecTask(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New(nil).WithMessagef("project: %d, execution id %d not found", p.Scope.Reference, executionID).WithCode(errors.NotFoundCode)
	}
	if task.ExecutionID != executionID {
		return errors.New(nil).WithMessagef("project: %d, execution id %d not found", p.Scope.Reference, executionID).WithCode(errors.NotFoundCode)
	}
	return nil
}
