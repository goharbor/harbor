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

package native

import (
	"net/http"

	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeDockerRegistry, func(registry *model.Registry) (adp.Adapter, error) {
		return newAdapter(registry)
	}); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeDockerRegistry, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeDockerRegistry)
}

func newAdapter(registry *model.Registry) (*Native, error) {
	reg, err := adp.NewDefaultImageRegistry(registry)
	if err != nil {
		return nil, err
	}
	return &Native{
		registry:             registry,
		DefaultImageRegistry: reg,
	}, nil
}

// NewWithClient ...
func NewWithClient(registry *model.Registry, client *http.Client) (*Native, error) {
	reg, err := adp.NewDefaultRegistryWithClient(registry, client)
	if err != nil {
		return nil, err
	}
	return &Native{
		registry:             registry,
		DefaultImageRegistry: reg,
	}, nil
}

// Native is adapter to native docker registry
type Native struct {
	*adp.DefaultImageRegistry
	registry *model.Registry
}

var _ adp.Adapter = Native{}

// Info ...
func (Native) Info() (info *model.RegistryInfo, err error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeDockerRegistry,
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeImage,
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

// PrepareForPush nothing need to do.
func (Native) PrepareForPush([]*model.Resource) error { return nil }
