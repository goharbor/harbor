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

	"github.com/goharbor/harbor/src/replication/ng/model"
)

var registry = map[model.RegistryType]Factory{}

// Factory creates a specific Adapter according to the params
type Factory func(*model.Registry) (Adapter, error)

// Info provides base info and capability declarations of the adapter
type Info struct {
	Name                     model.RegistryType   `json:"name"`
	Description              string               `json:"description"`
	SupportedResourceTypes   []model.ResourceType `json:"supported_resource_types"`
	SupportedResourceFilters []model.FilterType   `json:"supported_resource_filters"`
}

// Adapter interface defines the capabilities of registry
type Adapter interface {
	// Info return the information of this adapter
	Info() *Info
	// Lists the available namespaces under the specified registry with the
	// provided credential/token
	ListNamespaces(*model.NamespaceQuery) ([]*model.Namespace, error)
	// Create a new namespace
	// This method should guarantee it's idempotent
	// And returns nil if a namespace with the same name already exists
	CreateNamespace(*model.Namespace) error
	// Get the namespace specified by the name, the returning value should
	// contain the metadata about the namespace if it has
	GetNamespace(string) (*model.Namespace, error)
	// Fetch the content resource under the namespace by filters
	// SUGGESTION: Adapter provider can do their own filter based on the filter pattern
	// or call the default `DoFilter` function of the filter to complete resource filtering.
	FetchResources(namespace []string, filters []*model.Filter) ([]*model.Resource, error)
}

// RegisterFactory registers one adapter factory to the registry
func RegisterFactory(name model.RegistryType, factory Factory) error {
	if !name.Valid() {
		return errors.New("invalid adapter factory name")
	}
	if factory == nil {
		return errors.New("empty adapter factory")
	}
	if _, exist := registry[name]; exist {
		return fmt.Errorf("adapter factory for %s already exists", name)
	}
	registry[name] = factory
	return nil
}

// GetFactory gets the adapter factory by the specified name
func GetFactory(name model.RegistryType) (Factory, error) {
	factory, exist := registry[name]
	if !exist {
		return nil, fmt.Errorf("adapter factory for %s not found", name)
	}

	return factory, nil
}
