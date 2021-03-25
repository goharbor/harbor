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

package filter

import "github.com/goharbor/harbor/src/pkg/reg/model"

// DoFilterResources filter resources according to the filters
func DoFilterResources(resources []*model.Resource, filters []*model.Filter) ([]*model.Resource, error) {
	repoFilters, err := BuildRepositoryFilters(filters)
	if err != nil {
		return nil, err
	}
	artFilters, err := BuildArtifactFilters(filters)
	if err != nil {
		return nil, err
	}

	var result []*model.Resource
	for _, resource := range resources {
		repositories, err := repoFilters.Filter([]*model.Repository{resource.Metadata.Repository})
		if err != nil {
			return nil, err
		}
		if len(repositories) == 0 {
			continue
		}
		artifacts, err := artFilters.Filter(resource.Metadata.Artifacts)
		if err != nil {
			return nil, err
		}
		if len(artifacts) == 0 {
			continue
		}
		result = append(result, &model.Resource{
			Type: resource.Type,
			Metadata: &model.ResourceMetadata{
				Repository: repositories[0],
				Artifacts:  artifacts,
			},
			Registry:     resource.Registry,
			ExtendedInfo: resource.ExtendedInfo,
			Deleted:      resource.Deleted,
			IsDeleteTag:  resource.IsDeleteTag,
			Override:     resource.Override,
		})
	}

	// remove this after we deprecate chart museum
	return filterByResourceType(result, filters)
}

// After we deprecated chart museum, the resource types model.ResourceTypeArtifact and model.ResourceTypeChart
// are useless, this function should be removed as well
func filterByResourceType(resources []*model.Resource, filters []*model.Filter) ([]*model.Resource, error) {
	var resourceType string
	for _, filter := range filters {
		if filter.Type == model.FilterTypeResource {
			// model.ResourceTypeImage is handled by artifact filters in function "DoFilterResources"
			if filter.Value.(string) == model.ResourceTypeArtifact || filter.Value.(string) == model.ResourceTypeChart {
				resourceType = filter.Value.(string)
			}
			break
		}
	}
	// no resource type, return the candidates directly
	if len(resourceType) == 0 {
		return resources, nil
	}

	var result []*model.Resource
	for _, resource := range resources {
		if resource.Type == resourceType {
			result = append(result, resource)
		}
	}
	return result, nil
}
