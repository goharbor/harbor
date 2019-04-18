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

	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

// TODO unify the filter logic from different adapters into one?
// and move the code into a separated common package

// Filterable defines the interface that an object should implement
// if the object can be filtered
type Filterable interface {
	Match([]*model.Filter) (bool, error)
}

// FilterItem is a filterable object that can be used to match string pattern
type FilterItem struct {
	Value string
}

// Match ...
func (f *FilterItem) Match(filters []*model.Filter) (bool, error) {
	if len(filters) == 0 {
		return true, nil
	}
	matched := true
	for _, filter := range filters {
		pattern, ok := filter.Value.(string)
		if !ok {
			return false, fmt.Errorf("the type of filter value isn't string: %v", filter)
		}
		m, err := util.Match(pattern, f.Value)
		if err != nil {
			return false, err
		}
		if !m {
			matched = false
			break
		}
	}
	return matched, nil
}
