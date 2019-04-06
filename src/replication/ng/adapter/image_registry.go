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
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/goharbor/harbor/src/common/http/modifier"
	common_http_auth "github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/utils/log"
	registry_pkg "github.com/goharbor/harbor/src/common/utils/registry"
	util_registry "github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/replication/ng/model"
	"github.com/goharbor/harbor/src/replication/ng/util"
)

// const definition
const (
	UserAgentReplication = "harbor-replication-service"
)

// ImageRegistry defines the capabilities that an image registry should have
type ImageRegistry interface {
	FetchImages(namespaces []string, filters []*model.Filter) ([]*model.Resource, error)
	ManifestExist(repository, reference string) (exist bool, digest string, err error)
	PullManifest(repository, reference string, accepttedMediaTypes []string) (manifest distribution.Manifest, digest string, err error)
	PushManifest(repository, reference, mediaType string, payload []byte) error
	// the "reference" can be "tag" or "digest", the function needs to handle both
	DeleteManifest(repository, reference string) error
	BlobExist(repository, digest string) (exist bool, err error)
	PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error)
	PushBlob(repository, digest string, size int64, blob io.Reader) error
}

// DefaultImageRegistry provides a default implementation for interface ImageRegistry
type DefaultImageRegistry struct {
	sync.RWMutex
	registry *model.Registry
	client   *http.Client
	url      string
	clients  map[string]*registry_pkg.Repository
}

// NewDefaultImageRegistry returns an instance of DefaultImageRegistry
func NewDefaultImageRegistry(registry *model.Registry) *DefaultImageRegistry {
	transport := util.GetHTTPTransport(registry.Insecure)
	modifiers := []modifier.Modifier{
		&auth.UserAgentModifier{
			UserAgent: UserAgentReplication,
		},
	}
	if registry.Credential != nil {
		var cred modifier.Modifier
		if registry.Credential.Type == model.CredentialTypeSecret {
			cred = common_http_auth.NewSecretAuthorizer(registry.Credential.AccessSecret)
		} else {
			cred = auth.NewBasicAuthCredential(
				registry.Credential.AccessKey,
				registry.Credential.AccessSecret)
		}
		tokenServiceURL := ""
		// the registry is a local Harbor instance if the core URL is specified,
		// use the internal token service URL instead
		if len(registry.CoreURL) > 0 {
			tokenServiceURL = fmt.Sprintf("%s/service/token", registry.CoreURL)
		}
		authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
			Transport: transport,
		}, cred, tokenServiceURL)

		modifiers = append(modifiers, authorizer)
	}
	client := &http.Client{
		Transport: registry_pkg.NewTransport(transport, modifiers...),
	}
	return &DefaultImageRegistry{
		client:   client,
		registry: registry,
		clients:  map[string]*registry_pkg.Repository{},
		url:      registry.URL,
	}
}

func (d *DefaultImageRegistry) getClient(repository string) (*registry_pkg.Repository, error) {
	client := d.get(repository)
	if client != nil {
		return client, nil
	}

	return d.create(repository)
}

func (d *DefaultImageRegistry) get(repository string) *registry_pkg.Repository {
	d.RLock()
	defer d.RUnlock()
	client, exist := d.clients[repository]
	if exist {
		return client
	}
	return nil
}

func (d *DefaultImageRegistry) create(repository string) (*registry_pkg.Repository, error) {
	d.Lock()
	defer d.Unlock()
	// double check
	client, exist := d.clients[repository]
	if exist {
		return client, nil
	}

	client, err := registry_pkg.NewRepository(repository, d.url, d.client)
	if err != nil {
		return nil, err
	}
	d.clients[repository] = client
	return client, nil
}

// HealthCheck checks health status of a registry
func (d *DefaultImageRegistry) HealthCheck() (model.HealthStatus, error) {
	if d.registry.Credential == nil || (len(d.registry.Credential.AccessKey) == 0 && len(d.registry.Credential.AccessSecret) == 0) {
		return d.pingAnonymously()
	}

	// TODO(ChenDe): Support other credential type like OAuth, for the moment, only basic auth is supported.
	if d.registry.Credential.Type != model.CredentialTypeBasic {
		return model.Unknown, fmt.Errorf("unknown credential type '%s', only '%s' supported yet", d.registry.Credential.Type, model.CredentialTypeBasic)
	}

	transport := util.GetHTTPTransport(d.registry.Insecure)
	credential := auth.NewBasicAuthCredential(d.registry.Credential.AccessKey, d.registry.Credential.AccessSecret)
	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: transport,
	}, credential)
	registry, err := util_registry.NewRegistry(d.registry.URL, &http.Client{
		Transport: util_registry.NewTransport(transport, authorizer),
	})
	if err != nil {
		return model.Unknown, err
	}

	err = registry.Ping()
	if err != nil {
		return model.Unhealthy, err
	}
	return model.Healthy, nil
}

func (d *DefaultImageRegistry) pingAnonymously() (model.HealthStatus, error) {
	registry, err := util_registry.NewRegistry(d.registry.URL, &http.Client{
		Transport: util_registry.NewTransport(util.GetHTTPTransport(d.registry.Insecure)),
	})
	if err != nil {
		return model.Unknown, err
	}

	err = registry.PingSimple()
	if err != nil {
		return model.Unhealthy, err
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
