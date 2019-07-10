package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
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
	id, err := w.GetIDFromURL()
	if err != nil {
		w.SendBadRequestError(err)
		return
	}

	if !w.validateRBAC(rbac.ActionRead, w.project.ProjectID) {
		return
	}

	policy, err := webhook.PolicyManager.Get(id)
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

	projectID := policy.ProjectID
	if !w.validateName(policy) {
		return
	}
	if w.project.ProjectID != projectID {
		w.SendBadRequestError(fmt.Errorf("project ID in url %d not match project ID %d in request body", w.project.ProjectID, projectID))
		return
	}

	if !w.validateTargets(policy) {
		return
	}

	if !w.validateHookTypes(policy) {
		return
	}

	policy.Creator = w.SecurityCtx.GetUsername()
	ply, err := convertFromAPIModel(policy)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to convert webhook policy from api model: %v", err))
		return
	}

	id, err := webhook.PolicyManager.Create(ply)

	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to create the webhook policy: %v", err))
		return
	}
	w.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put ...
func (w *WebhookPolicyAPI) Put() {
	id, err := w.GetIDFromURL()
	if id < 0 || err != nil {
		w.SendBadRequestError(errors.New("invalid webhook policy ID"))
		return
	}

	if !w.validateRBAC(rbac.ActionUpdate, w.project.ProjectID) {
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

	if w.project.ProjectID != policy.ProjectID {
		w.SendBadRequestError(fmt.Errorf("project ID in url %d not match project ID %d in request body", w.project.ProjectID, policy.ProjectID))
		return
	}

	oriPolicy, err := webhook.PolicyManager.Get(id)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get the webhook policy %d: %v", id, err))
		return
	}
	if oriPolicy == nil {
		w.SendNotFoundError(fmt.Errorf("webhook policy %d not found", id))
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

	if err = webhook.PolicyManager.Update(ply); err != nil {
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

	_, res, err := webhook.PolicyManager.List(projectID)
	if err != nil {
		log.Errorf("failed to get policies: %v, projectID: %d", err, projectID)
		w.SendInternalServerError(errors.New(""))
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

// Delete ...
func (w *WebhookPolicyAPI) Delete() {
	id, err := w.GetIDFromURL()
	if id < 0 || err != nil {
		w.SendBadRequestError(errors.New("invalid webhook policy ID"))
		return
	}

	policy, err := webhook.PolicyManager.Get(id)
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

	if !w.validateRBAC(rbac.ActionDelete, projectID) {
		return
	}

	if err = webhook.PolicyManager.Delete(id); err != nil {
		log.Error("failed to delete webhook policy %d: %v", id, err)
		w.SendInternalServerError(errors.New(""))
		return
	}
}

// Test ...
func (w *WebhookPolicyAPI) Test() {
	policy := &apiModels.WebhookPolicy{}
	isValid, err := w.DecodeJSONReqAndValidate(policy)
	if !isValid {
		w.SendBadRequestError(err)
	}

	if !w.validateName(policy) {
		return
	}

	if !w.validateTargets(policy) {
		return
	}

	if !w.validateHookTypes(policy) {
		return
	}

	ply, err := convertFromAPIModel(policy)
	if err := webhook.PolicyManager.Test(ply); err != nil {
		log.Errorf("webhook policy %s test failed: %v", policy.Name, err)
	}
}

func (w *WebhookPolicyAPI) validateRBAC(action rbac.Action, projectID int64) bool {
	project, err := w.ProjectMgr.Get(projectID)
	if err != nil {
		//w.SendInternalServerError(fmt.Errorf("failed to get project %d", projectID))
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

func (w *WebhookPolicyAPI) validateName(policy *apiModels.WebhookPolicy) bool {
	p, err := webhook.PolicyManager.GetByNameAndProjectID(policy.Name, policy.ProjectID)
	if err != nil {
		w.SendInternalServerError(fmt.Errorf("failed to get webhook policy %s: %v", policy.Name, err))
		return false
	}
	if p != nil {
		w.SendConflictError(fmt.Errorf("webhook policy %s in project %d already exists", policy.Name, policy.ProjectID))
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

func convertToAPIModel(policy *model.WebhookPolicy) (*apiModels.WebhookPolicy, error) {
	if policy.ID == 0 {
		return nil, nil
	}
	ply := &apiModels.WebhookPolicy{
		ID:           policy.ID,
		Name:         policy.Name,
		Description:  policy.Description,
		ProjectID:    policy.ProjectID,
		HookTypes:    policy.HookTypes,
		CreationTime: policy.CreationTime,
		UpdateTime:   policy.UpdateTime,
		Enabled:      policy.Enabled,
		Creator:      policy.Creator,
	}

	var targets []*apiModels.HookTarget
	for _, t := range policy.Targets {
		target := &apiModels.HookTarget{
			// do not return secret info
			Type:       t.Type,
			Address:    t.Address,
			Attachment: t.Attachment,
		}
		targets = append(targets, target)
	}
	ply.Targets = targets
	return ply, nil
}

func convertFromAPIModel(policy *apiModels.WebhookPolicy) (*model.WebhookPolicy, error) {
	ply := &model.WebhookPolicy{
		Name:         policy.Name,
		Description:  policy.Description,
		ProjectID:    policy.ProjectID,
		HookTypes:    policy.HookTypes,
		Creator:      policy.Creator,
		CreationTime: policy.CreationTime,
		UpdateTime:   policy.UpdateTime,
		Enabled:      policy.Enabled,
	}

	targets := []model.HookTarget{}
	for _, t := range policy.Targets {
		target := model.HookTarget{
			Type:       t.Type,
			Address:    t.Address,
			Attachment: t.Attachment,
			Secret:     t.Secret,
		}
		targets = append(targets, target)
	}
	ply.Targets = targets

	return ply, nil
}
