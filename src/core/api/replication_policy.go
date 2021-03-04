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
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/system"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"net/http"
	"strconv"

	replica "github.com/goharbor/harbor/src/controller/replication"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/event"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/registry"
)

// TODO rename the file to "replication.go"

// ReplicationPolicyAPI handles the replication policy requests
type ReplicationPolicyAPI struct {
	BaseController
	resource types.Resource
}

// Prepare ...
func (r *ReplicationPolicyAPI) Prepare() {
	r.BaseController.Prepare()
	if !r.SecurityCtx.IsAuthenticated() {
		r.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}
	r.resource = system.NewNamespace().Resource(rbac.ResourceReplicationPolicy)
}

// List the replication policies
func (r *ReplicationPolicyAPI) List() {
	if !r.SecurityCtx.Can(r.Context(), rbac.ActionList, r.resource) {
		r.SendForbiddenError(errors.New(r.SecurityCtx.GetUsername()))
		return
	}
	page, size, err := r.GetPaginationParams()
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	// TODO: support more query
	query := &model.PolicyQuery{
		Name: r.GetString("name"),
		Page: page,
		Size: size,
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
	if !r.SecurityCtx.Can(r.Context(), rbac.ActionCreate, r.resource) {
		r.SendForbiddenError(errors.New(r.SecurityCtx.GetUsername()))
		return
	}
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
		r.SendBadRequestError(fmt.Errorf("registry %d not found", registryID))
		return false
	}
	return true
}

// Get the specified replication policy
func (r *ReplicationPolicyAPI) Get() {
	if !r.SecurityCtx.Can(r.Context(), rbac.ActionRead, r.resource) {
		r.SendForbiddenError(errors.New(r.SecurityCtx.GetUsername()))
		return
	}
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
	if !r.SecurityCtx.Can(r.Context(), rbac.ActionUpdate, r.resource) {
		r.SendForbiddenError(errors.New(r.SecurityCtx.GetUsername()))
		return
	}
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
	if !r.SecurityCtx.Can(r.Context(), rbac.ActionDelete, r.resource) {
		r.SendForbiddenError(errors.New(r.SecurityCtx.GetUsername()))
		return
	}
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

	ctx := orm.Context()
	executions, err := replica.Ctl.ListExecutions(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"PolicyID": id,
		},
	})
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	for _, execution := range executions {
		if execution.Status != job.RunningStatus.String() {
			continue
		}
		r.SendPreconditionFailedError(fmt.Errorf("the policy %d has running executions, can not be deleted", id))
		return
	}
	for _, execution := range executions {
		if err = task.ExecMgr.Delete(ctx, execution.ID); err != nil {
			r.SendInternalServerError(err)
			return
		}
	}

	if err := replication.PolicyCtl.Remove(id); err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to delete the policy %d: %v", id, err))
		return
	}
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
