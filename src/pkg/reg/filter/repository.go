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

import (
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
)

// DoFilterRepositories filter repositories according to the filters
func DoFilterRepositories(repositories []*model.Repository, filters []*model.Filter) ([]*model.Repository, error) {
	fl, err := BuildRepositoryFilters(filters)
	if err != nil {
		return nil, err
	}
	return fl.Filter(repositories)
}

// BuildRepositoryFilters from the defined filters
func BuildRepositoryFilters(filters []*model.Filter) (RepositoryFilters, error) {
	var fs RepositoryFilters
	for _, filter := range filters {
		var f RepositoryFilter
		switch filter.Type {
		case model.FilterTypeName:
			f = &repositoryNameFilter{
				pattern: filter.Value.(string),
			}
		}
		if f != nil {
			fs = append(fs, f)
		}
	}
	return fs, nil
}

// RepositoryFilter filter repositoreis
type RepositoryFilter interface {
	Filter([]*model.Repository) ([]*model.Repository, error)
}

// RepositoryFilters is an array of repository filters
type RepositoryFilters []RepositoryFilter

// Filter repositories
func (r RepositoryFilters) Filter(repositories []*model.Repository) ([]*model.Repository, error) {
	var err error
	for _, filter := range r {
		repositories, err = filter.Filter(repositories)
		if err != nil {
			return nil, err
		}
	}
	return repositories, nil
}

type repositoryNameFilter struct {
	pattern string
}

func (r *repositoryNameFilter) Filter(repositories []*model.Repository) ([]*model.Repository, error) {
	if len(r.pattern) == 0 {
		return repositories, nil
	}
	var result []*model.Repository
	for _, repository := range repositories {
		match, err := util.Match(r.pattern, repository.Name)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, repository)
			continue
		}
	}
	return result, nil
}
