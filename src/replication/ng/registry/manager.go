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

package registry

import (
	"github.com/goharbor/harbor/src/replication/ng/model"
)

// Manager manages registries
type Manager interface {
	// Add new registry
	Add(*model.Registry) (int64, error)
	// List registries, returns total count, registry list and error
	List(...*model.RegistryQuery) (int64, []*model.Registry, error)
	// Get the specified registry
	Get(int64) (*model.Registry, error)
	// Update the registry, the "props" are the properties of registry
	// that need to be updated
	Update(registry *model.Registry, props ...string) error
	// Remove the registry with the specified ID
	Remove(int64) error
}
