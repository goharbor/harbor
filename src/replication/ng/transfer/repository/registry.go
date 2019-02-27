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

package repository

import (
	"io"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/common/http/modifier"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	pkg_registry "github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

// const definition
const (
	// TODO: add filter for the agent in registry webhook handler
	UserAgentReplicator = "harbor-replicator"
)

// Registry defines an the interface for registry service
type Registry interface {
	ManifestExist(repository, reference string) (exist bool, digest string, err error)
	PullManifest(repository, reference string, accepttedMediaTypes []string) (manifest distribution.Manifest, digest string, err error)
	PushManifest(repository, reference, mediaType string, payload []byte) error
	BlobExist(repository, digest string) (exist bool, err error)
	PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error)
	PushBlob(repository, digest string, size int64, blob io.Reader) error
}

// NewRegistry returns an instance of the default registry implementation
// TODO: passing the tokenServiceURL
func NewRegistry(reg *model.Registry, repository string,
	tokenServiceURL ...string) (Registry, error) {
	// use the same HTTP connection pool for all clients
	transport := pkg_registry.GetHTTPTransport(reg.Insecure)
	modifiers := []modifier.Modifier{
		&auth.UserAgentModifier{
			UserAgent: UserAgentReplicator,
		},
	}
	if reg.Credential != nil {
		cred := auth.NewBasicAuthCredential(
			reg.Credential.AccessKey,
			reg.Credential.AccessSecret)
		authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
			Transport: transport,
		}, cred, tokenServiceURL...)

		modifiers = append(modifiers, authorizer)
	}

	client, err := pkg_registry.NewRepository(repository, reg.URL,
		&http.Client{
			Transport: pkg_registry.NewTransport(transport, modifiers...),
		})
	if err != nil {
		return nil, err
	}

	return &registry{
		client: client,
	}, nil
}

type registry struct {
	client *pkg_registry.Repository
}

func (r *registry) ManifestExist(repository, reference string) (bool, string, error) {
	digest, exist, err := r.client.ManifestExist(reference)
	return exist, digest, err
}
func (r *registry) PullManifest(repository, reference string, accepttedMediaTypes []string) (distribution.Manifest, string, error) {
	digest, mediaType, payload, err := r.client.PullManifest(reference, accepttedMediaTypes)
	if err != nil {
		return nil, "", err
	}
	if strings.Contains(mediaType, "application/json") {
		mediaType = schema1.MediaTypeManifest
	}
	manifest, _, err := pkg_registry.UnMarshal(mediaType, payload)
	if err != nil {
		return nil, "", err
	}
	return manifest, digest, nil
}
func (r *registry) PushManifest(repository, reference, mediaType string, payload []byte) error {
	_, err := r.client.PushManifest(reference, mediaType, payload)
	return err
}
func (r *registry) BlobExist(repository, digest string) (bool, error) {
	return r.client.BlobExist(digest)
}
func (r *registry) PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error) {
	return r.client.PullBlob(digest)
}
func (r *registry) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	return r.client.PushBlob(digest, size, blob)
}
