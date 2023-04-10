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

	"github.com/go-openapi/runtime/middleware"

	"github.com/goharbor/harbor/src/common/rbac"
	webhook_ctl "github.com/goharbor/harbor/src/controller/webhook"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg"
	policyModel "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/webhookjob"
)

func newWebhookJobAPI() *webhookJobAPI {
	return &webhookJobAPI{
		webhookCtl: webhook_ctl.Ctl,
		projectMgr: pkg.ProjectMgr,
	}
}

type webhookJobAPI struct {
	BaseAPI
	webhookCtl webhook_ctl.Controller
	projectMgr project.Manager
}

func (n *webhookJobAPI) ListWebhookJobs(ctx context.Context, params webhookjob.ListWebhookJobsParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionList, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	policy, err := n.webhookCtl.GetPolicy(ctx, params.PolicyID)
	if err != nil {
		return n.SendError(ctx, err)
	}

	if err := n.requirePolicyAccess(ctx, projectNameOrID, policy); err != nil {
		return n.SendError(ctx, err)
	}

	query, err := n.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return n.SendError(ctx, err)
	}

	if len(params.Status) != 0 {
		query.Keywords["status"] = params.Status
	}

	total, err := n.webhookCtl.CountExecutions(ctx, params.PolicyID, query)
	if err != nil {
		return n.SendError(ctx, err)
	}
	// the relationship of webhook execution and task is 1:1, so we can think the execution is the job as before.
	jobs, err := n.webhookCtl.ListExecutions(ctx, params.PolicyID, query)
	if err != nil {
		return n.SendError(ctx, err)
	}

	var results []*models.WebhookJob
	for _, j := range jobs {
		results = append(results, model.NewWebhookJob(j).ToSwagger())
	}

	return webhookjob.NewListWebhookJobsOK().
		WithXTotalCount(total).
		WithLink(n.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

// requirePolicyAccess checks whether the project has the permission to the policy.
func (n *webhookJobAPI) requirePolicyAccess(ctx context.Context, projectNameIrID interface{}, policy *policyModel.Policy) error {
	p, err := n.projectMgr.Get(ctx, projectNameIrID)
	if err != nil {
		return err
	}
	// check the projectID whether match with the projectID in policy
	if p.ProjectID != policy.ProjectID {
		return errors.NotFoundError(errors.Errorf("project id %d does not match", p.ProjectID))
	}

	return nil
}
