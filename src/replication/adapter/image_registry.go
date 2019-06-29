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

package adapter

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/replication/filter"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/goharbor/harbor/src/common/http/modifier"
	common_http_auth "github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/utils/log"
	registry_pkg "github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

// const definition
const (
	UserAgentReplication = "harbor-replication-service"
)

// ImageRegistry defines the capabilities that an image registry should have
type ImageRegistry interface {
	FetchImages(filters []*model.Filter) ([]*model.Resource, error)
	ManifestExist(repository, reference string) (exist bool, digest string, err error)
	PullManifest(repository, reference string, accepttedMediaTypes []string) (manifest distribution.Manifest, digest string, err error)
	PushManifest(repository, reference, mediaType string, payload []byte) error
	// the "reference" can be "tag" or "digest", the function needs to handle both
	DeleteManifest(repository, reference string) error
	BlobExist(repository, digest string) (exist bool, err error)
	PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error)
	PushBlob(repository, digest string, size int64, blob io.Reader) error
}

// Repository defines an repository object, it can be image repository, chart repository and etc.
type Repository struct {
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
}

// GetName returns the name
func (r *Repository) GetName() string {
	return r.Name
}

// GetFilterableType returns the filterable type
func (r *Repository) GetFilterableType() filter.FilterableType {
	return filter.FilterableTypeRepository
}

// GetResourceType returns the resource type
func (r *Repository) GetResourceType() string {
	return r.ResourceType
}

// GetLabels returns the labels
func (r *Repository) GetLabels() []string {
	return nil
}

// VTag defines an vTag object, it can be image tag, chart version and etc.
type VTag struct {
	ResourceType string   `json:"resource_type"`
	Name         string   `json:"name"`
	Labels       []string `json:"labels"`
}

// GetFilterableType returns the filterable type
func (v *VTag) GetFilterableType() filter.FilterableType {
	return filter.FilterableTypeVTag
}

// GetResourceType returns the resource type
func (v *VTag) GetResourceType() string {
	return v.ResourceType
}

// GetName returns the name
func (v *VTag) GetName() string {
	return v.Name
}

// GetLabels returns the labels
func (v *VTag) GetLabels() []string {
	return v.Labels
}

// DefaultImageRegistry provides a default implementation for interface ImageRegistry
type DefaultImageRegistry struct {
	sync.RWMutex
	*registry_pkg.Registry
	registry *model.Registry
	client   *http.Client
	clients  map[string]*registry_pkg.Repository
}

// NewDefaultRegistryWithClient returns an instance of DefaultImageRegistry
func NewDefaultRegistryWithClient(registry *model.Registry, client *http.Client) (*DefaultImageRegistry, error) {
	reg, err := registry_pkg.NewRegistry(registry.URL, client)
	if err != nil {
		return nil, err
	}

	return &DefaultImageRegistry{
		Registry: reg,
		client:   client,
		registry: registry,
		clients:  map[string]*registry_pkg.Repository{},
	}, nil
}

// NewDefaultImageRegistry returns an instance of DefaultImageRegistry
func NewDefaultImageRegistry(registry *model.Registry) (*DefaultImageRegistry, error) {
	var authorizer modifier.Modifier
	if registry.Credential != nil && len(registry.Credential.AccessSecret) != 0 {
		var cred modifier.Modifier
		if registry.Credential.Type == model.CredentialTypeSecret {
			cred = common_http_auth.NewSecretAuthorizer(registry.Credential.AccessSecret)
		} else {
			cred = auth.NewBasicAuthCredential(
				registry.Credential.AccessKey,
				registry.Credential.AccessSecret)
		}
		authorizer = auth.NewStandardTokenAuthorizer(&http.Client{
			Transport: util.GetHTTPTransport(registry.Insecure),
		}, cred, registry.TokenServiceURL)
	}
	return NewDefaultImageRegistryWithCustomizedAuthorizer(registry, authorizer)
}

// NewDefaultImageRegistryWithCustomizedAuthorizer returns an instance of DefaultImageRegistry with the customized authorizer
func NewDefaultImageRegistryWithCustomizedAuthorizer(registry *model.Registry, authorizer modifier.Modifier) (*DefaultImageRegistry, error) {
	transport := util.GetHTTPTransport(registry.Insecure)
	modifiers := []modifier.Modifier{
		&auth.UserAgentModifier{
			UserAgent: UserAgentReplication,
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
	return &DefaultImageRegistry{
		Registry: reg,
		client:   client,
		registry: registry,
		clients:  map[string]*registry_pkg.Repository{},
	}, nil
}

func (d *DefaultImageRegistry) getClient(repository string) (*registry_pkg.Repository, error) {
	d.RLock()
	client, exist := d.clients[repository]
	d.RUnlock()
	if exist {
		return client, nil
	}

	return d.create(repository)
}

func (d *DefaultImageRegistry) create(repository string) (*registry_pkg.Repository, error) {
	d.Lock()
	defer d.Unlock()
	// double check
	client, exist := d.clients[repository]
	if exist {
		return client, nil
	}

	client, err := registry_pkg.NewRepository(repository, d.registry.URL, d.client)
	if err != nil {
		return nil, err
	}
	d.clients[repository] = client
	return client, nil
}

// HealthCheck checks health status of a registry
func (d *DefaultImageRegistry) HealthCheck() (model.HealthStatus, error) {
	var err error
	if d.registry.Credential == nil ||
		(len(d.registry.Credential.AccessKey) == 0 && len(d.registry.Credential.AccessSecret) == 0) {
		err = d.PingSimple()
	} else {
		err = d.Ping()
	}
	if err != nil {
		log.Errorf("failed to ping registry %s: %v", d.registry.URL, err)
		return model.Unhealthy, nil
	}
	return model.Healthy, nil
}

// FetchImages ...
func (d *DefaultImageRegistry) FetchImages(namespaces []string, filters []*model.Filter) ([]*model.Resource, error) {
	return nil, errors.New("not implemented")
}

// ManifestExist ...
func (d *DefaultImageRegistry) ManifestExist(repository, reference string) (bool, string, error) {
	client, err := d.getClient(repository)
	if err != nil {
		return false, "", err
	}
	digest, exist, err := client.ManifestExist(reference)
	return exist, digest, err
}

// PullManifest ...
func (d *DefaultImageRegistry) PullManifest(repository, reference string, accepttedMediaTypes []string) (distribution.Manifest, string, error) {
	client, err := d.getClient(repository)
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
func (d *DefaultImageRegistry) PushManifest(repository, reference, mediaType string, payload []byte) error {
	client, err := d.getClient(repository)
	if err != nil {
		return err
	}
	_, err = client.PushManifest(reference, mediaType, payload)
	return err
}

// DeleteManifest ...
func (d *DefaultImageRegistry) DeleteManifest(repository, reference string) error {
	client, err := d.getClient(repository)
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
func (d *DefaultImageRegistry) BlobExist(repository, digest string) (bool, error) {
	client, err := d.getClient(repository)
	if err != nil {
		return false, err
	}
	return client.BlobExist(digest)
}

// PullBlob ...
func (d *DefaultImageRegistry) PullBlob(repository, digest string) (int64, io.ReadCloser, error) {
	client, err := d.getClient(repository)
	if err != nil {
		return 0, nil, err
	}
	return client.PullBlob(digest)
}

// PushBlob ...
func (d *DefaultImageRegistry) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	client, err := d.getClient(repository)
	if err != nil {
		return err
	}
	return client.PushBlob(digest, size, blob)
}

func isDigest(str string) bool {
	return strings.Contains(str, ":")
}

// ListTag ...
func (d *DefaultImageRegistry) ListTag(repository string) ([]string, error) {
	client, err := d.getClient(repository)
	if err != nil {
		return []string{}, err
	}
	return client.ListTag()
}
