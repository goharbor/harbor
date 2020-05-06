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
	"fmt"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/registry"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/filter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeDockerRegistry, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeDockerRegistry, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeDockerRegistry)
}

var _ adp.Adapter = &Adapter{}

type factory struct{}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return NewAdapter(r), nil
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}

var (
	_ adp.Adapter          = (*Adapter)(nil)
	_ adp.ArtifactRegistry = (*Adapter)(nil)
)

// Adapter implements an adapter for Docker registry. It can be used to all registries
// that implement the registry V2 API
type Adapter struct {
	registry *model.Registry
	registry.Client
}

// NewAdapter returns an instance of the Adapter
func NewAdapter(reg *model.Registry) *Adapter {
	adapter := &Adapter{
		registry: reg,
	}
	username, password := "", ""
	if reg.Credential != nil {
		username = reg.Credential.AccessKey
		password = reg.Credential.AccessSecret
	}
	adapter.Client = registry.NewClient(reg.URL, username, password, reg.Insecure)
	return adapter
}

// NewAdapterWithAuthorizer returns an instance of the Adapter with provided authorizer
func NewAdapterWithAuthorizer(reg *model.Registry, authorizer lib.Authorizer) *Adapter {
	return &Adapter{
		registry: reg,
		Client:   registry.NewClientWithAuthorizer(reg.URL, authorizer, reg.Insecure),
	}
}

// Info returns the basic information about the adapter
func (a *Adapter) Info() (info *model.RegistryInfo, err error) {
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

// PrepareForPush does nothing
func (a *Adapter) PrepareForPush([]*model.Resource) error {
	return nil
}

// HealthCheck checks health status of a registry
func (a *Adapter) HealthCheck() (model.HealthStatus, error) {
	var err error
	if a.registry.Credential == nil ||
		(len(a.registry.Credential.AccessKey) == 0 && len(a.registry.Credential.AccessSecret) == 0) {
		err = a.PingSimple()
	} else {
		err = a.Ping()
	}
	if err != nil {
		log.Errorf("failed to ping registry %s: %v", a.registry.URL, err)
		return model.Unhealthy, nil
	}
	return model.Healthy, nil
}

// FetchArtifacts ...
func (a *Adapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {
	repositories, err := a.listRepositories(filters)
	if err != nil {
		return nil, err
	}
	if len(repositories) == 0 {
		return nil, nil
	}

	var rawResources = make([]*model.Resource, len(repositories))
	runner := utils.NewLimitedConcurrentRunner(adp.MaxConcurrency)
	defer runner.Cancel()

	for i, r := range repositories {
		index := i
		repo := r
		runner.AddTask(func() error {
			artifacts, err := a.listArtifacts(repo.Name, filters)
			if err != nil {
				return fmt.Errorf("failed to list artifacts of repository %s: %v", repo.Name, err)
			}
			if len(artifacts) == 0 {
				return nil
			}
			rawResources[index] = &model.Resource{
				Type:     model.ResourceTypeImage,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: repo.Name,
					},
					Artifacts: artifacts,
				},
			}

			return nil
		})
	}
	runner.Wait()

	if runner.IsCancelled() {
		return nil, fmt.Errorf("FetchArtifacts error when collect tags for repos")
	}

	var resources []*model.Resource
	for _, r := range rawResources {
		if r != nil {
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func (a *Adapter) listRepositories(filters []*model.Filter) ([]*model.Repository, error) {
	pattern := ""
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			pattern = filter.Value.(string)
			break
		}
	}
	var repositories []string
	var err error
	// if the pattern of repository name filter is a specific repository name, just returns
	// the parsed repositories and will check the existence later when filtering the tags
	if paths, ok := util.IsSpecificPath(pattern); ok {
		repositories = paths
	} else {
		// search repositories from catalog API
		repositories, err = a.Catalog()
		if err != nil {
			return nil, err
		}
	}

	var result []*model.Repository
	for _, repository := range repositories {
		result = append(result, &model.Repository{
			Name: repository,
		})
	}
	return filter.DoFilterRepositories(result, filters)
}

func (a *Adapter) listArtifacts(repository string, filters []*model.Filter) ([]*model.Artifact, error) {
	tags, err := a.ListTags(repository)
	if err != nil {
		return nil, err
	}
	var artifacts []*model.Artifact
	for _, tag := range tags {
		artifacts = append(artifacts, &model.Artifact{
			Tags: []string{tag},
		})
	}
	return filter.DoFilterArtifacts(artifacts, filters)
}

// PingSimple checks whether the registry is available. It checks the connectivity and certificate (if TLS enabled)
// only, regardless of 401/403 error.
func (a *Adapter) PingSimple() error {
	err := a.Ping()
	if err == nil {
		return nil
	}
	if errors.IsErr(err, errors.UnAuthorizedCode) || errors.IsErr(err, errors.ForbiddenCode) {
		return nil
	}
	return err
}

// DeleteTag isn't supported for docker registry
func (a *Adapter) DeleteTag(repository, tag string) error {
	return errors.New("the tag deletion isn't supported")
}
