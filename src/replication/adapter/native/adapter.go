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
	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
)

const registryTypeNative model.RegistryType = "native"

func init() {
	if err := adp.RegisterFactory(registryTypeNative, func(registry *model.Registry) (adp.Adapter, error) {
		return newAdapter(registry)
	}); err != nil {
		log.Errorf("failed to register factory for %s: %v", registryTypeNative, err)
		return
	}
	log.Infof("the factory for adapter %s registered", registryTypeNative)
}

func newAdapter(registry *model.Registry) (*native, error) {
	reg, err := adp.NewDefaultImageRegistry(registry)
	if err != nil {
		return nil, err
	}
	return &native{
		registry:             registry,
		DefaultImageRegistry: reg,
	}, nil
}

type native struct {
	*adp.DefaultImageRegistry
	registry *model.Registry
}

var _ adp.Adapter = native{}

func (native) Info() (info *model.RegistryInfo, err error) {
	return &model.RegistryInfo{
		Type: registryTypeNative,
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeRepository,
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
func (native) PrepareForPush(*model.Resource) error { return nil }
