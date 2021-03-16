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
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/system"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
)

// ReplicationAdapterAPI handles the replication adapter requests
type ReplicationAdapterAPI struct {
	BaseController
	resource types.Resource
}

// Prepare ...
func (r *ReplicationAdapterAPI) Prepare() {
	r.BaseController.Prepare()
	if !r.SecurityCtx.IsAuthenticated() {
		r.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}
	r.resource = system.NewNamespace().Resource(rbac.ResourceReplicationAdapter)
}

// List the replication adapters
func (r *ReplicationAdapterAPI) List() {
	if !r.SecurityCtx.Can(r.Context(), rbac.ActionList, r.resource) {
		r.SendForbiddenError(errors.New(r.SecurityCtx.GetUsername()))
		return
	}
	types := []model.RegistryType{}
	types = append(types, adapter.ListRegisteredAdapterTypes()...)
	r.WriteJSONData(types)
}

// ListAdapterInfos the replication adapter infos
func (r *ReplicationAdapterAPI) ListAdapterInfos() {
	if !r.SecurityCtx.Can(r.Context(), rbac.ActionList, r.resource) {
		r.SendForbiddenError(errors.New(r.SecurityCtx.GetUsername()))
		return
	}
	r.WriteJSONData(adapter.ListAdapterInfos())
}
