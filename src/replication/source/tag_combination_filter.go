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
	"strings"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/models"
)

// TagCombinationFilter implements Filter interface for merging tag filter items
// whose repository are same into one repository filter item
type TagCombinationFilter struct{}

// NewTagCombinationFilter returns an instance of TagCombinationFilter
func NewTagCombinationFilter() *TagCombinationFilter {
	return &TagCombinationFilter{}
}

// Init the filter. nil implement for now
func (t *TagCombinationFilter) Init() error {
	return nil
}

// GetConverter returns the converter
func (t *TagCombinationFilter) GetConverter() Converter {
	return nil
}

// DoFilter filters resources
func (t *TagCombinationFilter) DoFilter(filterItems []models.FilterItem) []models.FilterItem {
	repos := map[string][]string{}
	for _, item := range filterItems {
		if item.Kind != replication.FilterItemKindTag {
			log.Warningf("unexpected filter item kind, expected: %s, got: %s, skip",
				replication.FilterItemKindTag, item.Kind)
			continue
		}

		strs := strings.Split(item.Value, ":")
		if len(strs) != 2 {
			log.Warningf("unexpected image format: %s, skip", item.Value)
			continue
		}

		repos[strs[0]] = append(repos[strs[0]], strs[1])
	}

	// TODO append operation
	items := []models.FilterItem{}
	for repo, tags := range repos {
		items = append(items, models.FilterItem{
			Kind:  replication.FilterItemKindRepository,
			Value: repo,
			Metadata: map[string]interface{}{
				"tags": tags,
			},
		})
	}

	return items
}
