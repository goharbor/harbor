package handler

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/notification/job"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/webhookjob"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/webhookjob"
)

func newNotificationJobAPI() *notificationJobAPI {
	return &notificationJobAPI{
		webhookjobMgr:    job.Mgr,
		webhookPolicyMgr: policy.Mgr,
	}
}

type notificationJobAPI struct {
	BaseAPI
	webhookjobMgr    job.Manager
	webhookPolicyMgr policy.Manager
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

	return operation.NewListWebhookJobsOK().
		WithXTotalCount(total).
		WithLink(n.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}
