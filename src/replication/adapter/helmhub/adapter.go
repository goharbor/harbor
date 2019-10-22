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

package helmhub

import (
	"errors"
	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeHelmHub, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeHelmHub, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeHelmHub)
}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}

type adapter struct {
	registry *model.Registry
	client   *Client
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	return &adapter{
		registry: registry,
		client:   NewClient(registry),
	}, nil
}

func (a *adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeHelmHub,
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeChart,
		},
		SupportedResourceFilters: []*model.FilterStyle{
			{
				Type:  model.FilterTypeName,
				Style: model.FilterStyleTypeText,
			},
			{
				Type:  model.FilterTypeTag,
				Style: model.FilterStyleTypeText,
			},
		},
		SupportedTriggers: []model.TriggerType{
			model.TriggerTypeManual,
			model.TriggerTypeScheduled,
		},
	}, nil
}

func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	return errors.New("not supported")
}

// HealthCheck checks health status of a registry
func (a *adapter) HealthCheck() (model.HealthStatus, error) {
	err := a.client.checkHealthy()
	if err == nil {
		return model.Healthy, nil
	}
	return model.Unhealthy, err
}
