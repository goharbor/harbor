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

package googlegcr

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeGoogleGcr, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeGoogleGcr, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeGoogleGcr)
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	dockerRegistryAdapter, err := native.NewAdapter(registry)
	if err != nil {
		return nil, err
	}

	return &adapter{
		registry: registry,
		Adapter:  dockerRegistryAdapter,
	}, nil
}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return getAdapterInfo()
}

type adapter struct {
	*native.Adapter
	registry *model.Registry
}

var _ adp.Adapter = adapter{}

func (adapter) Info() (info *model.RegistryInfo, err error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeGoogleGcr,
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

func getAdapterInfo() *model.AdapterPattern {
	info := &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints: []*model.Endpoint{
				{
					Key:   "gcr.io",
					Value: "https://gcr.io",
				},
				{
					Key:   "us.gcr.io",
					Value: "https://us.gcr.io",
				},
				{
					Key:   "eu.gcr.io",
					Value: "https://eu.gcr.io",
				},
				{
					Key:   "asia.gcr.io",
					Value: "https://asia.gcr.io",
				},
			},
		},
		CredentialPattern: &model.CredentialPattern{
			AccessKeyType:    model.AccessKeyTypeFix,
			AccessKeyData:    "_json_key",
			AccessSecretType: model.AccessSecretTypeFile,
			AccessSecretData: "No Change",
		},
	}
	return info
}

// HealthCheck checks health status of a registry
func (a adapter) HealthCheck() (model.HealthStatus, error) {
	var err error
	if a.registry.Credential == nil ||
		len(a.registry.Credential.AccessKey) == 0 || len(a.registry.Credential.AccessSecret) == 0 {
		log.Errorf("no credential to ping registry %s", a.registry.URL)
		return model.Unhealthy, nil
	}
	if err = a.PingGet(); err != nil {
		log.Errorf("failed to ping registry %s: %v", a.registry.URL, err)
		return model.Unhealthy, nil
	}
	return model.Healthy, nil
}
