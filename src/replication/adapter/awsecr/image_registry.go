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

package awsecr

import (
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

var _ adp.ImageRegistry = adapter{}

func (n adapter) FetchImages(filters []*model.Filter) ([]*model.Resource, error) {
	nameFilterPattern := ""
	tagFilterPattern := ""
	for _, filter := range filters {
		switch filter.Type {
		case model.FilterTypeName:
			nameFilterPattern = filter.Value.(string)
		case model.FilterTypeTag:
			tagFilterPattern = filter.Value.(string)
		}
	}
	repositories, err := n.filterRepositories(nameFilterPattern)
	if err != nil {
		return nil, err
	}

	var resources []*model.Resource
	for _, repository := range repositories {
		tags, err := n.filterTags(repository, tagFilterPattern)
		if err != nil {
			return nil, err
		}
		if len(tags) == 0 {
			continue
		}
		resources = append(resources, &model.Resource{
			Type:     model.ResourceTypeImage,
			Registry: n.registry,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: repository,
				},
				Vtags: tags,
			},
		})
	}

	return resources, nil
}

func (n adapter) filterRepositories(pattern string) ([]string, error) {
	// if the pattern is a specific repository name, just returns the parsed repositories
	// and will check the existence later when filtering the tags
	if repositories, ok := util.IsSpecificPath(pattern); ok {
		return repositories, nil
	}
	// search repositories from catalog api
	repositories, err := n.Catalog()
	if err != nil {
		return nil, err
	}
	// if the pattern is null, just return the result of catalog API
	if len(pattern) == 0 {
		return repositories, nil
	}
	result := []string{}
	for _, repository := range repositories {
		match, err := util.Match(pattern, repository)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, repository)
		}
	}
	return result, nil
}

func (n adapter) filterTags(repository, pattern string) ([]string, error) {
	tags, err := n.ListTag(repository)
	if err != nil {
		return nil, err
	}
	if len(pattern) == 0 {
		return tags, nil
	}

	var result []string
	for _, tag := range tags {
		match, err := util.Match(pattern, tag)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, tag)
		}
	}
	return result, nil
}
