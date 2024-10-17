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
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/task"
	webhook_ctl "github.com/goharbor/harbor/src/controller/webhook"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification"
	policy_model "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/webhook"
)

func newWebhookAPI() *webhookAPI {
	return &webhookAPI{
		execCtl:    task.ExecutionCtl,
		taskCtl:    task.Ctl,
		webhookCtl: webhook_ctl.Ctl,
	}
}

type webhookAPI struct {
	BaseAPI
	execCtl    task.ExecutionController
	taskCtl    task.Controller
	webhookCtl webhook_ctl.Controller
}

func (n *webhookAPI) Prepare(_ context.Context, _ string, _ interface{}) middleware.Responder {
	return nil
}

func (n *webhookAPI) requirePolicyInProject(ctx context.Context, projectIDOrName interface{}, policyID int64) error {
	projectID, err := getProjectID(ctx, projectIDOrName)
	if err != nil {
		return err
	}

	l, err := n.webhookCtl.GetPolicy(ctx, policyID)
	if err != nil {
		return err
	}

	if projectID != l.ProjectID {
		return errors.NotFoundError(fmt.Errorf("project id:%d, webhook policy id: %d not found", projectID, policyID))
	}
	return nil
}

func (n *webhookAPI) requireExecutionInPolicy(ctx context.Context, execID, policyID int64) error {
	exec, err := n.execCtl.Get(ctx, execID)
	if err != nil {
		return err
	}

	if exec.VendorID == policyID && (exec.VendorType == job.WebhookJobVendorType || exec.VendorType == job.SlackJobVendorType || exec.VendorType == job.TeamsJobVendorType) {
		return nil
	}

	return errors.NotFoundError(fmt.Errorf("execution %d not found in policy %d", execID, policyID))
}

func (n *webhookAPI) requireTaskInExecution(ctx context.Context, taskID, execID int64) error {
	task, err := n.taskCtl.Get(ctx, taskID)
	if err != nil {
		return err
	}

	if task.ExecutionID == execID {
		return nil
	}

	return errors.NotFoundError(fmt.Errorf("task %d not found in execution %d", taskID, execID))
}

func (n *webhookAPI) ListWebhookPoliciesOfProject(ctx context.Context, params webhook.ListWebhookPoliciesOfProjectParams) middleware.Responder {
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

	total, err := n.webhookCtl.CountPolicies(ctx, query)
	if err != nil {
		return n.SendError(ctx, err)
	}

	policies, err := n.webhookCtl.ListPolicies(ctx, query)
	if err != nil {
		return n.SendError(ctx, err)
	}
	var results []*models.WebhookPolicy
	for _, p := range policies {
		results = append(results, model.NewWebhookPolicy(p).ToSwagger())
	}

	return webhook.NewListWebhookPoliciesOfProjectOK().
		WithXTotalCount(total).
		WithLink(n.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (n *webhookAPI) CreateWebhookPolicyOfProject(ctx context.Context, params webhook.CreateWebhookPolicyOfProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionCreate, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	policy := &policy_model.Policy{}
	if err := lib.JSONCopy(policy, params.Policy); err != nil {
		log.Warningf("failed to call JSONCopy on notification policy when CreateWebhookPolicyOfProject, error: %v", err)
	}

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
	id, err := n.webhookCtl.CreatePolicy(ctx, policy)
	if err != nil {
		return n.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), id)
	return webhook.NewCreateWebhookPolicyOfProjectCreated().WithLocation(location)
}

func (n *webhookAPI) UpdateWebhookPolicyOfProject(ctx context.Context, params webhook.UpdateWebhookPolicyOfProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionUpdate, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	projectID, err := getProjectID(ctx, projectNameOrID)
	if err != nil {
		return n.SendError(ctx, err)
	}
	policyID := params.WebhookPolicyID
	if err := n.requirePolicyInProject(ctx, projectID, policyID); err != nil {
		return n.SendError(ctx, err)
	}
	policy := &policy_model.Policy{}
	if err := lib.JSONCopy(policy, params.Policy); err != nil {
		log.Warningf("failed to call JSONCopy on notification policy when UpdateWebhookPolicyOfProject, error: %v", err)
	}

	if ok, err := n.validateEventTypes(policy); !ok {
		return n.SendError(ctx, err)
	}
	if ok, err := n.validateTargets(policy); !ok {
		return n.SendError(ctx, err)
	}

	policy.ID = policyID
	policy.ProjectID = projectID
	if err := n.webhookCtl.UpdatePolicy(ctx, policy); err != nil {
		return n.SendError(ctx, err)
	}

	return webhook.NewUpdateWebhookPolicyOfProjectOK()
}

func (n *webhookAPI) DeleteWebhookPolicyOfProject(ctx context.Context, params webhook.DeleteWebhookPolicyOfProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionDelete, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.requirePolicyInProject(ctx, projectNameOrID, params.WebhookPolicyID); err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.webhookCtl.DeletePolicy(ctx, params.WebhookPolicyID); err != nil {
		return n.SendError(ctx, err)
	}
	return webhook.NewDeleteWebhookPolicyOfProjectOK()
}

func (n *webhookAPI) GetWebhookPolicyOfProject(ctx context.Context, params webhook.GetWebhookPolicyOfProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	projectID, err := getProjectID(ctx, projectNameOrID)
	if err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.RequireProjectAccess(ctx, projectID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.requirePolicyInProject(ctx, projectID, params.WebhookPolicyID); err != nil {
		return n.SendError(ctx, err)
	}

	policy, err := n.webhookCtl.GetPolicy(ctx, params.WebhookPolicyID)
	if err != nil {
		return n.SendError(ctx, err)
	}

	return webhook.NewGetWebhookPolicyOfProjectOK().WithPayload(model.NewWebhookPolicy(policy).ToSwagger())
}

func (n *webhookAPI) ListExecutionsOfWebhookPolicy(ctx context.Context, params webhook.ListExecutionsOfWebhookPolicyParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	projectID, err := getProjectID(ctx, projectNameOrID)
	if err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.RequireProjectAccess(ctx, projectID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.requirePolicyInProject(ctx, projectID, params.WebhookPolicyID); err != nil {
		return n.SendError(ctx, err)
	}

	query, err := n.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return n.SendError(ctx, err)
	}

	total, err := n.webhookCtl.CountExecutions(ctx, params.WebhookPolicyID, query)
	if err != nil {
		return n.SendError(ctx, err)
	}

	executions, err := n.webhookCtl.ListExecutions(ctx, params.WebhookPolicyID, query)
	if err != nil {
		return n.SendError(ctx, err)
	}

	var payloads []*models.Execution
	for _, exec := range executions {
		p, err := convertExecutionToPayload(exec)
		if err != nil {
			return n.SendError(ctx, err)
		}
		payloads = append(payloads, p)
	}

	return webhook.NewListExecutionsOfWebhookPolicyOK().WithPayload(payloads).WithXTotalCount(total).
		WithLink(n.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String())
}

func (n *webhookAPI) ListTasksOfWebhookExecution(ctx context.Context, params webhook.ListTasksOfWebhookExecutionParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	projectID, err := getProjectID(ctx, projectNameOrID)
	if err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.RequireProjectAccess(ctx, projectID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.requirePolicyInProject(ctx, projectID, params.WebhookPolicyID); err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.requireExecutionInPolicy(ctx, params.ExecutionID, params.WebhookPolicyID); err != nil {
		return n.SendError(ctx, err)
	}

	query, err := n.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return n.SendError(ctx, err)
	}

	total, err := n.webhookCtl.CountTasks(ctx, params.ExecutionID, query)
	if err != nil {
		return n.SendError(ctx, err)
	}

	tasks, err := n.webhookCtl.ListTasks(ctx, params.ExecutionID, query)
	if err != nil {
		return n.SendError(ctx, err)
	}

	var payloads []*models.Task
	for _, task := range tasks {
		p, err := convertTaskToPayload(task)
		if err != nil {
			return n.SendError(ctx, err)
		}
		payloads = append(payloads, p)
	}

	return webhook.NewListTasksOfWebhookExecutionOK().WithPayload(payloads).WithXTotalCount(total).
		WithLink(n.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String())
}

func (n *webhookAPI) GetLogsOfWebhookTask(ctx context.Context, params webhook.GetLogsOfWebhookTaskParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	projectID, err := getProjectID(ctx, projectNameOrID)
	if err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.RequireProjectAccess(ctx, projectID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.requirePolicyInProject(ctx, projectID, params.WebhookPolicyID); err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.requireExecutionInPolicy(ctx, params.ExecutionID, params.WebhookPolicyID); err != nil {
		return n.SendError(ctx, err)
	}
	if err := n.requireTaskInExecution(ctx, params.TaskID, params.ExecutionID); err != nil {
		return n.SendError(ctx, err)
	}

	l, err := n.webhookCtl.GetTaskLog(ctx, params.TaskID)
	if err != nil {
		return n.SendError(ctx, err)
	}

	return webhook.NewGetLogsOfWebhookTaskOK().WithPayload(string(l))
}

func (n *webhookAPI) LastTrigger(ctx context.Context, params webhook.LastTriggerParams) middleware.Responder {
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
	policies, err := n.webhookCtl.ListPolicies(ctx, query)
	if err != nil {
		return n.SendError(ctx, err)
	}
	triggers, err := n.constructPolicyWithTriggerTime(ctx, policies)
	if err != nil {
		return n.SendError(ctx, err)
	}

	return webhook.NewLastTriggerOK().WithPayload(triggers)
}

func (n *webhookAPI) GetSupportedEventTypes(ctx context.Context, params webhook.GetSupportedEventTypesParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	var notificationTypes = &models.SupportedWebhookEventTypes{}
	for _, notifyType := range notification.GetSupportedNotifyTypes() {
		notificationTypes.NotifyType = append(notificationTypes.NotifyType, models.NotifyType(notifyType))
	}

	for _, eventType := range notification.GetSupportedEventTypes() {
		notificationTypes.EventType = append(notificationTypes.EventType, models.EventType(eventType))
	}
	// currently only http type support payload format
	httpPayloadFormats := &models.PayloadFormat{NotifyType: models.NotifyType("http")}
	for _, formatType := range notification.GetSupportedPayloadFormats() {
		httpPayloadFormats.Formats = append(httpPayloadFormats.Formats, models.PayloadFormatType(formatType))
	}
	notificationTypes.PayloadFormats = []*models.PayloadFormat{httpPayloadFormats}

	return webhook.NewGetSupportedEventTypesOK().WithPayload(notificationTypes)
}

func (n *webhookAPI) validateTargets(policy *policy_model.Policy) (bool, error) {
	if len(policy.Targets) == 0 {
		return false, errors.New(nil).WithMessagef("empty notification target with policy %s", policy.Name).WithCode(errors.BadRequestCode)
	}
	for i, target := range policy.Targets {
		url, err := utils.ParseEndpoint(target.Address)
		if err != nil {
			return false, errors.New(err).WithCode(errors.BadRequestCode)
		}
		// Prevent SSRF security issue #3755
		target.Address = url.Scheme + "://" + url.Host + url.Path

		if !isNotifyTypeSupported(target.Type) {
			return false, errors.New(nil).WithMessagef("unsupported target type %s with policy %s", target.Type, policy.Name).WithCode(errors.BadRequestCode)
		}
		// don't allow set the payload format for slack type
		// slack should be migrated as a kind of payload in the future
		if len(target.PayloadFormat) > 0 && target.Type == "slack" {
			return false, errors.New(nil).WithMessage("set payload format is not allowed for slack").WithCode(errors.BadRequestCode)
		}

		if len(target.PayloadFormat) > 0 && target.Type == "teams" {
			return false, errors.New(nil).WithMessage("set payload format is not allowed for teams").WithCode(errors.BadRequestCode)
		}

		if len(target.PayloadFormat) > 0 && !isPayloadFormatSupported(target.PayloadFormat) {
			return false, errors.New(nil).WithMessagef("unsupported payload format type: %s", target.PayloadFormat).WithCode(errors.BadRequestCode)
		}
		// set payload format to Default is not specified when the type is http
		if len(target.PayloadFormat) == 0 && target.Type == "http" {
			policy.Targets[i].PayloadFormat = "Default"
		}
	}
	return true, nil
}

func (n *webhookAPI) validateEventTypes(policy *policy_model.Policy) (bool, error) {
	if len(policy.EventTypes) == 0 {
		return false, errors.New(nil).WithMessage("empty event type").WithCode(errors.BadRequestCode)
	}
	for _, eventType := range policy.EventTypes {
		if !isEventTypeSupported(eventType) {
			return false, errors.New(nil).WithMessagef("unsupported event type %s", eventType).WithCode(errors.BadRequestCode)
		}
	}
	return true, nil
}

// constructPolicyWithTriggerTime construct notification policy information displayed in UI
// including event type, enabled, creation time, last trigger time
func (n *webhookAPI) constructPolicyWithTriggerTime(ctx context.Context, policies []*policy_model.Policy) ([]*models.WebhookLastTrigger, error) {
	res := []*models.WebhookLastTrigger{}
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

			ltTime, err := n.webhookCtl.GetLastTriggerTime(ctx, t, policy.ID)
			if err != nil {
				return nil, err
			}
			if !ltTime.IsZero() {
				ply.LastTriggerTime = strfmt.DateTime(ltTime)
			}
			res = append(res, ply)
		}
	}
	return res, nil
}

func isEventTypeSupported(eventType string) bool {
	for _, t := range notification.GetSupportedEventTypes() {
		if t.String() == eventType {
			return true
		}
	}

	return false
}

func isNotifyTypeSupported(notifyType string) bool {
	for _, t := range notification.GetSupportedNotifyTypes() {
		if t.String() == notifyType {
			return true
		}
	}

	return false
}

func isPayloadFormatSupported(payloadFormat string) bool {
	for _, t := range notification.GetSupportedPayloadFormats() {
		if t.String() == payloadFormat {
			return true
		}
	}

	return false
}
