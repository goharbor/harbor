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
	"io"

	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

// ImageRegistry defines the capabilities that an image registry should have
type ImageRegistry interface {
	FetchImages(namespaces []string, filters []*model.Filter) ([]*model.Resource, error)
	ManifestExist(repository, reference string) (exist bool, digest string, err error)
	PullManifest(repository, reference string, accepttedMediaTypes []string) (manifest distribution.Manifest, digest string, err error)
	PushManifest(repository, reference, mediaType string, payload []byte) error
	BlobExist(repository, digest string) (exist bool, err error)
	PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error)
	PushBlob(repository, digest string, size int64, blob io.Reader) error
}

// TODO implement the functions

// DefaultImageRegistry provides a default implementation for interface ImageRegistry
type DefaultImageRegistry struct{}

// FetchImages ...
func (d *DefaultImageRegistry) FetchImages(namespaces []string, filters []*model.Filter) ([]*model.Resource, error) {
	return nil, nil
}

// ManifestExist ...
func (d *DefaultImageRegistry) ManifestExist(repository, reference string) (exist bool, digest string, err error) {
	return false, "", nil
}

// PullManifest ...
func (d *DefaultImageRegistry) PullManifest(repository, reference string, accepttedMediaTypes []string) (manifest distribution.Manifest, digest string, err error) {
	return nil, "", nil
}

// PushManifest ...
func (d *DefaultImageRegistry) PushManifest(repository, reference, mediaType string, payload []byte) error {
	return nil
}

// BlobExist ...
func (d *DefaultImageRegistry) BlobExist(repository, digest string) (exist bool, err error) {
	return false, nil
}

// PullBlob ...
func (d *DefaultImageRegistry) PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error) {
	return 0, nil, nil
}

// PushBlob ...
func (d *DefaultImageRegistry) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	return nil
}
