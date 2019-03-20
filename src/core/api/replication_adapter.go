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

	"github.com/goharbor/harbor/src/replication/ng/adapter"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

// ReplicationAdapterAPI handles the replication adapter requests
type ReplicationAdapterAPI struct {
	BaseController
}

// Prepare ...
func (r *ReplicationAdapterAPI) Prepare() {
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

// List the replication adapters
func (r *ReplicationAdapterAPI) List() {
	infos := []*adapter.Info{}
	for _, info := range adapter.ListAdapterInfos() {
		infos = append(infos, process(info))
	}
	r.WriteJSONData(infos)
}

// Get one specified replication adapter
func (r *ReplicationAdapterAPI) Get() {
	t := r.GetStringFromPath(":type")
	info := adapter.GetAdapterInfo(model.RegistryType(t))
	if info == nil {
		r.HandleNotFound(fmt.Sprintf("adapter for %s not found", t))
		return
	}
	info = process(info)
	r.WriteJSONData(info)
}

// merge "SupportedResourceTypes" into "SupportedResourceFilters" for UI to render easier
func process(info *adapter.Info) *adapter.Info {
	if info == nil {
		return nil
	}

	in := &adapter.Info{
		Type:              info.Type,
		Description:       info.Description,
		SupportedTriggers: info.SupportedTriggers,
	}

	filters := []*adapter.Filter{}
	for _, filter := range info.SupportedResourceFilters {
		if filter.Type != model.FilterTypeResource {
			filters = append(filters, filter)
		}
	}
	values := []string{}
	for _, resourceType := range info.SupportedResourceTypes {
		values = append(values, string(resourceType))
	}
	filters = append(filters, &adapter.Filter{
		Type:   model.FilterTypeResource,
		Style:  adapter.FilterStyleRadio,
		Values: values,
	})
	in.SupportedResourceFilters = filters

	return in
}
