// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
)

// RepFilterAPI ...
type RepFilterAPI struct {
	BaseController
	filter *models.RepFilter
}

// Prepare ...
func (r *RepFilterAPI) Prepare() {
	r.BaseController.Prepare()

	if !r.SecurityCtx.IsAuthenticated() {
		r.HandleUnauthorized()
		return
	}

	if !r.SecurityCtx.IsSysAdmin() {
		r.HandleForbidden(r.SecurityCtx.GetUsername())
		return
	}

	if len(r.GetStringFromPath(":id")) > 0 {
		id, err := r.GetInt64FromPath(":id")
		if err != nil {
			r.HandleInternalServerError(err.Error())
			return
		}

		filter, err := dao.GetRepFilterByID(id)
		if err != nil {
			r.HandleInternalServerError(fmt.Sprintf("failed to get replication filter %d: %v", id, err))
			return
		}

		if filter == nil {
			r.HandleNotFound(fmt.Sprintf("replication filter %d not found", id))
			return
		}

		r.filter = filter
	}

}

// GetTypes returns all replication filter types
func (r *RepFilterAPI) GetTypes() {
	types, err := dao.GetRepFilterTypes()
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get replication filter types: %v", err))
		return
	}

	r.Data["json"] = types
	r.ServeJSON()
}

// List returns the filters defined for the replication rule
func (r *RepFilterAPI) List() {
	policyID, err := r.GetInt64("policy_id")
	if err != nil || policyID <= 0 {
		r.HandleBadRequest(fmt.Sprintf("invalid policy_id %s", r.GetString("policy_id")))
		return
	}

	policy, err := dao.GetRepPolicy(policyID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get policy %d: %v", policyID, err))
		return
	}

	if policy == nil {
		r.HandleNotFound(fmt.Sprintf("policy %d not found", policyID))
		return
	}

	filters, err := dao.GetRepFiltersByPolicyID(policyID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get filters of policy %d: %v", policyID, err))
		return
	}

	r.Data["json"] = filters
	r.ServeJSON()
}

// Post creates a new replication filter
func (r *RepFilterAPI) Post() {
	filter := &models.RepFilter{}
	r.DecodeJSONReqAndValidate(filter)

	policy, err := dao.GetRepPolicy(filter.RepPolicyID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get replication policy %d: %v", filter.RepPolicyID, err))
		return
	}

	if policy == nil {
		r.HandleNotFound(fmt.Sprintf("replication policy %d not found", filter.RepPolicyID))
		return
	}

	filterType, err := dao.GetRepFilterType(filter.RepFilterTypeID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get replication filter type %d: %v", filter.RepFilterTypeID, err))
		return
	}

	if filterType == nil {
		r.HandleNotFound(fmt.Sprintf("replication filter type %d not found", filter.RepFilterTypeID))
		return
	}

	id, err := dao.AddRepFilter(filter)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to add replication filter: %v", err))
		return
	}

	r.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put updates an exist replication filter
func (r *RepFilterAPI) Put() {
	filter := &models.RepFilter{}
	r.DecodeJSONReqAndValidate(filter)

	if filter.RepPolicyID != r.filter.RepPolicyID {
		policy, err := dao.GetRepPolicy(filter.RepPolicyID)
		if err != nil {
			r.HandleInternalServerError(fmt.Sprintf("failed to get replication policy %d: %v", filter.RepPolicyID, err))
			return
		}

		if policy == nil {
			r.HandleNotFound(fmt.Sprintf("replication policy %d not found", filter.RepPolicyID))
			return
		}
	}

	if filter.RepFilterTypeID != r.filter.RepFilterTypeID {
		filterType, err := dao.GetRepFilterType(filter.RepFilterTypeID)
		if err != nil {
			r.HandleInternalServerError(fmt.Sprintf("failed to get replication filter type %d: %v", filter.RepFilterTypeID, err))
			return
		}

		if filterType == nil {
			r.HandleNotFound(fmt.Sprintf("replication filter type %d not found", filter.RepFilterTypeID))
			return
		}
	}

	filter.ID = r.filter.ID
	if err := dao.UpdateRepFilter(filter); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to update replication filter %d: %v", filter.ID, err))
		return
	}
}

// Delete a replication filter
func (r *RepFilterAPI) Delete() {
	if err := dao.DeleteRepFilter(r.filter.ID); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to delete replication filter %d: %v", r.filter.ID, err))
		return
	}
}
