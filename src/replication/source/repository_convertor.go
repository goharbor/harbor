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

package source

import (
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/models"
	"github.com/goharbor/harbor/src/replication/registry"
)

// RepositoryConverter implement Converter interface, convert projects to repositories
type RepositoryConverter struct {
	registry registry.Adaptor
}

// NewRepositoryConverter returns an instance of RepositoryConverter
func NewRepositoryConverter(registry registry.Adaptor) *RepositoryConverter {
	return &RepositoryConverter{
		registry: registry,
	}
}

// Convert projects to repositories
func (r *RepositoryConverter) Convert(items []models.FilterItem) []models.FilterItem {
	result := []models.FilterItem{}
	for _, item := range items {
		// just put it to the result list if the item is not a project
		if item.Kind != replication.FilterItemKindProject {
			result = append(result, item)
			continue
		}

		repositories := r.registry.GetRepositories(item.Value)
		for _, repository := range repositories {
			result = append(result, models.FilterItem{
				Kind:      replication.FilterItemKindRepository,
				Value:     repository.Name,
				Operation: item.Operation,
			})
		}
	}
	return result
}
