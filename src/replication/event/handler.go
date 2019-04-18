// Copyright Project Harbor Authors
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

package event

import (
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/replication/util"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/config"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/operation"
	"github.com/goharbor/harbor/src/replication/policy"
	"github.com/goharbor/harbor/src/replication/registry"
)

// Handler is the handler to handle event
type Handler interface {
	Handle(event *Event) error
}

// NewHandler ...
func NewHandler(policyCtl policy.Controller, registryMgr registry.Manager, opCtl operation.Controller) Handler {
	return &handler{
		policyCtl:   policyCtl,
		registryMgr: registryMgr,
		opCtl:       opCtl,
	}
}

type handler struct {
	policyCtl   policy.Controller
	registryMgr registry.Manager
	opCtl       operation.Controller
}

func (h *handler) Handle(event *Event) error {
	if event == nil || event.Resource == nil ||
		event.Resource.Metadata == nil ||
		len(event.Resource.Metadata.Vtags) == 0 {
		return errors.New("invalid event")
	}
	var policies []*model.Policy
	var err error
	switch event.Type {
	case EventTypeImagePush, EventTypeChartUpload,
		EventTypeImageDelete, EventTypeChartDelete:
		policies, err = h.getRelatedPolicies(event.Resource)
	default:
		return fmt.Errorf("unsupported event type %s", event.Type)
	}
	if err != nil {
		return err
	}

	if len(policies) == 0 {
		log.Debugf("no policy found for the event %v, do nothing", event)
		return nil
	}

	for _, policy := range policies {
		if err := PopulateRegistries(h.registryMgr, policy); err != nil {
			return err
		}
		id, err := h.opCtl.StartReplication(policy, event.Resource, model.TriggerTypeEventBased)
		if err != nil {
			return err
		}
		log.Debugf("%s event received, the replication execution %d started", event.Type, id)
	}
	return nil
}

func (h *handler) getRelatedPolicies(resource *model.Resource) ([]*model.Policy, error) {
	_, policies, err := h.policyCtl.List()
	if err != nil {
		return nil, err
	}
	result := []*model.Policy{}
	for _, policy := range policies {
		// disabled
		if !policy.Enabled {
			continue
		}
		// currently, the events are produced only by local Harbor,
		// so they should only apply to the policies whose source registry is local Harbor
		if !(policy.SrcRegistry == nil || policy.SrcRegistry.ID == 0) {
			continue
		}
		// has no trigger
		if policy.Trigger == nil {
			continue
		}
		// trigger type isn't event based
		if policy.Trigger.Type != model.TriggerTypeEventBased {
			continue
		}
		// doesn't replicate deletion
		if resource.Deleted && !policy.Deletion {
			continue
		}
		// doesn't match the name filter
		m, err := match(policy.Filters, resource)
		if err != nil {
			return nil, err
		}
		if !m {
			continue
		}
		result = append(result, policy)
	}
	return result, nil
}

// TODO unify the match logic with other?
func match(filters []*model.Filter, resource *model.Resource) (bool, error) {
	match := true
	repository := resource.Metadata.Repository.Name
	for _, filter := range filters {
		if filter.Type != model.FilterTypeName {
			continue
		}
		m, err := util.Match(filter.Value.(string), repository)
		if err != nil {
			return false, err
		}
		if !m {
			match = false
			break
		}
	}
	return match, nil
}

// PopulateRegistries populates the source registry and destination registry properties for policy
func PopulateRegistries(registryMgr registry.Manager, policy *model.Policy) error {
	if policy == nil {
		return nil
	}
	registry, err := getRegistry(registryMgr, policy.SrcRegistry)
	if err != nil {
		return err
	}
	policy.SrcRegistry = registry
	registry, err = getRegistry(registryMgr, policy.DestRegistry)
	if err != nil {
		return err
	}
	policy.DestRegistry = registry
	return nil
}

func getRegistry(registryMgr registry.Manager, registry *model.Registry) (*model.Registry, error) {
	if registry == nil || registry.ID == 0 {
		return GetLocalRegistry(), nil
	}
	reg, err := registryMgr.Get(registry.ID)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, fmt.Errorf("registry %d not found", registry.ID)
	}
	return reg, nil
}

// GetLocalRegistry returns the info of the local Harbor registry
func GetLocalRegistry() *model.Registry {
	return &model.Registry{
		Type:    model.RegistryTypeHarbor,
		Name:    "Local",
		URL:     config.Config.RegistryURL,
		CoreURL: config.Config.CoreURL,
		Status:  "healthy",
		Credential: &model.Credential{
			Type: model.CredentialTypeSecret,
			// use secret to do the auth for the local Harbor
			AccessSecret: config.Config.JobserviceSecret,
		},
		Insecure: true,
	}
}
