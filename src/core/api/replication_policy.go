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

	common_model "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/event"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/registry"
)

// TODO rename the file to "replication.go"

// ReplicationPolicyAPI handles the replication policy requests
type ReplicationPolicyAPI struct {
	BaseController
}

// Prepare ...
func (r *ReplicationPolicyAPI) Prepare() {
	r.BaseController.Prepare()
	if !r.SecurityCtx.IsSysAdmin() {
		if !r.SecurityCtx.IsAuthenticated() {
			r.SendUnAuthorizedError(errors.New("UnAuthorized"))
			return
		}
		r.SendForbiddenError(errors.New(r.SecurityCtx.GetUsername()))
		return
	}
}

// List the replication policies
func (r *ReplicationPolicyAPI) List() {
	page, size, err := r.GetPaginationParams()
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	// TODO: support more query
	query := &model.PolicyQuery{
		Name: r.GetString("name"),
		Pagination: common_model.Pagination{
			Page: page,
			Size: size,
		},
	}

	total, policies, err := replication.PolicyCtl.List(query)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to list policies: %v", err))
		return
	}
	for _, policy := range policies {
		if err = populateRegistries(replication.RegistryMgr, policy); err != nil {
			r.SendInternalServerError(fmt.Errorf("failed to populate registries for policy %d: %v", policy.ID, err))
			return
		}
	}
	r.SetPaginationHeader(total, query.Page, query.Size)
	r.WriteJSONData(policies)
}

// Create the replication policy
func (r *ReplicationPolicyAPI) Create() {
	policy := &model.Policy{}
	isValid, err := r.DecodeJSONReqAndValidate(policy)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}

	if !r.validateName(policy) {
		return
	}
	if !r.validateRegistry(policy) {
		return
	}

	policy.Creator = r.SecurityCtx.GetUsername()
	id, err := replication.PolicyCtl.Create(policy)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to create the policy: %v", err))
		return
	}
	r.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// make sure the policy name doesn't exist
func (r *ReplicationPolicyAPI) validateName(policy *model.Policy) bool {
	p, err := replication.PolicyCtl.GetByName(policy.Name)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to get policy %s: %v", policy.Name, err))
		return false
	}
	if p != nil {
		r.SendConflictError(fmt.Errorf("policy %s already exists", policy.Name))
		return false
	}
	return true
}

// make sure the registry referred exists
func (r *ReplicationPolicyAPI) validateRegistry(policy *model.Policy) bool {
	var registryID int64
	if policy.SrcRegistry != nil && policy.SrcRegistry.ID > 0 {
		registryID = policy.SrcRegistry.ID
	} else {
		registryID = policy.DestRegistry.ID
	}
	registry, err := replication.RegistryMgr.Get(registryID)
	if err != nil {
		r.SendConflictError(fmt.Errorf("failed to get registry %d: %v", registryID, err))
		return false
	}
	if registry == nil {
		r.SendNotFoundError(fmt.Errorf("registry %d not found", registryID))
		return false
	}
	return true
}

// Get the specified replication policy
func (r *ReplicationPolicyAPI) Get() {
	id, err := r.GetInt64FromPath(":id")
	if id <= 0 || err != nil {
		r.SendBadRequestError(errors.New("invalid policy ID"))
		return
	}

	policy, err := replication.PolicyCtl.Get(id)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to get the policy %d: %v", id, err))
		return
	}
	if policy == nil {
		r.SendNotFoundError(fmt.Errorf("policy %d not found", id))
		return
	}
	if err = populateRegistries(replication.RegistryMgr, policy); err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to populate registries for policy %d: %v", policy.ID, err))
		return
	}

	r.WriteJSONData(policy)
}

// Update the replication policy
func (r *ReplicationPolicyAPI) Update() {
	id, err := r.GetInt64FromPath(":id")
	if id <= 0 || err != nil {
		r.SendBadRequestError(errors.New("invalid policy ID"))
		return
	}

	originalPolicy, err := replication.PolicyCtl.Get(id)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to get the policy %d: %v", id, err))
		return
	}
	if originalPolicy == nil {
		r.SendNotFoundError(fmt.Errorf("policy %d not found", id))
		return
	}

	policy := &model.Policy{}
	isValid, err := r.DecodeJSONReqAndValidate(policy)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}

	if policy.Name != originalPolicy.Name &&
		!r.validateName(policy) {
		return
	}

	if !r.validateRegistry(policy) {
		return
	}

	policy.ID = id
	if err := replication.PolicyCtl.Update(policy); err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to update the policy %d: %v", id, err))
		return
	}
}

// Delete the replication policy
func (r *ReplicationPolicyAPI) Delete() {
	id, err := r.GetInt64FromPath(":id")
	if id <= 0 || err != nil {
		r.SendBadRequestError(errors.New("invalid policy ID"))
		return
	}

	policy, err := replication.PolicyCtl.Get(id)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to get the policy %d: %v", id, err))
		return
	}
	if policy == nil {
		r.SendNotFoundError(fmt.Errorf("policy %d not found", id))
		return
	}

	isRunning, err := hasRunningExecutions(id)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to check the execution status of policy %d: %v", id, err))
		return
	}

	if isRunning {
		r.SendPreconditionFailedError(fmt.Errorf("the policy %d has running executions, can not be deleted", id))
		return
	}

	if err := replication.PolicyCtl.Remove(id); err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to delete the policy %d: %v", id, err))
		return
	}
}

// the execution's status will not be updated if it is not queried
// so need to check the status of tasks to determine the status of
// the execution
func hasRunningExecutions(policyID int64) (bool, error) {
	_, executions, err := replication.OperationCtl.ListExecutions(&models.ExecutionQuery{
		PolicyID: policyID,
	})
	if err != nil {
		return false, err
	}
	for _, execution := range executions {
		if execution.Status != models.ExecutionStatusInProgress {
			continue
		}
		_, tasks, err := replication.OperationCtl.ListTasks(&models.TaskQuery{
			ExecutionID: execution.ID,
		})
		if err != nil {
			return false, err
		}
		for _, task := range tasks {
			if isTaskRunning(task) {
				return true, nil
			}
		}
	}
	return false, nil
}

// return true if the status of the task is running or pending
func isTaskRunning(task *models.Task) bool {
	if task == nil {
		return false
	}
	switch task.Status {
	case models.TaskStatusSucceed,
		models.TaskStatusStopped,
		models.TaskStatusFailed:
		return false
	}
	return true
}

// ignore the credential for the registries
func populateRegistries(registryMgr registry.Manager, policy *model.Policy) error {
	if err := event.PopulateRegistries(registryMgr, policy); err != nil {
		return err
	}
	if policy.SrcRegistry != nil {
		hideAccessSecret(policy.SrcRegistry.Credential)
	}
	if policy.DestRegistry != nil {
		hideAccessSecret(policy.DestRegistry.Credential)
	}
	return nil
}
