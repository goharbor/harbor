package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/common/utils/log"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/pkg/notification"
)

// NotificationPolicyAPI ...
type NotificationPolicyAPI struct {
	BaseController
	project *models.Project
}

// notificationPolicyForUI defines the structure of notification policy info display in UI
type notificationPolicyForUI struct {
	EventType       string     `json:"event_type"`
	Enabled         bool       `json:"enabled"`
	CreationTime    *time.Time `json:"creation_time"`
	LastTriggerTime *time.Time `json:"last_trigger_time,omitempty"`
}

// Prepare ...
func (w *NotificationPolicyAPI) Prepare() {
	w.BaseController.Prepare()
	if !w.SecurityCtx.IsAuthenticated() {
		w.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	pid, err := w.GetInt64FromPath(":pid")
	if err != nil {
		w.SendBadRequestError(fmt.Errorf("failed to get project ID: %v", err))
		return
	}
	if pid <= 0 {
		w.SendBadRequestError(fmt.Errorf("invalid project ID: %d", pid))
		return
	}

	project, err := w.ProjectMgr.Get(pid)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get project %d: %v", pid, err))
		return
	}
	if project == nil {
		w.SendNotFoundError(fmt.Errorf("project %d not found", pid))
		return
	}
	w.project = project
}

// Get ...
func (w *NotificationPolicyAPI) Get() {
	if !w.validateRBAC(rbac.ActionRead, w.project.ProjectID) {
		return
	}

	id, err := w.GetIDFromURL()
	if err != nil {
		w.SendBadRequestError(err)
		return
	}

	policy, err := notification.PolicyMgr.Get(id)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get the notification policy %d: %v", id, err))
		return
	}
	if policy == nil {
		w.SendNotFoundError(fmt.Errorf("notification policy %d not found", id))
		return
	}

	if w.project.ProjectID != policy.ProjectID {
		w.SendBadRequestError(fmt.Errorf("notification policy %d with projectID %d not belong to project %d in URL", id, policy.ProjectID, w.project.ProjectID))
		return
	}

	w.WriteJSONData(policy)
}

// Post ...
func (w *NotificationPolicyAPI) Post() {
	if !w.validateRBAC(rbac.ActionCreate, w.project.ProjectID) {
		return
	}

	policy := &models.NotificationPolicy{}
	isValid, err := w.DecodeJSONReqAndValidate(policy)
	if !isValid {
		w.SendBadRequestError(err)
		return
	}

	if !w.validateTargets(policy) {
		return
	}

	if !w.validateEventTypes(policy) {
		return
	}

	if policy.ID != 0 {
		w.SendBadRequestError(fmt.Errorf("cannot accept policy creating request with ID: %d", policy.ID))
		return
	}

	policy.Creator = w.SecurityCtx.GetUsername()
	policy.ProjectID = w.project.ProjectID

	id, err := notification.PolicyMgr.Create(policy)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to create the notification policy: %v", err))
		return
	}
	w.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put ...
func (w *NotificationPolicyAPI) Put() {
	if !w.validateRBAC(rbac.ActionUpdate, w.project.ProjectID) {
		return
	}

	id, err := w.GetIDFromURL()
	if id < 0 || err != nil {
		w.SendBadRequestError(errors.New("invalid notification policy ID"))
		return
	}

	oriPolicy, err := notification.PolicyMgr.Get(id)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get the notification policy %d: %v", id, err))
		return
	}
	if oriPolicy == nil {
		w.SendNotFoundError(fmt.Errorf("notification policy %d not found", id))
		return
	}

	policy := &models.NotificationPolicy{}
	isValid, err := w.DecodeJSONReqAndValidate(policy)
	if !isValid {
		w.SendBadRequestError(err)
		return
	}

	if !w.validateTargets(policy) {
		return
	}

	if !w.validateEventTypes(policy) {
		return
	}

	if w.project.ProjectID != oriPolicy.ProjectID {
		w.SendBadRequestError(fmt.Errorf("notification policy %d with projectID %d not belong to project %d in URL", id, oriPolicy.ProjectID, w.project.ProjectID))
		return
	}

	policy.ID = id
	policy.ProjectID = w.project.ProjectID

	if err = notification.PolicyMgr.Update(policy); err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to update the notification policy: %v", err))
		return
	}
}

// List ...
func (w *NotificationPolicyAPI) List() {
	projectID := w.project.ProjectID
	if !w.validateRBAC(rbac.ActionList, projectID) {
		return
	}

	res, err := notification.PolicyMgr.List(projectID)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to list notification policies by projectID %d: %v", projectID, err))
		return
	}

	policies := []*models.NotificationPolicy{}
	if res != nil {
		for _, policy := range res {
			policies = append(policies, policy)
		}
	}

	w.WriteJSONData(policies)
}

// ListGroupByEventType lists notification policy trigger info grouped by event type for UI,
// displays event type, status(enabled/disabled), create time, last trigger time
func (w *NotificationPolicyAPI) ListGroupByEventType() {
	projectID := w.project.ProjectID
	if !w.validateRBAC(rbac.ActionList, projectID) {
		return
	}

	res, err := notification.PolicyMgr.List(projectID)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to list notification policies by projectID %d: %v", projectID, err))
		return
	}

	policies, err := constructPolicyWithTriggerTime(res)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to list the notification policy trigger information: %v", err))
		return
	}
	w.WriteJSONData(policies)
}

// Delete ...
func (w *NotificationPolicyAPI) Delete() {
	projectID := w.project.ProjectID
	if !w.validateRBAC(rbac.ActionDelete, projectID) {
		return
	}

	id, err := w.GetIDFromURL()
	if id < 0 || err != nil {
		w.SendBadRequestError(errors.New("invalid notification policy ID"))
		return
	}

	policy, err := notification.PolicyMgr.Get(id)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get the notification policy %d: %v", id, err))
		return
	}
	if policy == nil {
		w.SendNotFoundError(fmt.Errorf("notification policy %d not found", id))
		return
	}

	if projectID != policy.ProjectID {
		w.SendBadRequestError(fmt.Errorf("notification policy %d with projectID %d not belong to project %d in URL", id, policy.ProjectID, projectID))
		return
	}

	if err = notification.PolicyMgr.Delete(id); err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to delete notification policy %d: %v", id, err))
		return
	}
}

// Test ...
func (w *NotificationPolicyAPI) Test() {
	projectID := w.project.ProjectID
	if !w.validateRBAC(rbac.ActionCreate, projectID) {
		return
	}

	policy := &models.NotificationPolicy{}
	isValid, err := w.DecodeJSONReqAndValidate(policy)
	if !isValid {
		w.SendBadRequestError(err)
		return
	}

	if !w.validateTargets(policy) {
		return
	}

	if err := notification.PolicyMgr.Test(policy); err != nil {
		log.Errorf("notification policy %s test failed: %v", policy.Name, err)
		w.SendBadRequestError(fmt.Errorf("notification policy %s test failed", policy.Name))
		return
	}
}

func (w *NotificationPolicyAPI) validateRBAC(action rbac.Action, projectID int64) bool {
	if w.SecurityCtx.IsSysAdmin() {
		return true
	}

	return w.RequireProjectAccess(projectID, action, rbac.ResourceNotificationPolicy)
}

func (w *NotificationPolicyAPI) validateTargets(policy *models.NotificationPolicy) bool {
	if len(policy.Targets) == 0 {
		w.SendBadRequestError(fmt.Errorf("empty notification target with policy %s", policy.Name))
		return false
	}

	for _, target := range policy.Targets {
		url, err := utils.ParseEndpoint(target.Address)
		if err != nil {
			w.SendBadRequestError(err)
			return false
		}
		// Prevent SSRF security issue #3755
		target.Address = url.Scheme + "://" + url.Host + url.Path

		_, ok := notification.SupportedNotifyTypes[target.Type]
		if !ok {
			w.SendBadRequestError(fmt.Errorf("unsupport target type %s with policy %s", target.Type, policy.Name))
			return false
		}
	}

	return true
}

func (w *NotificationPolicyAPI) validateEventTypes(policy *models.NotificationPolicy) bool {
	if len(policy.EventTypes) == 0 {
		w.SendBadRequestError(errors.New("empty event type"))
		return false
	}

	for _, eventType := range policy.EventTypes {
		_, ok := notification.SupportedEventTypes[eventType]
		if !ok {
			w.SendBadRequestError(fmt.Errorf("unsupport event type %s", eventType))
			return false
		}
	}

	return true
}

func getLastTriggerTimeGroupByEventType(eventType string, policyID int64) (time.Time, error) {
	jobs, err := notification.JobMgr.ListJobsGroupByEventType(policyID)
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

// constructPolicyWithTriggerTime construct notification policy information displayed in UI
// including event type, enabled, creation time, last trigger time
func constructPolicyWithTriggerTime(policies []*models.NotificationPolicy) ([]*notificationPolicyForUI, error) {
	res := []*notificationPolicyForUI{}
	if policies != nil {
		for _, policy := range policies {
			for _, t := range policy.EventTypes {
				ply := &notificationPolicyForUI{
					EventType:    t,
					Enabled:      policy.Enabled,
					CreationTime: &policy.CreationTime,
				}
				if !policy.CreationTime.IsZero() {
					ply.CreationTime = &policy.CreationTime
				}

				ltTime, err := getLastTriggerTimeGroupByEventType(t, policy.ID)
				if err != nil {
					return nil, err
				}
				if !ltTime.IsZero() {
					ply.LastTriggerTime = &ltTime
				}
				res = append(res, ply)
			}
		}
	}
	return res, nil
}
