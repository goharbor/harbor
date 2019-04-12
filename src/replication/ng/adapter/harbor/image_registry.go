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

package harbor

import (
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

type repository struct {
	Name string `json:"name"`
}

func (r *repository) Match(filters []*model.Filter) (bool, error) {
	supportedFilters := []*model.Filter{}
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			supportedFilters = append(supportedFilters, filter)
		}
	}
	// trim the project part
	_, name := utils.ParseRepository(r.Name)
	item := &FilterItem{
		Value: name,
	}
	return item.Match(supportedFilters)
}

type tag struct {
	Name string `json:"name"`
}

func (t *tag) Match(filters []*model.Filter) (bool, error) {
	supportedFilters := []*model.Filter{}
	for _, filter := range filters {
		if filter.Type == model.FilterTypeTag {
			supportedFilters = append(supportedFilters, filter)
		}
	}
	item := &FilterItem{
		Value: t.Name,
	}
	return item.Match(supportedFilters)
}

func (a *adapter) FetchImages(namespaces []string, filters []*model.Filter) ([]*model.Resource, error) {
	resources := []*model.Resource{}
	for _, namespace := range namespaces {
		project, err := a.getProject(namespace)
		if err != nil {
			return nil, err
		}
		repositories := []*repository{}
		url := fmt.Sprintf("%s/api/repositories?project_id=%d", a.coreServiceURL, project.ID)
		if err = a.client.Get(url, &repositories); err != nil {
			return nil, err
		}
		repositories, err = filterRepositories(repositories, filters)
		if err != nil {
			return nil, err
		}
		for _, repository := range repositories {
			url := fmt.Sprintf("%s/api/repositories/%s/tags", a.coreServiceURL, repository.Name)
			tags := []*tag{}
			if err = a.client.Get(url, &tags); err != nil {
				return nil, err
			}
			tags, err = filterTags(tags, filters)
			if err != nil {
				return nil, err
			}
			if len(tags) == 0 {
				continue
			}
			vtags := []string{}
			for _, tag := range tags {
				vtags = append(vtags, tag.Name)
			}
			resources = append(resources, &model.Resource{
				Type:     model.ResourceTypeRepository,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Namespace: &model.Namespace{
						Name: namespace,
						// TODO filling the metadata
					},
					Repository: &model.Repository{
						Name: strings.TrimPrefix(repository.Name, namespace+"/"),
					},
					Vtags: vtags,
				},
			})
		}
	}

	return resources, nil
}

// override the default implementation from the default image registry
// by calling Harbor API directly
func (a *adapter) DeleteManifest(repository, reference string) error {
	url := fmt.Sprintf("%s/api/repositories/%s/tags/%s", a.coreServiceURL, repository, reference)
	return a.client.Delete(url)
}

func filterRepositories(repositories []*repository, filters []*model.Filter) ([]*repository, error) {
	result := []*repository{}
	for _, repository := range repositories {
		match, err := repository.Match(filters)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, repository)
		}
	}
	return result, nil
}

func filterTags(tags []*tag, filters []*model.Filter) ([]*tag, error) {
	result := []*tag{}
	for _, tag := range tags {
		match, err := tag.Match(filters)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, tag)
		}
	}
	return result, nil
}
