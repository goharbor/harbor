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
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/event"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/registry"
)

// ReplicationPolicyAPI handles the replication policy requests
type ReplicationPolicyAPI struct {
	BaseController
}

// Prepare ...
func (r *ReplicationPolicyAPI) Prepare() {
	r.BaseController.Prepare()
	if !r.SecurityCtx.IsSysAdmin() {
		if !r.SecurityCtx.IsAuthenticated() {
			r.HandleUnauthorized()
			return
		}
		r.HandleForbidden(r.SecurityCtx.GetUsername())
		return
	}
}

// List the replication policies
func (r *ReplicationPolicyAPI) List() {
	// TODO: support more query
	query := &model.PolicyQuery{
		Name: r.GetString("name"),
	}
	query.Page, query.Size = r.GetPaginationParams()

	total, policies, err := replication.PolicyCtl.List(query)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to list policies: %v", err))
		return
	}
	for _, policy := range policies {
		if err = populateRegistries(replication.RegistryMgr, policy); err != nil {
			r.HandleInternalServerError(fmt.Sprintf("failed to populate registries for policy %d: %v", policy.ID, err))
			return
		}
	}
	r.SetPaginationHeader(total, query.Page, query.Size)
	r.WriteJSONData(policies)
}

// Create the replication policy
func (r *ReplicationPolicyAPI) Create() {
	policy := &model.Policy{}
	r.DecodeJSONReqAndValidate(policy)

	if !r.validateName(policy) {
		return
	}
	if !r.validateRegistry(policy) {
		return
	}

	id, err := replication.PolicyCtl.Create(policy)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to create the policy: %v", err))
		return
	}
	r.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// make sure the policy name doesn't exist
func (r *ReplicationPolicyAPI) validateName(policy *model.Policy) bool {
	p, err := replication.PolicyCtl.GetByName(policy.Name)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get policy %s: %v", policy.Name, err))
		return false
	}
	if p != nil {
		r.HandleConflict(fmt.Sprintf("policy %s already exists", policy.Name))
		return false
	}
	return true
}

// make sure the registry referred exists
func (r *ReplicationPolicyAPI) validateRegistry(policy *model.Policy) bool {
	srcRegistry, err := replication.RegistryMgr.Get(policy.SrcRegistry.ID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get source registry %d: %v", policy.SrcRegistry.ID, err))
		return false
	}
	if srcRegistry == nil {
		r.HandleNotFound(fmt.Sprintf("source registry %d not found", policy.SrcRegistry.ID))
		return false
	}
	dstRegistry, err := replication.RegistryMgr.Get(policy.DestRegistry.ID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get destination registry %d: %v", policy.DestRegistry.ID, err))
		return false
	}
	if dstRegistry == nil {
		r.HandleNotFound(fmt.Sprintf("destination registry %d not found", policy.DestRegistry.ID))
		return false
	}
	// one of the source registry or destination registry must be local Harbor
	if srcRegistry.Type != model.RegistryTypeLocalHarbor && dstRegistry.Type != model.RegistryTypeLocalHarbor {
		r.HandleBadRequest(fmt.Sprintf("at least one of the registries' type is %s", model.RegistryTypeLocalHarbor))
		return false
	}

	return true
}

// Get the specified replication policy
func (r *ReplicationPolicyAPI) Get() {
	id, err := r.GetInt64FromPath(":id")
	if id <= 0 || err != nil {
		r.HandleBadRequest("invalid policy ID")
		return
	}

	policy, err := replication.PolicyCtl.Get(id)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get the policy %d: %v", id, err))
		return
	}
	if policy == nil {
		r.HandleNotFound(fmt.Sprintf("policy %d not found", id))
		return
	}
	if err = populateRegistries(replication.RegistryMgr, policy); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to populate registries for policy %d: %v", policy.ID, err))
		return
	}

	r.WriteJSONData(policy)
}

// Update the replication policy
func (r *ReplicationPolicyAPI) Update() {
	id, err := r.GetInt64FromPath(":id")
	if id <= 0 || err != nil {
		r.HandleBadRequest("invalid policy ID")
		return
	}

	originalPolicy, err := replication.PolicyCtl.Get(id)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get the policy %d: %v", id, err))
		return
	}
	if originalPolicy == nil {
		r.HandleNotFound(fmt.Sprintf("policy %d not found", id))
		return
	}

	policy := &model.Policy{}
	r.DecodeJSONReqAndValidate(policy)
	if policy.Name != originalPolicy.Name &&
		!r.validateName(policy) {
		return
	}

	if !r.validateRegistry(policy) {
		return
	}

	policy.ID = id
	if err := replication.PolicyCtl.Update(policy); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to update the policy %d: %v", id, err))
		return
	}
}

// Delete the replication policy
func (r *ReplicationPolicyAPI) Delete() {
	id, err := r.GetInt64FromPath(":id")
	if id <= 0 || err != nil {
		r.HandleBadRequest("invalid policy ID")
		return
	}

	policy, err := replication.PolicyCtl.Get(id)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get the policy %d: %v", id, err))
		return
	}
	if policy == nil {
		r.HandleNotFound(fmt.Sprintf("policy %d not found", id))
		return
	}

	_, executions, err := replication.OperationCtl.ListExecutions(&models.ExecutionQuery{
		PolicyID: id,
	})
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get the executions of policy %d: %v", id, err))
		return
	}

	for _, execution := range executions {
		if execution.Status == models.ExecutionStatusInProgress {
			r.HandleStatusPreconditionFailed(fmt.Sprintf("the policy %d has running executions, can not be deleted", id))
			return
		}
	}

	if err := replication.PolicyCtl.Remove(id); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to delete the policy %d: %v", id, err))
		return
	}
}

// ignore the credential for the registries
func populateRegistries(registryMgr registry.Manager, policy *model.Policy) error {
	if err := event.PopulateRegistries(registryMgr, policy); err != nil {
		return err
	}
	if policy.SrcRegistry != nil {
		policy.SrcRegistry.Credential = nil
	}
	if policy.DestRegistry != nil {
		policy.DestRegistry.Credential = nil
	}
	return nil
}
