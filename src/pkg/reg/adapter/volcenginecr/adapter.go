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

package volcenginecr

import (
	"errors"
	"path"
	"strings"

	volcCR "github.com/volcengine/volcengine-go-sdk/service/cr"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	volcSession "github.com/volcengine/volcengine-go-sdk/volcengine/session"

	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
	"github.com/goharbor/harbor/src/pkg/registry/auth/bearer"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeVolcCR, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeVolcCR, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeVolcCR)
}

type factory struct{}

/**
	* Implement Factory Interface
**/
var _ adp.Factory = &factory{}

type adapter struct {
	*native.Adapter
	registryName *string
	volcCrClient *volcCR.CR
	registry     *model.Registry
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return getAdapterInfo()
}

func getAdapterInfo() *model.AdapterPattern {
	return &model.AdapterPattern{}
}

func newAdapter(registry *model.Registry) (a *adapter, err error) {
	// get region and registryName from url
	region, registryName, err := getRegionRegistryName(registry.URL)
	if err != nil {
		log.Errorf("getRegion failed. error=%v", err)
		return nil, err
	}

	// Create VolcCR API client
	config := volcengine.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(registry.Credential.AccessKey, registry.Credential.AccessSecret, "")).
		WithRegion(region)
	sess, err := volcSession.NewSession(config)
	if err != nil {
		log.Errorf("getSession error. error=%v", err)
		return nil, err
	}
	client := volcCR.New(sess)

	// Get AuthorizationToken for docker login
	bearRealm, bearService, err := getRealmService(registry.URL, registry.Insecure)
	if err != nil {
		log.Error("fail to ping the registry", "url", registry.URL)
		return nil, err
	}
	cred := NewAuth(client, registryName)
	var transport = util.GetHTTPTransport(registry.Insecure)
	authorizer := bearer.NewAuthorizer(bearRealm, bearService, cred, transport)

	return &adapter{
		registry:     registry,
		registryName: &registryName,
		volcCrClient: client,
		Adapter:      native.NewAdapterWithAuthorizer(registry, authorizer),
	}, nil
}

func (a *adapter) Info() (info *model.RegistryInfo, err error) {
	info = &model.RegistryInfo{
		Type: model.RegistryTypeVolcCR,
		SupportedResourceTypes: []string{
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
		SupportedTriggers: []string{
			model.TriggerTypeManual,
			model.TriggerTypeScheduled,
		},
	}
	return
}

func (a *adapter) PrepareForPush(resources []*model.Resource) (err error) {
	for _, resource := range resources {
		if resource == nil {
			return errors.New("the resource cannot be null")
		}
		if resource.Metadata == nil {
			return errors.New("[volcengine-cr.PrepareForPush] the metadata of resource cannot be null")
		}
		if resource.Metadata.Repository == nil {
			return errors.New("[volcengine-cr.PrepareForPush] the namespace of resource cannot be null")
		}
		if len(resource.Metadata.Repository.Name) == 0 {
			return errors.New("[volcengine-cr.PrepareForPush] the name of the namespace cannot be null")
		}
		var paths = strings.Split(resource.Metadata.Repository.Name, "/")
		if len(paths) < 2 {
			return errors.New("[volcengine-cr.PrepareForPush] the name of the repository and namespace cannot be null")
		}
		var namespace = paths[0]
		var repository = path.Join(paths[1:]...)

		log.Debugf("namespace=%s", namespace)
		err = a.createNamespace(namespace)
		if err != nil {
			log.Errorf("PrepareForPush error :%v", err)
			return
		}
		log.Debugf("namespace=%s, repository=%s", namespace, repository)
		err = a.createRepository(namespace, repository)
		if err != nil {
			log.Errorf("PrepareForPush error :%v", err)
			return
		}
	}
	return
}
