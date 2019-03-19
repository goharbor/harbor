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

// As the Info isn't a valid map key, so we use the slice
// as the adapter registry
var registry = []*item{}

// const definition
const (
	FilterStyleText  = "input"
	FilterStyleRadio = "radio"
)

// FilterStyle is used for UI to determine how to render the filter
type FilterStyle string

type item struct {
	info    *Info
	factory Factory
}

// Filter ...
type Filter struct {
	Type   model.FilterType `json:"type"`
	Style  FilterStyle      `json:"style"`
	Values []string         `json:"values,omitempty"`
}

// Info provides base info and capability declarations of the adapter
type Info struct {
	Type                     model.RegistryType   `json:"type"`
	Description              string               `json:"description"`
	SupportedResourceTypes   []model.ResourceType `json:"-"`
	SupportedResourceFilters []*Filter            `json:"supported_resource_filters"`
	SupportedTriggers        []model.TriggerType  `json:"supported_triggers"`
}

// Factory creates a specific Adapter according to the params
type Factory func(*model.Registry) (Adapter, error)

// Adapter interface defines the capabilities of registry
type Adapter interface {
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
}

// RegisterFactory registers one adapter factory to the registry
func RegisterFactory(info *Info, factory Factory) error {
	if len(info.Type) == 0 {
		return errors.New("invalid registry type")
	}
	if len(info.SupportedResourceTypes) == 0 {
		return errors.New("must support at least one resource type")
	}
	if len(info.SupportedTriggers) == 0 {
		return errors.New("must support at least one trigger")
	}
	if factory == nil {
		return errors.New("empty adapter factory")
	}
	for _, item := range registry {
		if item.info.Type == info.Type {
			return fmt.Errorf("adapter factory for %s already exists", info.Type)
		}
	}
	registry = append(registry, &item{
		info:    info,
		factory: factory,
	})
	return nil
}

// GetFactory gets the adapter factory by the specified name
func GetFactory(t model.RegistryType) (Factory, error) {
	for _, item := range registry {
		if item.info.Type == t {
			return item.factory, nil
		}
	}
	return nil, fmt.Errorf("adapter factory for %s not found", t)
}

// ListAdapterInfos lists the info of registered Adapters
func ListAdapterInfos() []*Info {
	infos := []*Info{}
	for _, item := range registry {
		infos = append(infos, item.info)
	}
	return infos
}

// GetAdapterInfo returns the info of a specified registry type
func GetAdapterInfo(t model.RegistryType) *Info {
	for _, item := range registry {
		if item.info.Type == t {
			return item.info
		}
	}
	return nil
}
