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
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
	"net/http"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeGoogleGcr, func(registry *model.Registry) (adp.Adapter, error) {
		return newAdapter(registry)
	}); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeGoogleGcr, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeGoogleGcr)
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	var credential auth.Credential
	if registry.Credential != nil && len(registry.Credential.AccessSecret) != 0 {
		credential = auth.NewBasicAuthCredential(
			registry.Credential.AccessKey,
			registry.Credential.AccessSecret)
	}
	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: util.GetHTTPTransport(registry.Insecure),
	}, credential)

	reg, err := adp.NewDefaultImageRegistryWithCustomizedAuthorizer(registry, authorizer)
	if err != nil {
		return nil, err
	}

	return &adapter{
		registry:             registry,
		DefaultImageRegistry: reg,
	}, nil
}

type adapter struct {
	*adp.DefaultImageRegistry
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

// PrepareForPush nothing need to do.
func (a adapter) PrepareForPush(resources []*model.Resource) error {
	return nil
}
