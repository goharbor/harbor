package handler

import (
	"context"

	"github.com/go-openapi/runtime/middleware"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/notification/job"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	policyModel "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/webhookjob"
)

func newNotificationJobAPI() *notificationJobAPI {
	return &notificationJobAPI{
		webhookjobMgr:    job.Mgr,
		webhookPolicyMgr: policy.Mgr,
		projectMgr:       pkg.ProjectMgr,
	}
}

type notificationJobAPI struct {
	BaseAPI
	webhookjobMgr    job.Manager
	webhookPolicyMgr policy.Manager
	projectMgr       project.Manager
}

func (n *notificationJobAPI) ListWebhookJobs(ctx context.Context, params webhookjob.ListWebhookJobsParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionList, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	policy, err := n.webhookPolicyMgr.Get(ctx, params.PolicyID)
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
	query.Keywords["PolicyID"] = policy.ID
	if len(params.Status) != 0 {
		query.Keywords["Status"] = params.Status
	}

	total, err := n.webhookjobMgr.Count(ctx, query)
	if err != nil {
		return n.SendError(ctx, err)
	}

	jobs, err := n.webhookjobMgr.List(ctx, query)
	if err != nil {
		return n.SendError(ctx, err)
	}

	var results []*models.WebhookJob
	for _, j := range jobs {
		results = append(results, model.NewNotificationJob(j).ToSwagger())
	}

	return webhookjob.NewListWebhookJobsOK().
		WithXTotalCount(total).
		WithLink(n.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

// requirePolicyAccess checks whether the project has the permission to the policy.
func (n *notificationJobAPI) requirePolicyAccess(ctx context.Context, projectNameIrID interface{}, policy *policyModel.Policy) error {
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
