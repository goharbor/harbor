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

	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/replication/filter"
	"github.com/goharbor/harbor/src/replication/model"
)

// const definition
const (
	UserAgentReplication = "harbor-replication-service"
	MaxConcurrency       = 100
)

var registry = map[model.RegistryType]Factory{}

// Factory creates a specific Adapter according to the params
type Factory func(*model.Registry) (Adapter, error)

// Adapter interface defines the capabilities of registry
type Adapter interface {
	// Info return the information of this adapter
	Info() (*model.RegistryInfo, error)
	// PrepareForPush does the prepare work that needed for pushing/uploading the resources
	// eg: create the namespace or repository
	PrepareForPush([]*model.Resource) error
	// HealthCheck checks health status of registry
	HealthCheck() (model.HealthStatus, error)
}

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

// ChartRegistry defines the capabilities that a chart registry should have
type ChartRegistry interface {
	FetchCharts(filters []*model.Filter) ([]*model.Resource, error)
	ChartExist(name, version string) (bool, error)
	DownloadChart(name, version string) (io.ReadCloser, error)
	UploadChart(name, version string, chart io.Reader) error
	DeleteChart(name, version string) error
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

// RegisterFactory registers one adapter factory to the registry
func RegisterFactory(t model.RegistryType, factory Factory) error {
	if len(t) == 0 {
		return errors.New("invalid registry type")
	}
	if factory == nil {
		return errors.New("empty adapter factory")
	}

	if _, exist := registry[t]; exist {
		return fmt.Errorf("adapter factory for %s already exists", t)
	}
	registry[t] = factory
	return nil
}

// GetFactory gets the adapter factory by the specified name
func GetFactory(t model.RegistryType) (Factory, error) {
	factory, exist := registry[t]
	if !exist {
		return nil, fmt.Errorf("adapter factory for %s not found", t)
	}
	return factory, nil
}

// HasFactory checks whether there is given type adapter factory
func HasFactory(t model.RegistryType) bool {
	_, ok := registry[t]
	return ok
}

// ListRegisteredAdapterTypes lists the registered Adapter type
func ListRegisteredAdapterTypes() []model.RegistryType {
	types := []model.RegistryType{}
	for t := range registry {
		types = append(types, t)
	}
	return types
}
