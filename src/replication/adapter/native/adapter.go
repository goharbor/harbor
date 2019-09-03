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
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/goharbor/harbor/src/common/http/modifier"
	common_http_auth "github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	registry_pkg "github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeDockerRegistry, func(registry *model.Registry) (adp.Adapter, error) {
		return NewAdapter(registry)
	}); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeDockerRegistry, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeDockerRegistry)
}

var _ adp.Adapter = &Adapter{}

// Adapter implements an adapter for Docker registry. It can be used to all registries
// that implement the registry V2 API
type Adapter struct {
	sync.RWMutex
	*registry_pkg.Registry
	registry *model.Registry
	client   *http.Client
	clients  map[string]*registry_pkg.Repository // client for repositories
}

// NewAdapter returns an instance of the Adapter
func NewAdapter(registry *model.Registry) (*Adapter, error) {
	var cred modifier.Modifier
	if registry.Credential != nil && len(registry.Credential.AccessSecret) != 0 {
		if registry.Credential.Type == model.CredentialTypeSecret {
			cred = common_http_auth.NewSecretAuthorizer(registry.Credential.AccessSecret)
		} else {
			cred = auth.NewBasicAuthCredential(
				registry.Credential.AccessKey,
				registry.Credential.AccessSecret)
		}
	}
	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: util.GetHTTPTransport(registry.Insecure),
	}, cred, registry.TokenServiceURL)

	return NewAdapterWithCustomizedAuthorizer(registry, authorizer)
}

// NewAdapterWithCustomizedAuthorizer returns an instance of the Adapter with the customized authorizer
func NewAdapterWithCustomizedAuthorizer(registry *model.Registry, authorizer modifier.Modifier) (*Adapter, error) {
	transport := util.GetHTTPTransport(registry.Insecure)
	modifiers := []modifier.Modifier{
		&auth.UserAgentModifier{
			UserAgent: adp.UserAgentReplication,
		},
	}
	if authorizer != nil {
		modifiers = append(modifiers, authorizer)
	}
	client := &http.Client{
		Transport: registry_pkg.NewTransport(transport, modifiers...),
	}
	reg, err := registry_pkg.NewRegistry(registry.URL, client)
	if err != nil {
		return nil, err
	}
	return &Adapter{
		Registry: reg,
		registry: registry,
		client:   client,
		clients:  map[string]*registry_pkg.Repository{},
	}, nil
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

// FetchImages ...
func (a *Adapter) FetchImages(filters []*model.Filter) ([]*model.Resource, error) {
	repositories, err := a.getRepositories(filters)
	if err != nil {
		return nil, err
	}
	if len(repositories) == 0 {
		return nil, nil
	}
	for _, filter := range filters {
		if err = filter.DoFilter(&repositories); err != nil {
			return nil, err
		}
	}

	var rawResources = make([]*model.Resource, len(repositories))
	runner := utils.NewLimitedConcurrentRunner(adp.MaxConcurrency)
	defer runner.Cancel()

	for i, r := range repositories {
		index := i
		repo := r
		runner.AddTask(func() error {
			vTags, err := a.getVTags(repo.Name)
			if err != nil {
				return fmt.Errorf("List tags for repo '%s' error: %v", repo.Name, err)
			}
			if len(vTags) == 0 {
				return nil
			}
			for _, filter := range filters {
				if err = filter.DoFilter(&vTags); err != nil {
					return fmt.Errorf("Filter tags %v error: %v", vTags, err)
				}
			}
			if len(vTags) == 0 {
				return nil
			}
			tags := []string{}
			for _, vTag := range vTags {
				tags = append(tags, vTag.Name)
			}
			rawResources[index] = &model.Resource{
				Type:     model.ResourceTypeImage,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: repo.Name,
					},
					Vtags: tags,
				},
			}

			return nil
		})
	}
	runner.Wait()

	if runner.IsCancelled() {
		return nil, fmt.Errorf("FetchImages error when collect tags for repos")
	}

	var resources []*model.Resource
	for _, r := range rawResources {
		if r != nil {
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func (a *Adapter) getRepositories(filters []*model.Filter) ([]*adp.Repository, error) {
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

	result := []*adp.Repository{}
	for _, repository := range repositories {
		result = append(result, &adp.Repository{
			ResourceType: string(model.ResourceTypeImage),
			Name:         repository,
		})
	}
	return result, nil
}

func (a *Adapter) getVTags(repository string) ([]*adp.VTag, error) {
	tags, err := a.ListTag(repository)
	if err != nil {
		return nil, err
	}
	var result []*adp.VTag
	for _, tag := range tags {
		result = append(result, &adp.VTag{
			ResourceType: string(model.ResourceTypeImage),
			Name:         tag,
		})
	}
	return result, nil
}

// ManifestExist ...
func (a *Adapter) ManifestExist(repository, reference string) (bool, string, error) {
	client, err := a.getClient(repository)
	if err != nil {
		return false, "", err
	}
	digest, exist, err := client.ManifestExist(reference)
	return exist, digest, err
}

// PullManifest ...
func (a *Adapter) PullManifest(repository, reference string, accepttedMediaTypes []string) (distribution.Manifest, string, error) {
	client, err := a.getClient(repository)
	if err != nil {
		return nil, "", err
	}
	digest, mediaType, payload, err := client.PullManifest(reference, accepttedMediaTypes)
	if err != nil {
		return nil, "", err
	}
	if strings.Contains(mediaType, "application/json") {
		mediaType = schema1.MediaTypeManifest
	}
	manifest, _, err := registry_pkg.UnMarshal(mediaType, payload)
	if err != nil {
		return nil, "", err
	}
	return manifest, digest, nil
}

// PushManifest ...
func (a *Adapter) PushManifest(repository, reference, mediaType string, payload []byte) error {
	client, err := a.getClient(repository)
	if err != nil {
		return err
	}
	_, err = client.PushManifest(reference, mediaType, payload)
	return err
}

// DeleteManifest ...
func (a *Adapter) DeleteManifest(repository, reference string) error {
	client, err := a.getClient(repository)
	if err != nil {
		return err
	}
	digest := reference
	if !isDigest(digest) {
		dgt, exist, err := client.ManifestExist(reference)
		if err != nil {
			return err
		}
		if !exist {
			log.Debugf("the manifest of %s:%s doesn't exist", repository, reference)
			return nil
		}
		digest = dgt
	}
	return client.DeleteManifest(digest)
}

// BlobExist ...
func (a *Adapter) BlobExist(repository, digest string) (bool, error) {
	client, err := a.getClient(repository)
	if err != nil {
		return false, err
	}
	return client.BlobExist(digest)
}

// PullBlob ...
func (a *Adapter) PullBlob(repository, digest string) (int64, io.ReadCloser, error) {
	client, err := a.getClient(repository)
	if err != nil {
		return 0, nil, err
	}
	return client.PullBlob(digest)
}

// PushBlob ...
func (a *Adapter) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	client, err := a.getClient(repository)
	if err != nil {
		return err
	}
	return client.PushBlob(digest, size, blob)
}

func isDigest(str string) bool {
	return strings.Contains(str, ":")
}

// ListTag ...
func (a *Adapter) ListTag(repository string) ([]string, error) {
	client, err := a.getClient(repository)
	if err != nil {
		return []string{}, err
	}
	return client.ListTag()
}

func (a *Adapter) getClient(repository string) (*registry_pkg.Repository, error) {
	a.RLock()
	client, exist := a.clients[repository]
	a.RUnlock()
	if exist {
		return client, nil
	}

	return a.create(repository)
}

func (a *Adapter) create(repository string) (*registry_pkg.Repository, error) {
	a.Lock()
	defer a.Unlock()
	// double check
	client, exist := a.clients[repository]
	if exist {
		return client, nil
	}

	client, err := registry_pkg.NewRepository(repository, a.registry.URL, a.client)
	if err != nil {
		return nil, err
	}
	a.clients[repository] = client
	return client, nil
}
