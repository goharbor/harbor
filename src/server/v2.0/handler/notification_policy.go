package handler

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notification/job"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	policy_model "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/webhook"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/webhook"
	"strings"
	"time"
)

func newNotificationPolicyAPI() *notificationPolicyAPI {
	return &notificationPolicyAPI{
		webhookjobMgr:    job.Mgr,
		webhookPolicyMgr: policy.Mgr,
	}
}

type notificationPolicyAPI struct {
	BaseAPI
	webhookjobMgr    job.Manager
	webhookPolicyMgr policy.Manager
}

func (n *notificationPolicyAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	return nil
}

func (n *notificationPolicyAPI) ListWebhookPoliciesOfProject(ctx context.Context, params webhook.ListWebhookPoliciesOfProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionList, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	projectID, err := getProjectID(ctx, projectNameOrID)
	if err != nil {
		return n.SendError(ctx, err)
	}

	query, err := n.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return n.SendError(ctx, err)
	}
	query.Keywords["ProjectID"] = projectID

	total, err := n.webhookPolicyMgr.Count(ctx, query)
	if err != nil {
		return n.SendError(ctx, err)
	}

	policies, err := n.webhookPolicyMgr.List(ctx, query)
	if err != nil {
		return n.SendError(ctx, err)
	}
	var results []*models.WebhookPolicy
	for _, p := range policies {
		results = append(results, model.NewNotifiactionPolicy(p).ToSwagger())
	}

	return operation.NewListWebhookPoliciesOfProjectOK().
		WithXTotalCount(total).
		WithLink(n.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (n *notificationPolicyAPI) CreateWebhookPolicyOfProject(ctx context.Context, params webhook.CreateWebhookPolicyOfProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionCreate, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	policy := &policy_model.Policy{}
	lib.JSONCopy(policy, params.Policy)

	if ok, err := n.validateEventTypes(policy); !ok {
		return n.SendError(ctx, err)
	}
	if ok, err := n.validateTargets(policy); !ok {
		return n.SendError(ctx, err)
	}

	projectID, err := getProjectID(ctx, projectNameOrID)
	if err != nil {
		return n.SendError(ctx, err)
	}
	policy.ProjectID = projectID
	id, err := n.webhookPolicyMgr.Create(ctx, policy)
	if err != nil {
		return n.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), id)
	return operation.NewCreateWebhookPolicyOfProjectCreated().WithLocation(location)
}

func (n *notificationPolicyAPI) UpdateWebhookPolicyOfProject(ctx context.Context, params webhook.UpdateWebhookPolicyOfProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionUpdate, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	policy := &policy_model.Policy{}
	lib.JSONCopy(policy, params.Policy)

	if ok, err := n.validateEventTypes(policy); !ok {
		return n.SendError(ctx, err)
	}
	if ok, err := n.validateTargets(policy); !ok {
		return n.SendError(ctx, err)
	}

	projectID, err := getProjectID(ctx, projectNameOrID)
	if err != nil {
		return n.SendError(ctx, err)
	}
	policy.ProjectID = projectID
	if err := n.webhookPolicyMgr.Update(ctx, policy); err != nil {
		return n.SendError(ctx, err)
	}

	return operation.NewUpdateWebhookPolicyOfProjectOK()
}

func (n *notificationPolicyAPI) DeleteWebhookPolicyOfProject(ctx context.Context, params webhook.DeleteWebhookPolicyOfProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionDelete, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	if err := n.webhookPolicyMgr.Delete(ctx, params.WebhookPolicyID); err != nil {
		return n.SendError(ctx, err)
	}
	return operation.NewDeleteWebhookPolicyOfProjectOK()
}

func (n *notificationPolicyAPI) GetWebhookPolicyOfProject(ctx context.Context, params webhook.GetWebhookPolicyOfProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	policy, err := n.webhookPolicyMgr.Get(ctx, params.WebhookPolicyID)
	if err != nil {
		return n.SendError(ctx, err)
	}

	return operation.NewGetWebhookPolicyOfProjectOK().WithPayload(model.NewNotifiactionPolicy(policy).ToSwagger())
}

func (n *notificationPolicyAPI) LastTrigger(ctx context.Context, params webhook.LastTriggerParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	projectID, err := getProjectID(ctx, projectNameOrID)
	if err != nil {
		return n.SendError(ctx, err)
	}

	query := &q.Query{
		Keywords: q.KeyWords{
			"ProjectID": projectID,
		},
	}
	policies, err := n.webhookPolicyMgr.List(ctx, query)
	if err != nil {
		return n.SendError(ctx, err)
	}
	triggers, err := n.constructPolicyWithTriggerTime(ctx, policies)
	if err != nil {
		return n.SendError(ctx, err)
	}

	return operation.NewLastTriggerOK().WithPayload(triggers)
}

func (n *notificationPolicyAPI) GetSupportedEventTypes(ctx context.Context, params webhook.GetSupportedEventTypesParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	var notificationTypes = &models.SupportedWebhookEventTypes{}
	for key := range notification.SupportedNotifyTypes {
		notificationTypes.NotifyType = append(notificationTypes.NotifyType, models.NotifyType(key))
	}

	for key := range notification.SupportedEventTypes {
		notificationTypes.EventType = append(notificationTypes.EventType, models.EventType(key))
	}

	return operation.NewGetSupportedEventTypesOK().WithPayload(notificationTypes)
}

func (n *notificationPolicyAPI) getLastTriggerTimeGroupByEventType(ctx context.Context, eventType string, policyID int64) (time.Time, error) {
	jobs, err := n.webhookjobMgr.ListJobsGroupByEventType(ctx, policyID)
	if err != nil {
		return time.Time{}, err
	}

	for _, job := range jobs {
		if eventType == job.EventType {
			return job.CreationTime, nil
		}
	}
	return time.Time{}, nil
}

func (n *notificationPolicyAPI) validateTargets(policy *policy_model.Policy) (bool, error) {
	if len(policy.Targets) == 0 {
		return false, errors.New(nil).WithMessage("empty notification target with policy %s", policy.Name).WithCode(errors.BadRequestCode)
	}
	for _, target := range policy.Targets {
		url, err := utils.ParseEndpoint(target.Address)
		if err != nil {
			return false, errors.New(err).WithCode(errors.BadRequestCode)
		}
		// Prevent SSRF security issue #3755
		target.Address = url.Scheme + "://" + url.Host + url.Path

		_, ok := notification.SupportedNotifyTypes[target.Type]
		if !ok {
			return false, errors.New(nil).WithMessage("unsupported target type %s with policy %s", target.Type, policy.Name).WithCode(errors.BadRequestCode)
		}
	}
	return true, nil
}

func (n *notificationPolicyAPI) validateEventTypes(policy *policy_model.Policy) (bool, error) {
	if len(policy.EventTypes) == 0 {
		return false, errors.New(nil).WithMessage("empty event type").WithCode(errors.BadRequestCode)
	}
	for _, eventType := range policy.EventTypes {
		_, ok := notification.SupportedEventTypes[eventType]
		if !ok {
			return false, errors.New(nil).WithMessage("unsupported event type %s", eventType).WithCode(errors.BadRequestCode)
		}
	}
	return true, nil
}

// constructPolicyWithTriggerTime construct notification policy information displayed in UI
// including event type, enabled, creation time, last trigger time
func (n *notificationPolicyAPI) constructPolicyWithTriggerTime(ctx context.Context, policies []*policy_model.Policy) ([]*models.WebhookLastTrigger, error) {
	res := []*models.WebhookLastTrigger{}
	if policies != nil {
		for _, policy := range policies {
			for _, t := range policy.EventTypes {
				ply := &models.WebhookLastTrigger{
					PolicyName:   policy.Name,
					EventType:    t,
					Enabled:      policy.Enabled,
					CreationTime: strfmt.DateTime(policy.CreationTime),
				}
				if !policy.CreationTime.IsZero() {
					ply.CreationTime = strfmt.DateTime(policy.CreationTime)
				}

				ltTime, err := n.getLastTriggerTimeGroupByEventType(ctx, t, policy.ID)
				if err != nil {
					return nil, err
				}
				if !ltTime.IsZero() {
					ply.LastTriggerTime = strfmt.DateTime(ltTime)
				}
				res = append(res, ply)
			}
		}
	}
	return res, nil
}
