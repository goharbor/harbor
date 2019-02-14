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
	"fmt"

	"errors"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
	api_models "github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/core"
	rep_models "github.com/goharbor/harbor/src/replication/models"
)

// RepPolicyAPI handles /api/replicationPolicies /api/replicationPolicies/:id/enablement
type RepPolicyAPI struct {
	BaseController
}

// Prepare validates whether the user has system admin role
func (pa *RepPolicyAPI) Prepare() {
	pa.BaseController.Prepare()
	if !pa.SecurityCtx.IsAuthenticated() {
		pa.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	if !(pa.Ctx.Request.Method == http.MethodGet || pa.SecurityCtx.IsSysAdmin()) {
		pa.SendForbiddenError(errors.New(pa.SecurityCtx.GetUsername()))
		return
	}
}

// Get ...
func (pa *RepPolicyAPI) Get() {
	id, err := pa.GetIDFromURL()
	if err != nil {
		pa.SendBadRequestError(err)
		return
	}
	policy, err := core.GlobalController.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", id, err)
		pa.SendInternalServerError(fmt.Errorf("failed to get policy %d: %v", id, err))
		return

	}

	if policy.ID == 0 {
		pa.SendNotFoundError(fmt.Errorf("policy %d not found", id))
		return
	}

	resource := rbac.NewProjectNamespace(policy.ProjectIDs[0]).Resource(rbac.ResourceReplication)
	if !pa.SecurityCtx.Can(rbac.ActionRead, resource) {
		pa.SendForbiddenError(errors.New(pa.SecurityCtx.GetUsername()))
		return
	}

	ply, err := convertFromRepPolicy(pa.ProjectMgr, policy)
	if err != nil {
		pa.ParseAndHandleError(fmt.Sprintf("failed to convert from replication policy"), err)
		return
	}

	pa.Data["json"] = ply
	pa.ServeJSON()
}

// List ...
func (pa *RepPolicyAPI) List() {
	queryParam := rep_models.QueryParameter{
		Name: pa.GetString("name"),
	}
	projectIDStr := pa.GetString("project_id")
	if len(projectIDStr) > 0 {
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil || projectID <= 0 {
			pa.SendBadRequestError(fmt.Errorf("invalid project ID: %s", projectIDStr))
			return
		}
		queryParam.ProjectID = projectID
	}
	var err error
	queryParam.Page, queryParam.PageSize, err = pa.GetPaginationParams()
	if err != nil {
		pa.SendBadRequestError(err)
		return
	}

	result, err := core.GlobalController.GetPolicies(queryParam)
	if err != nil {
		log.Errorf("failed to get policies: %v, query parameters: %v", err, queryParam)
		pa.SendInternalServerError(fmt.Errorf("failed to get policies: %v, query parameters: %v", err, queryParam))
		return
	}

	var total int64
	policies := []*api_models.ReplicationPolicy{}
	if result != nil {
		total = result.Total
		for _, policy := range result.Policies {
			resource := rbac.NewProjectNamespace(policy.ProjectIDs[0]).Resource(rbac.ResourceReplication)
			if !pa.SecurityCtx.Can(rbac.ActionRead, resource) {
				continue
			}
			ply, err := convertFromRepPolicy(pa.ProjectMgr, *policy)
			if err != nil {
				pa.ParseAndHandleError(fmt.Sprintf("failed to convert from replication policy"), err)
				return
			}
			policies = append(policies, ply)
		}
	}

	pa.SetPaginationHeader(total, queryParam.Page, queryParam.PageSize)

	pa.Data["json"] = policies
	pa.ServeJSON()
}

// Post creates a replicartion policy
func (pa *RepPolicyAPI) Post() {
	policy := &api_models.ReplicationPolicy{}
	isValid, err := pa.DecodeJSONReqAndValidate(policy)
	if !isValid {
		pa.SendBadRequestError(err)
		return
	}

	// check the name
	exist, err := exist(policy.Name)
	if err != nil {
		pa.SendInternalServerError(fmt.Errorf("failed to check the existence of policy %s: %v", policy.Name, err))
		return
	}

	if exist {
		pa.SendConflictError(fmt.Errorf("name %s is already used", policy.Name))
		return
	}

	// check the existence of projects
	for _, project := range policy.Projects {
		pro, err := pa.ProjectMgr.Get(project.ProjectID)
		if err != nil {
			pa.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %d", project.ProjectID), err)
			return
		}
		if pro == nil {
			pa.SendNotFoundError(fmt.Errorf("project %d not found", project.ProjectID))
			return
		}
		project.Name = pro.Name
	}

	// check the existence of targets
	for _, target := range policy.Targets {
		t, err := dao.GetRepTarget(target.ID)
		if err != nil {
			pa.SendInternalServerError(fmt.Errorf("failed to get target %d: %v", target.ID, err))
			return
		}

		if t == nil {
			pa.SendNotFoundError(fmt.Errorf("target %d not found", target.ID))
			return
		}
	}

	// check the existence of labels
	for _, filter := range policy.Filters {
		if filter.Kind == replication.FilterItemKindLabel {
			labelID := filter.Value.(int64)
			label, err := dao.GetLabel(labelID)
			if err != nil {
				pa.SendInternalServerError(fmt.Errorf("failed to get label %d: %v", labelID, err))
				return
			}
			if label == nil || label.Deleted {
				pa.SendNotFoundError(fmt.Errorf("label %d not found", labelID))
				return
			}
		}
	}

	id, err := core.GlobalController.CreatePolicy(convertToRepPolicy(policy))
	if err != nil {
		pa.SendInternalServerError(fmt.Errorf("failed to create policy: %v", err))
		return
	}

	if policy.ReplicateExistingImageNow {
		go func() {
			if _, err = startReplication(id); err != nil {
				log.Errorf("failed to send replication signal for policy %d: %v", id, err)
				return
			}
			log.Infof("replication signal for policy %d sent", id)
		}()
	}

	pa.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

func exist(name string) (bool, error) {
	result, err := core.GlobalController.GetPolicies(rep_models.QueryParameter{
		Name: name,
	})
	if err != nil {
		return false, err
	}

	for _, policy := range result.Policies {
		if policy.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// Put updates the replication policy
func (pa *RepPolicyAPI) Put() {
	id, err := pa.GetIDFromURL()
	if err != nil {
		pa.SendBadRequestError(err)
		return
	}

	originalPolicy, err := core.GlobalController.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", id, err)
		pa.SendInternalServerError(fmt.Errorf("failed to get policy %d: %v", id, err))
		return
	}

	if originalPolicy.ID == 0 {
		pa.SendNotFoundError(fmt.Errorf("policy %d not found", id))
		return
	}

	policy := &api_models.ReplicationPolicy{}
	isValid, err := pa.DecodeJSONReqAndValidate(policy)
	if !isValid {
		pa.SendBadRequestError(err)
		return
	}

	policy.ID = id

	// check the name
	if policy.Name != originalPolicy.Name {
		exist, err := exist(policy.Name)
		if err != nil {
			pa.SendInternalServerError(fmt.Errorf("failed to check the existence of policy %s: %v", policy.Name, err))
			return
		}

		if exist {
			pa.SendConflictError(fmt.Errorf("name %s is already used", policy.Name))
			return
		}
	}

	// check the existence of projects
	for _, project := range policy.Projects {
		pro, err := pa.ProjectMgr.Get(project.ProjectID)
		if err != nil {
			pa.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %d", project.ProjectID), err)
			return
		}
		if pro == nil {
			pa.SendNotFoundError(fmt.Errorf("project %d not found", project.ProjectID))
			return
		}
		project.Name = pro.Name
	}

	// check the existence of targets
	for _, target := range policy.Targets {
		t, err := dao.GetRepTarget(target.ID)
		if err != nil {
			pa.SendInternalServerError(fmt.Errorf("failed to get target %d: %v", target.ID, err))
			return
		}

		if t == nil {
			pa.SendNotFoundError(fmt.Errorf("target %d not found", target.ID))
			return
		}
	}

	// check the existence of labels
	for _, filter := range policy.Filters {
		if filter.Kind == replication.FilterItemKindLabel {
			labelID := filter.Value.(int64)
			label, err := dao.GetLabel(labelID)
			if err != nil {
				pa.SendInternalServerError(fmt.Errorf("failed to get label %d: %v", labelID, err))
				return
			}
			if label == nil || label.Deleted {
				pa.SendNotFoundError(fmt.Errorf("label %d not found", labelID))
				return
			}
		}
	}

	if err = core.GlobalController.UpdatePolicy(convertToRepPolicy(policy)); err != nil {
		pa.SendInternalServerError(fmt.Errorf("failed to update policy %d: %v", id, err))
		return
	}

	if policy.ReplicateExistingImageNow {
		go func() {
			if _, err = startReplication(id); err != nil {
				log.Errorf("failed to send replication signal for policy %d: %v", id, err)
				return
			}
			log.Infof("replication signal for policy %d sent", id)
		}()
	}
}

// Delete the replication policy
func (pa *RepPolicyAPI) Delete() {
	id, err := pa.GetIDFromURL()
	if err != nil {
		pa.SendBadRequestError(err)
		return
	}

	policy, err := core.GlobalController.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", id, err)
		pa.SendInternalServerError(fmt.Errorf("failed to get policy %d: %v", id, err))
		return
	}

	if policy.ID == 0 {
		pa.SendNotFoundError(fmt.Errorf("policy %d not found", id))
		return
	}

	count, err := dao.GetTotalCountOfRepJobs(&models.RepJobQuery{
		PolicyID: id,
		Statuses: []string{models.JobRunning, models.JobRetrying, models.JobPending},
		// only get the transfer and delete jobs, do not get schedule job
		Operations: []string{models.RepOpTransfer, models.RepOpDelete},
	})
	if err != nil {
		log.Errorf("failed to filter jobs of policy %d: %v", id, err)
		pa.SendInternalServerError(fmt.Errorf("failed to filter jobs of policy %d: %v", id, err))
		return

	}
	if count > 0 {
		pa.SendPreconditionFailedError(errors.New("policy has running/retrying/pending jobs, can not be deleted"))
		return
	}

	if err = core.GlobalController.RemovePolicy(id); err != nil {
		log.Errorf("failed to delete policy %d: %v", id, err)
		pa.SendInternalServerError(fmt.Errorf("failed to delete policy %d: %v", id, err))
		return
	}
}

func convertFromRepPolicy(projectMgr promgr.ProjectManager, policy rep_models.ReplicationPolicy) (*api_models.ReplicationPolicy, error) {
	if policy.ID == 0 {
		return nil, nil
	}

	// populate simple properties
	ply := &api_models.ReplicationPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		ReplicateDeletion: policy.ReplicateDeletion,
		Trigger:           policy.Trigger,
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
	}

	// populate projects
	for _, projectID := range policy.ProjectIDs {
		project, err := projectMgr.Get(projectID)
		if err != nil {
			return nil, err
		}

		ply.Projects = append(ply.Projects, project)
	}

	// populate targets
	for _, targetID := range policy.TargetIDs {
		target, err := dao.GetRepTarget(targetID)
		if err != nil {
			return nil, err
		}
		target.Password = ""
		ply.Targets = append(ply.Targets, target)
	}

	// populate label used in label filter
	for _, filter := range policy.Filters {
		if filter.Kind == replication.FilterItemKindLabel {
			labelID := filter.Value.(int64)
			label, err := dao.GetLabel(labelID)
			if err != nil {
				return nil, err
			}
			filter.Value = label
		}
		ply.Filters = append(ply.Filters, filter)
	}

	// TODO call the method from replication controller
	errJobCount, err := dao.GetTotalCountOfRepJobs(&models.RepJobQuery{
		PolicyID: policy.ID,
		Statuses: []string{models.JobError},
	})
	if err != nil {
		return nil, err
	}
	ply.ErrorJobCount = errJobCount

	return ply, nil
}

func convertToRepPolicy(policy *api_models.ReplicationPolicy) rep_models.ReplicationPolicy {
	if policy == nil {
		return rep_models.ReplicationPolicy{}
	}

	ply := rep_models.ReplicationPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		Filters:           policy.Filters,
		ReplicateDeletion: policy.ReplicateDeletion,
		Trigger:           policy.Trigger,
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
	}

	for _, project := range policy.Projects {
		ply.ProjectIDs = append(ply.ProjectIDs, project.ProjectID)
		ply.Namespaces = append(ply.Namespaces, project.Name)
	}

	for _, target := range policy.Targets {
		ply.TargetIDs = append(ply.TargetIDs, target.ID)
	}

	return ply
}
