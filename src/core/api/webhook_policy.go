package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	apiModels "github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/webhook"
	"github.com/goharbor/harbor/src/webhook/model"
)

// WebhookPolicyAPI ...
type WebhookPolicyAPI struct {
	BaseController
	project *models.Project
}

// Prepare ...
func (w *WebhookPolicyAPI) Prepare() {
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
func (w *WebhookPolicyAPI) Get() {
	if !w.validateRBAC(rbac.ActionRead, w.project.ProjectID) {
		return
	}

	id, err := w.GetIDFromURL()
	if err != nil {
		w.SendBadRequestError(err)
		return
	}

	policy, err := webhook.PolicyCtl.Get(id)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get the webhook policy %d: %v", id, err))
		return
	}
	if policy == nil {
		w.SendNotFoundError(fmt.Errorf("webhook policy %d not found", id))
		return
	}

	projectID := policy.ProjectID
	if projectID == 0 {
		w.SendNotFoundError(fmt.Errorf("webhook policy %d with projectID %d not found", id, projectID))
		return
	}
	if w.project.ProjectID != projectID {
		w.SendBadRequestError(fmt.Errorf("webhook policy %d with projectID %d not belong to project %d in URL", id, projectID, w.project.ProjectID))
	}

	ply, err := convertToAPIModel(policy)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to convert webhook policy to api model: %v", err))
		return
	}

	w.WriteJSONData(ply)
}

// Post ...
func (w *WebhookPolicyAPI) Post() {
	if !w.validateRBAC(rbac.ActionCreate, w.project.ProjectID) {
		return
	}

	policy := &apiModels.WebhookPolicy{}
	isValid, err := w.DecodeJSONReqAndValidate(policy)
	if !isValid {
		w.SendBadRequestError(err)
		return
	}

	if !w.validatePolicyExist() {
		return
	}

	if !w.validateTargets(policy) {
		return
	}

	if !w.validateHookTypes(policy) {
		return
	}

	if policy.ID != 0 {
		w.SendBadRequestError(fmt.Errorf("cannot accept policy creating request with ID: %d", policy.ID))
		return
	}

	ply, err := convertFromAPIModel(policy)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to convert webhook policy from api model: %v", err))
		return
	}

	ply.Creator = w.SecurityCtx.GetUsername()
	ply.ProjectID = w.project.ProjectID
	id, err := webhook.PolicyCtl.Create(ply)

	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to create the webhook policy: %v", err))
		return
	}
	w.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put ...
func (w *WebhookPolicyAPI) Put() {
	if !w.validateRBAC(rbac.ActionUpdate, w.project.ProjectID) {
		return
	}

	id, err := w.GetIDFromURL()
	if id < 0 || err != nil {
		w.SendBadRequestError(errors.New("invalid webhook policy ID"))
		return
	}

	oriPolicy, err := webhook.PolicyCtl.Get(id)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get the webhook policy %d: %v", id, err))
		return
	}
	if oriPolicy == nil {
		w.SendNotFoundError(fmt.Errorf("webhook policy %d not found", id))
		return
	}

	policy := &apiModels.WebhookPolicy{}
	isValid, err := w.DecodeJSONReqAndValidate(policy)
	if !isValid {
		w.SendBadRequestError(err)
		return
	}

	if !w.validateTargets(policy) {
		return
	}

	if !w.validateHookTypes(policy) {
		return
	}

	if w.project.ProjectID != oriPolicy.ProjectID {
		w.SendBadRequestError(fmt.Errorf("webhook policy %d with projectID %d not belong to project %d in URL", id, oriPolicy.ProjectID, w.project.ProjectID))
		return
	}

	ply, err := convertFromAPIModel(policy)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to convert webhook policy from api model: %v", err))
		return
	}
	ply.ID = id
	ply.ProjectID = w.project.ProjectID

	if err = webhook.PolicyCtl.Update(ply); err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to update the webhook policy: %v", err))
		return
	}
}

// List ...
func (w *WebhookPolicyAPI) List() {
	projectID := w.project.ProjectID
	if !w.validateRBAC(rbac.ActionList, projectID) {
		return
	}

	_, res, err := webhook.PolicyCtl.List(projectID)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to list webhook policies by projectID %d: %v", projectID, err))
		return
	}

	policies := []*apiModels.WebhookPolicy{}
	if res != nil {
		for _, policy := range res {
			ply, err := convertToAPIModel(policy)
			if err != nil {
				w.SendInternalServerError(fmt.Errorf("failed to convert webhook policy to api model: %v", err))
				return
			}
			policies = append(policies, ply)
		}
	}

	w.WriteJSONData(policies)
}

// ListGroupByHookType lists webhook policy trigger info grouped by hook type for UI,
// displays hook type, status(enabled/disabled), create time, last trigger time
func (w *WebhookPolicyAPI) ListGroupByHookType() {
	projectID := w.project.ProjectID
	if !w.validateRBAC(rbac.ActionList, projectID) {
		return
	}

	_, res, err := webhook.PolicyCtl.List(projectID)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to list webhook policies by projectID %d: %v", projectID, err))
		return
	}

	policies, err := constructPolicyForUI(res)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to list the webhook policy trigger information: %v", err))
		return
	}
	w.WriteJSONData(policies)
}

// Delete ...
func (w *WebhookPolicyAPI) Delete() {
	projectID := w.project.ProjectID
	if !w.validateRBAC(rbac.ActionDelete, projectID) {
		return
	}

	id, err := w.GetIDFromURL()
	if id < 0 || err != nil {
		w.SendBadRequestError(errors.New("invalid webhook policy ID"))
		return
	}

	policy, err := webhook.PolicyCtl.Get(id)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get the webhook policy %d: %v", id, err))
		return
	}
	if policy == nil {
		w.SendNotFoundError(fmt.Errorf("webhook policy %d not found", id))
		return
	}

	if policy.ProjectID == 0 {
		w.SendNotFoundError(fmt.Errorf("webhook policy %d with projectID %d not found", id, projectID))
		return
	}

	if projectID != policy.ProjectID {
		w.SendBadRequestError(fmt.Errorf("webhook policy %d with projectID %d not belong to project %d in URL", id, policy.ProjectID, projectID))
		return
	}

	if err = webhook.PolicyCtl.Delete(id); err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to delete webhook policy %d: %v", id, err))
		return
	}
}

// Test ...
func (w *WebhookPolicyAPI) Test() {
	policy := &apiModels.WebhookPolicy{}
	isValid, err := w.DecodeJSONReqAndValidate(policy)
	if !isValid {
		w.SendBadRequestError(err)
		return
	}

	if !w.validateTargets(policy) {
		return
	}

	ply, err := convertFromAPIModel(policy)
	if err := webhook.PolicyCtl.Test(ply); err != nil {
		w.SendBadRequestError(fmt.Errorf("webhook policy %s test failed: %v", policy.Name, err))
		return
	}
}

func (w *WebhookPolicyAPI) validateRBAC(action rbac.Action, projectID int64) bool {
	if w.SecurityCtx.IsSysAdmin() {
		return true
	}

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

func (w *WebhookPolicyAPI) validatePolicyExist() bool {
	count, _, err := webhook.PolicyCtl.List(w.project.ProjectID)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to list webhook policy in project %d: %v", w.project.ProjectID, err))
		return false
	}
	// for the sake of UI, user can create only one policy for each project
	if count == 1 {
		w.SendConflictError(fmt.Errorf("webhook policy in project %d already exists", w.project.ProjectID))
		return false
	}
	return true
}

func (w *WebhookPolicyAPI) validateName(policy *apiModels.WebhookPolicy) bool {
	p, err := webhook.PolicyCtl.GetByNameAndProjectID(policy.Name, w.project.ProjectID)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get webhook policy %s: %v", policy.Name, err))
		return false
	}
	if p != nil {
		w.SendConflictError(fmt.Errorf("webhook policy %s in project %d already exists", policy.Name, w.project.ProjectID))
		return false
	}
	return true
}

func (w *WebhookPolicyAPI) validateTargets(policy *apiModels.WebhookPolicy) bool {
	if len(policy.Targets) == 0 {
		w.SendBadRequestError(fmt.Errorf("empty webhook target with policy %s", policy.Name))
		return false
	}

	for _, target := range policy.Targets {
		if target.Address == "" {
			w.SendBadRequestError(fmt.Errorf("empty webhook target address with policy %s", policy.Name))
			return false
		}
		url, err := utils.ParseEndpoint(target.Address)
		if err != nil {
			w.SendBadRequestError(err)
			return false
		}
		// Prevent SSRF security issue #3755
		target.Address = url.Scheme + "://" + url.Host + url.Path

		t, ok := webhook.SupportedSendTypes[target.Type]
		if !ok || t != model.ValidType {
			w.SendBadRequestError(fmt.Errorf("unsupport target type %s with policy %s", target.Type, policy.Name))
			return false
		}
	}

	return true
}

func (w *WebhookPolicyAPI) validateHookTypes(policy *apiModels.WebhookPolicy) bool {
	if len(policy.HookTypes) == 0 {
		w.SendBadRequestError(errors.New("empty hook type"))
		return false
	}

	for _, hookType := range policy.HookTypes {
		t, ok := webhook.SupportedHookTypes[hookType]
		if !ok || t != model.ValidType {
			w.SendBadRequestError(fmt.Errorf("unsupport hook type %s", hookType))
			return false
		}
	}

	return true
}

func getLastTriggerTimeByHookType(hookType string, policyID int64) (time.Time, error) {
	infos, err := webhook.JobCtl.ListLastTriggerInfos(policyID)
	if err != nil {
		return time.Time{}, err
	}

	for _, info := range infos {
		if hookType == info.HookType {
			return info.CreationTime, nil
		}
	}
	return time.Time{}, nil
}

func convertToAPIModel(policy *models.WebhookPolicy) (*apiModels.WebhookPolicy, error) {
	if policy.ID == 0 {
		return nil, nil
	}
	ply := &apiModels.WebhookPolicy{
		ID:           policy.ID,
		Name:         policy.Name,
		Description:  policy.Description,
		HookTypes:    policy.HookTypes,
		CreationTime: policy.CreationTime,
		UpdateTime:   policy.UpdateTime,
		Enabled:      policy.Enabled,
		Creator:      policy.Creator,
	}

	var targets []*apiModels.HookTarget
	for _, t := range policy.Targets {
		target := &apiModels.HookTarget{
			Type:           t.Type,
			Address:        t.Address,
			Token:          t.Token,
			SkipCertVerify: t.SkipCertVerify,
		}
		targets = append(targets, target)
	}
	ply.Targets = targets
	return ply, nil
}

func convertFromAPIModel(policy *apiModels.WebhookPolicy) (*models.WebhookPolicy, error) {
	ply := &models.WebhookPolicy{
		Name:         policy.Name,
		Description:  policy.Description,
		HookTypes:    policy.HookTypes,
		CreationTime: policy.CreationTime,
		UpdateTime:   policy.UpdateTime,
		Enabled:      policy.Enabled,
	}

	targets := []models.HookTarget{}
	for _, t := range policy.Targets {
		target := models.HookTarget{
			Type:           t.Type,
			Address:        t.Address,
			Token:          t.Token,
			SkipCertVerify: t.SkipCertVerify,
		}
		targets = append(targets, target)
	}
	ply.Targets = targets

	return ply, nil
}

// constructPolicyForUI construct webhook policy information displayed in UI
// including hook type, enabled, creation time, last trigger time
func constructPolicyForUI(policies []*models.WebhookPolicy) ([]*apiModels.WebhookPolicyForUI, error) {
	res := []*apiModels.WebhookPolicyForUI{}
	if policies != nil {
		for _, policy := range policies {
			for _, t := range policy.HookTypes {
				ply := &apiModels.WebhookPolicyForUI{
					HookType:     t,
					Enabled:      policy.Enabled,
					CreationTime: &policy.CreationTime,
				}
				if !policy.CreationTime.IsZero() {
					ply.CreationTime = &policy.CreationTime
				}

				ltTime, err := getLastTriggerTimeByHookType(t, policy.ID)
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
