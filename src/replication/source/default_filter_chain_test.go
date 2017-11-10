// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vmware/harbor/src/replication"
	"github.com/vmware/harbor/src/replication/models"
)

func TestBuild(t *testing.T) {
	chain := NewDefaultFilterChain(nil)
	require.Nil(t, chain.Build(nil))
}

func TestFilters(t *testing.T) {
	filters := []Filter{NewPatternFilter("project", "*")}
	chain := NewDefaultFilterChain(filters)
	assert.EqualValues(t, filters, chain.Filters())
}

func TestDoFilter(t *testing.T) {
	projectFilter := NewPatternFilter(replication.FilterItemKindProject, "library*")
	repositoryFilter := NewPatternFilter(replication.FilterItemKindRepository,
		"library/ubuntu*", &fakeRepositoryConvertor{})
	filters := []Filter{projectFilter, repositoryFilter}

	items := []models.FilterItem{
		models.FilterItem{
			Kind:  replication.FilterItemKindProject,
			Value: "library",
		},
		models.FilterItem{
			Kind:  replication.FilterItemKindProject,
			Value: "test",
		},
	}
	chain := NewDefaultFilterChain(filters)
	items = chain.DoFilter(items)
	assert.EqualValues(t, []models.FilterItem{
		models.FilterItem{
			Kind:  replication.FilterItemKindRepository,
			Value: "library/ubuntu",
		},
	}, items)

}

type fakeRepositoryConvertor struct{}

func (f *fakeRepositoryConvertor) Convert(items []models.FilterItem) []models.FilterItem {
	result := []models.FilterItem{}
	for _, item := range items {
		result = append(result, models.FilterItem{
			Kind:  replication.FilterItemKindRepository,
			Value: item.Value + "/ubuntu",
		})
	}
	return result
}
