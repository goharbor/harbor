package api

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/webhook"
)

// WebhookExecutionAPI ...
type WebhookExecutionAPI struct {
	BaseController
}

// Prepare ...
func (w *WebhookExecutionAPI) Prepare() {
	w.BaseController.Prepare()
	if !w.SecurityCtx.IsAuthenticated() {
		w.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}
}

// TODO: list execution by webhook type with UI design
// List ...
func (w *WebhookExecutionAPI) List() {
	policyID, err := w.GetInt64("policy_id")
	if err != nil || policyID <= 0 {
		w.SendBadRequestError(fmt.Errorf("invalid policy_id: %s", w.GetString("policy_id")))
	}

	policy, err := webhook.PolicyManager.Get(policyID)
	if err != nil {
		log.Error("failed to get policy %d: %v", policyID, err)
		w.SendInternalServerError(errors.New(""))
		return
	}
	if policy.ID == 0 {
		w.SendNotFoundError(fmt.Errorf("policy %v not found", policyID))
		return
	}

	if !w.validateRBAC(rbac.ActionList, policy.ProjectID) {
		return
	}

	query := &models.WebhookExecutionQuery{
		PolicyID: policyID,
	}

	query.Statuses = w.GetStrings("status")

	startTimeStr := w.GetString("start_time")
	if len(startTimeStr) != 0 {
		i, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			w.SendBadRequestError(fmt.Errorf("invalid start_time: %s", startTimeStr))
			return
		}
		t := time.Unix(i, 0)
		query.StartTime = &t
	}
	endTimeStr := w.GetString("end_time")
	if len(startTimeStr) != 0 {
		i, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			w.SendBadRequestError(fmt.Errorf("invalid end_time: %s", endTimeStr))
			return
		}
		t := time.Unix(i, 0)
		query.EndTime = &t
	}

	query.Page, query.Size, err = w.GetPaginationParams()
	if err != nil {
		w.SendBadRequestError(err)
	}

	total, executions, err := webhook.ExecutionCtl.ListWebhookExecutions(query)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to list webhook executions: %v", err))
		return
	}
	w.SetPaginationHeader(total, query.Page, query.Size)
	w.WriteJSONData(executions)
}

// Delete ...
func (w *WebhookExecutionAPI) Delete() {
	id, err := w.GetIDFromURL()
	if id <= 0 || err != nil {
		w.SendBadRequestError(errors.New("invalid webhook execution ID"))
		return
	}

	execution, err := webhook.ExecutionCtl.GetWebhookExecution(id)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get the webhook execution %d: %v", id, err))
		return
	}
	if execution == nil {
		w.SendNotFoundError(fmt.Errorf("webhook execution %d not found", id))
		return
	}

	if execution.Status == models.JobRunning {
		w.SendBadRequestError(fmt.Errorf("webhook execution status in %s, cannot delete", execution.Status))
		return
	}

	policy, err := webhook.PolicyManager.Get(execution.PolicyID)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get the webhook policy %d by execution id %d: %v", id, execution.PolicyID, err))
		return
	}
	if policy == nil {
		w.SendNotFoundError(fmt.Errorf("webhook policy %d by execution id %d not found", execution.PolicyID, id))
		return
	}

	if w.validateRBAC(rbac.ActionDelete, policy.ProjectID) {
		return
	}

	if err = webhook.ExecutionCtl.DeleteWebhookExecution(id); err != nil {
		log.Error("failed to delete webhook execution %d: %v", id, err)
		w.SendInternalServerError(errors.New(""))
		return
	}
}

func (w *WebhookExecutionAPI) validateRBAC(action rbac.Action, projectID int64) bool {
	project, err := w.ProjectMgr.Get(projectID)
	if err != nil {
		w.ParseAndHandleError(fmt.Sprintf("failed to get project %d", projectID), err)
		return false
	}

	resource := rbac.NewProjectNamespace(project.ProjectID).Resource(rbac.ResourceWebhookPolicy)
	if !w.SecurityCtx.Can(action, resource) {
		w.SendForbiddenError(errors.New(w.SecurityCtx.GetUsername()))
		return false
	}
	return true
}
