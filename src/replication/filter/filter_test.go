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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFilterable struct {
	filterableType FilterableType
	resourceType   string
	name           string
	labels         []string
}

func (f *fakeFilterable) GetFilterableType() FilterableType {
	return f.filterableType
}

func (f *fakeFilterable) GetResourceType() string {
	return f.resourceType
}

func (f *fakeFilterable) GetName() string {
	return f.name
}
func (f *fakeFilterable) GetLabels() []string {
	return f.labels
}

func TestFilterOfResourceTypeFilter(t *testing.T) {
	filterable := &fakeFilterable{
		filterableType: FilterableTypeRepository,
		resourceType:   "image",
		name:           "library/hello-world",
	}

	filter := NewResourceTypeFilter("image")
	result, err := filter.Filter(filterable)
	require.Nil(t, nil, err)
	if assert.Equal(t, 1, len(result)) {
		assert.True(t, reflect.DeepEqual(filterable, result[0]))
	}

	filter = NewResourceTypeFilter("chart")
	result, err = filter.Filter(filterable)
	require.Nil(t, nil, err)
	assert.Equal(t, 0, len(result))
}

func TestApplyToOfResourceTypeFilter(t *testing.T) {
	filterable := &fakeFilterable{
		filterableType: FilterableTypeRepository,
	}

	filter := NewResourceTypeFilter("image")
	assert.True(t, filter.ApplyTo(filterable))

	filterable.filterableType = FilterableTypeVTag
	assert.True(t, filter.ApplyTo(filterable))

	filterable.filterableType = FilterableType("unknown")
	assert.False(t, filter.ApplyTo(filterable))
}

func TestFilterOfNameFilter(t *testing.T) {
	filterable := &fakeFilterable{
		name: "foo",
	}
	// pass the filter
	filter := &nameFilter{
		pattern: "*",
	}
	result, err := filter.Filter(filterable)
	require.Nil(t, err)
	if assert.Equal(t, 1, len(result)) {
		assert.True(t, reflect.DeepEqual(filterable, result[0].(*fakeFilterable)))
	}

	// cannot pass the filter
	filter.pattern = "cannotpass"
	result, err = filter.Filter(filterable)
	require.Nil(t, err)
	assert.Equal(t, 0, len(result))
}

func TestApplyToOfNameFilter(t *testing.T) {
	filterable := &fakeFilterable{
		filterableType: FilterableTypeRepository,
	}

	filter := &nameFilter{
		filterableType: FilterableTypeRepository,
	}
	assert.True(t, filter.ApplyTo(filterable))

	filterable.filterableType = FilterableTypeVTag
	assert.False(t, filter.ApplyTo(filterable))
}

func TestFilterOfLabelFilter(t *testing.T) {
	filterable := &fakeFilterable{
		labels: []string{"production"},
	}
	// pass the filter
	filter := &labelFilter{
		labels: []string{"production"},
	}
	result, err := filter.Filter(filterable)
	require.Nil(t, err)
	if assert.Equal(t, 1, len(result)) {
		assert.True(t, reflect.DeepEqual(filterable, result[0].(*fakeFilterable)))
	}
	// cannot pass the filter
	filter.labels = []string{"production", "ci-pass"}
	result, err = filter.Filter(filterable)
	require.Nil(t, err)
	assert.Equal(t, 0, len(result))
}

func TestApplyToOfLabelFilter(t *testing.T) {
	filterable := &fakeFilterable{
		filterableType: FilterableTypeRepository,
	}

	filter := labelFilter{}
	assert.False(t, filter.ApplyTo(filterable))

	filterable.filterableType = FilterableTypeVTag
	assert.True(t, filter.ApplyTo(filterable))
}

func TestDoFilter(t *testing.T) {
	tag1 := &fakeFilterable{
		filterableType: FilterableTypeVTag,
		name:           "1.0",
		labels:         []string{"production"},
	}
	tag2 := &fakeFilterable{
		filterableType: FilterableTypeVTag,
		name:           "latest",
		labels:         []string{"dev"},
	}
	filterables := []Filterable{tag1, tag2}
	filters := []Filter{
		NewVTagNameFilter("*"),
		NewVTagLabelFilter([]string{"production"}),
	}
	err := DoFilter(&filterables, filters...)
	require.Nil(t, err)
	if assert.Equal(t, 1, len(filterables)) {
		assert.True(t, reflect.DeepEqual(tag1, filterables[0]))
	}
}
