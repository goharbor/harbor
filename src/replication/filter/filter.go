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
	"errors"
	"reflect"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/util"
)

// const definitions
const (
	FilterableTypeRepository = "repository"
	FilterableTypeVTag       = "vtag"
)

// FilterableType specifies the type of the filterable
type FilterableType string

// Filterable defines the methods that a filterable object must implement
type Filterable interface {
	// return what the type of the filterable object is(repository or vtag)
	GetFilterableType() FilterableType
	// return the resource type of the filterable object(image, chart, ...)
	GetResourceType() string
	GetName() string
	GetLabels() []string
}

// Filter defines the methods that a filter must implement
type Filter interface {
	// return whether the filter is applied to the specified Filterable
	ApplyTo(Filterable) bool
	Filter(...Filterable) ([]Filterable, error)
}

// NewResourceTypeFilter return a Filter to filter candidates according to the resource type
func NewResourceTypeFilter(resourceType string) Filter {
	return &resourceTypeFilter{
		resourceType: resourceType,
	}
}

// NewRepositoryNameFilter return a Filter to filter the repositories according to the name
func NewRepositoryNameFilter(pattern string) Filter {
	return &nameFilter{
		filterableType: FilterableTypeRepository,
		pattern:        pattern,
	}
}

// NewVTagNameFilter return a Filter to filter the vtags according to the name
func NewVTagNameFilter(pattern string) Filter {
	return &nameFilter{
		filterableType: FilterableTypeVTag,
		pattern:        pattern,
	}
}

// NewVTagLabelFilter return a Filter to filter vtags according to the label
func NewVTagLabelFilter(labels []string) Filter {
	return &labelFilter{
		labels: labels,
	}
}

type resourceTypeFilter struct {
	resourceType string
}

func (r *resourceTypeFilter) ApplyTo(filterable Filterable) bool {
	if filterable == nil {
		return false
	}
	switch filterable.GetFilterableType() {
	case FilterableTypeRepository, FilterableTypeVTag:
		return true
	default:
		return false
	}
}

func (r *resourceTypeFilter) Filter(filterables ...Filterable) ([]Filterable, error) {
	result := []Filterable{}
	for _, filterable := range filterables {
		if filterable.GetResourceType() == r.resourceType {
			result = append(result, filterable)
		}
	}
	return result, nil
}

type nameFilter struct {
	filterableType FilterableType
	pattern        string
}

func (n *nameFilter) ApplyTo(filterable Filterable) bool {
	if filterable == nil {
		return false
	}
	if filterable.GetFilterableType() == n.filterableType {
		return true
	}
	return false
}

func (n *nameFilter) Filter(filterables ...Filterable) ([]Filterable, error) {
	result := []Filterable{}
	for _, filterable := range filterables {
		name := filterable.GetName()
		match, err := util.Match(n.pattern, name)
		if err != nil {
			return nil, err
		}
		if match {
			log.Debugf("%q matches the pattern %q of name filter", name, n.pattern)
			result = append(result, filterable)
			continue
		}
		log.Debugf("%q doesn't match the pattern %q of name filter, skip", name, n.pattern)
	}
	return result, nil
}

type labelFilter struct {
	labels []string
}

func (l *labelFilter) ApplyTo(filterable Filterable) bool {
	if filterable == nil {
		return false
	}
	if filterable.GetFilterableType() == FilterableTypeVTag {
		return true
	}
	return false
}

func (l *labelFilter) Filter(filterables ...Filterable) ([]Filterable, error) {
	// if no specified label in the filter, just returns the input filterable
	// candidate as the result
	if len(l.labels) == 0 {
		return filterables, nil
	}
	result := []Filterable{}
	for _, filterable := range filterables {
		labels := map[string]struct{}{}
		for _, label := range filterable.GetLabels() {
			labels[label] = struct{}{}
		}
		match := true
		for _, label := range l.labels {
			if _, exist := labels[label]; !exist {
				match = false
				break
			}
		}
		// add the filterable to the result list if it contains
		// all labels defined for the filter
		if match {
			result = append(result, filterable)
		}
	}
	return result, nil
}

// DoFilter is a util function to help filter filterables easily.
// The parameter "filterables" must be a pointer points to a slice
// whose elements must be Filterable. After applying all the "filters"
// to the "filterables", the result is put back into the variable
// "filterables"
func DoFilter(filterables interface{}, filters ...Filter) error {
	if filterables == nil || len(filters) == 0 {
		return nil
	}

	value := reflect.ValueOf(filterables)
	// make sure the input is a pointer
	if value.Kind() != reflect.Ptr {
		return errors.New("the type of input should be pointer to a Filterable slice")
	}

	sliceValue := value.Elem()
	// make sure the input is a pointer points to a slice
	if sliceValue.Type().Kind() != reflect.Slice {
		return errors.New("the type of input should be pointer to a Filterable slice")
	}

	filterableType := reflect.TypeOf((*Filterable)(nil)).Elem()
	elemType := sliceValue.Type().Elem()
	// make sure the input is a pointer points to a Filterable slice
	if !elemType.Implements(filterableType) {
		return errors.New("the type of input should be pointer to a Filterable slice")
	}

	// convert the input to Filterable slice
	items := []Filterable{}
	for i := 0; i < sliceValue.Len(); i++ {
		items = append(items, sliceValue.Index(i).Interface().(Filterable))
	}

	// do filter
	var err error
	items, err = doFilter(items, filters...)
	if err != nil {
		return err
	}

	// convert back to the origin type
	result := reflect.MakeSlice(reflect.SliceOf(elemType), 0, len(items))
	for _, item := range items {
		result = reflect.Append(result, reflect.ValueOf(item))
	}
	value.Elem().Set(result)

	return nil
}

func doFilter(filterables []Filterable, filters ...Filter) ([]Filterable, error) {
	var appliedTo, notAppliedTo []Filterable
	var err error
	for _, filter := range filters {
		appliedTo, notAppliedTo = nil, nil
		for _, filterable := range filterables {
			if filter.ApplyTo(filterable) {
				appliedTo = append(appliedTo, filterable)
			} else {
				notAppliedTo = append(notAppliedTo, filterable)
			}
		}
		filterables, err = filter.Filter(appliedTo...)
		if err != nil {
			return nil, err
		}
		filterables = append(filterables, notAppliedTo...)
	}
	return filterables, nil
}
