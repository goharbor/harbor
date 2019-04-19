// Copyright 2018 Project Harbor Authors
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

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
)

// LabelAPI handles requests for label management
type LabelAPI struct {
	label *models.Label
	BaseController
}

// Prepare ...
func (l *LabelAPI) Prepare() {
	l.BaseController.Prepare()
	method := l.Ctx.Request.Method
	if method == http.MethodGet {
		return
	}

	// POST, PUT, DELETE need login first
	if !l.SecurityCtx.IsAuthenticated() {
		l.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	if method == http.MethodPut || method == http.MethodDelete {
		id, err := l.GetInt64FromPath(":id")
		if err != nil || id <= 0 {
			l.SendBadRequestError(errors.New("invalid lable ID"))
			return
		}

		label, err := dao.GetLabel(id)
		if err != nil {
			l.SendInternalServerError(fmt.Errorf("failed to get label %d: %v", id, err))
			return
		}

		if label == nil || label.Deleted {
			l.SendNotFoundError(fmt.Errorf("label %d not found", id))
			return
		}

		l.label = label
	}
}

func (l *LabelAPI) requireAccess(label *models.Label, action rbac.Action, subresources ...rbac.Resource) bool {
	var hasPermission bool

	switch label.Scope {
	case common.LabelScopeGlobal:
		hasPermission = l.SecurityCtx.IsSysAdmin()
	case common.LabelScopeProject:
		if len(subresources) == 0 {
			subresources = append(subresources, rbac.ResourceLabel)
		}
		resource := rbac.NewProjectNamespace(label.ProjectID).Resource(subresources...)
		hasPermission = l.SecurityCtx.Can(action, resource)
	}

	if !hasPermission {
		if !l.SecurityCtx.IsAuthenticated() {
			l.SendUnAuthorizedError(errors.New("UnAuthorized"))
		} else {
			l.SendForbiddenError(errors.New(l.SecurityCtx.GetUsername()))
		}
		return false
	}

	return true
}

// Post creates a label
func (l *LabelAPI) Post() {
	label := &models.Label{}
	isValid, err := l.DecodeJSONReqAndValidate(label)
	if !isValid {
		l.SendBadRequestError(err)
		return
	}

	label.Level = common.LabelLevelUser

	switch label.Scope {
	case common.LabelScopeGlobal:
		label.ProjectID = 0
	case common.LabelScopeProject:
		exist, err := l.ProjectMgr.Exists(label.ProjectID)
		if err != nil {
			l.SendInternalServerError(fmt.Errorf("failed to check the existence of project %d: %v",
				label.ProjectID, err))
			return
		}
		if !exist {
			l.SendNotFoundError(fmt.Errorf("project %d not found", label.ProjectID))
			return
		}
	}

	if !l.requireAccess(label, rbac.ActionCreate) {
		return
	}

	labels, err := dao.ListLabels(&models.LabelQuery{
		Name:      label.Name,
		Level:     label.Level,
		Scope:     label.Scope,
		ProjectID: label.ProjectID,
	})
	if err != nil {
		l.SendInternalServerError(fmt.Errorf("failed to list labels: %v", err))
		return
	}
	if len(labels) > 0 {
		l.SendConflictError(errors.New("conflict label"))
		return
	}

	id, err := dao.AddLabel(label)
	if err != nil {
		l.SendInternalServerError(fmt.Errorf("failed to create label: %v", err))
		return
	}

	l.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Get the label specified by ID
func (l *LabelAPI) Get() {
	id, err := l.GetInt64FromPath(":id")
	if err != nil || id <= 0 {
		l.SendBadRequestError(fmt.Errorf("invalid label ID: %s", l.GetStringFromPath(":id")))
		return
	}

	label, err := dao.GetLabel(id)
	if err != nil {
		l.SendInternalServerError(fmt.Errorf("failed to get label %d: %v", id, err))
		return
	}

	if label == nil || label.Deleted {
		l.SendNotFoundError(fmt.Errorf("label %d not found", id))
		return
	}

	if !l.requireAccess(label, rbac.ActionRead) {
		return
	}

	l.Data["json"] = label
	l.ServeJSON()
}

// List labels according to the query strings
func (l *LabelAPI) List() {
	query := &models.LabelQuery{
		Name:           l.GetString("name"),
		FuzzyMatchName: true,
		Level:          common.LabelLevelUser,
	}

	scope := l.GetString("scope")
	if scope != common.LabelScopeGlobal && scope != common.LabelScopeProject {
		l.SendBadRequestError(fmt.Errorf("invalid scope: %s", scope))
		return
	}
	query.Scope = scope

	if scope == common.LabelScopeProject {
		projectIDStr := l.GetString("project_id")
		if len(projectIDStr) == 0 {
			l.SendBadRequestError(errors.New("project_id is required"))
			return
		}
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil || projectID <= 0 {
			l.SendBadRequestError(fmt.Errorf("invalid project_id: %s", projectIDStr))
			return
		}

		resource := rbac.NewProjectNamespace(projectID).Resource(rbac.ResourceLabel)
		if !l.SecurityCtx.Can(rbac.ActionList, resource) {
			if !l.SecurityCtx.IsAuthenticated() {
				l.SendUnAuthorizedError(errors.New("UnAuthorized"))
				return
			}
			l.SendForbiddenError(errors.New(l.SecurityCtx.GetUsername()))
			return
		}
		query.ProjectID = projectID
	}

	total, err := dao.GetTotalOfLabels(query)
	if err != nil {
		l.SendInternalServerError(fmt.Errorf("failed to get total count of labels: %v", err))
		return
	}

	query.Page, query.Size, err = l.GetPaginationParams()
	if err != nil {
		l.SendBadRequestError(err)
		return
	}

	labels, err := dao.ListLabels(query)
	if err != nil {
		l.SendInternalServerError(fmt.Errorf("failed to list labels: %v", err))
		return
	}

	l.SetPaginationHeader(total, query.Page, query.Size)
	l.Data["json"] = labels
	l.ServeJSON()
}

// Put updates the label
func (l *LabelAPI) Put() {
	if !l.requireAccess(l.label, rbac.ActionUpdate) {
		return
	}

	label := &models.Label{}
	if err := l.DecodeJSONReq(label); err != nil {
		l.SendBadRequestError(err)
		return
	}

	oldName := l.label.Name

	// only name, description and color can be changed
	l.label.Name = label.Name
	l.label.Description = label.Description
	l.label.Color = label.Color

	isValidate, err := l.Validate(l.label)
	if !isValidate {
		if err != nil {
			l.SendBadRequestError(err)
			return
		}
	}

	if l.label.Name != oldName {
		labels, err := dao.ListLabels(&models.LabelQuery{
			Name:      l.label.Name,
			Level:     l.label.Level,
			Scope:     l.label.Scope,
			ProjectID: l.label.ProjectID,
		})
		if err != nil {
			l.SendInternalServerError(fmt.Errorf("failed to list labels: %v", err))
			return
		}
		if len(labels) > 0 {
			l.SendConflictError(errors.New("conflict label"))
			return
		}
	}

	if err := dao.UpdateLabel(l.label); err != nil {
		l.SendInternalServerError(fmt.Errorf("failed to update label %d: %v", l.label.ID, err))
		return
	}

}

// Delete the label
func (l *LabelAPI) Delete() {
	if !l.requireAccess(l.label, rbac.ActionDelete) {
		return
	}

	id := l.label.ID
	if err := dao.DeleteResourceLabelByLabel(id); err != nil {
		l.SendInternalServerError(fmt.Errorf("failed to delete resource label mappings of label %d: %v", id, err))
		return
	}
	if err := dao.DeleteLabel(id); err != nil {
		l.SendInternalServerError(fmt.Errorf("failed to delete label %d: %v", id, err))
		return
	}
}

// ListResources lists the resources that the label is referenced by
func (l *LabelAPI) ListResources() {
	id, err := l.GetInt64FromPath(":id")
	if err != nil || id <= 0 {
		l.SendBadRequestError(errors.New("invalid label ID"))
		return
	}

	label, err := dao.GetLabel(id)
	if err != nil {
		l.SendInternalServerError(fmt.Errorf("failed to get label %d: %v", id, err))
		return
	}

	if label == nil || label.Deleted {
		l.SendNotFoundError(fmt.Errorf("label %d not found", id))
		return
	}

	if !l.requireAccess(label, rbac.ActionList, rbac.ResourceLabelResource) {
		return
	}

	/*
		result, err := core.GlobalController.GetPolicies(rep_models.QueryParameter{})
		if err != nil {
			l.HandleInternalServerError(fmt.Sprintf("failed to get policies: %v", err))
			return
		}
		policies := []*rep_models.ReplicationPolicy{}
		if result != nil {
			for _, policy := range result.Policies {
				for _, filter := range policy.Filters {
					if filter.Kind != replication.FilterItemKindLabel {
						continue
					}
					if filter.Value.(int64) == label.ID {
						policies = append(policies, policy)
					}
				}
			}
		}
	*/
	resources := map[string]interface{}{}
	resources["replication_policies"] = nil
	l.Data["json"] = resources
	l.ServeJSON()
}
