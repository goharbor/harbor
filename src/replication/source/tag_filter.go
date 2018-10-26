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
	"github.com/goharbor/harbor/src/replication/registry"
)

// TagFilter implements Filter interface to filter tag
type TagFilter struct {
	pattern   string
	converter Converter
}

// NewTagFilter returns an instance of TagFilter
func NewTagFilter(pattern string, registry registry.Adaptor) *TagFilter {
	return &TagFilter{
		pattern:   pattern,
		converter: NewTagConverter(registry),
	}
}

// Init ...
func (t *TagFilter) Init() error {
	return nil
}

// GetConverter ...
func (t *TagFilter) GetConverter() Converter {
	return t.converter
}

// DoFilter filters tag of the image
func (t *TagFilter) DoFilter(items []models.FilterItem) []models.FilterItem {
	candidates := []string{}
	for _, item := range items {
		candidates = append(candidates, item.Value)
	}
	log.Debugf("tag filter candidates: %v", candidates)

	result := []models.FilterItem{}
	for _, item := range items {
		if item.Kind != replication.FilterItemKindTag {
			log.Warningf("unsupported type %s for tag filter, dropped", item.Kind)
			continue
		}

		if len(t.pattern) == 0 {
			log.Debugf("pattern is null, add %s to the tag filter result list", item.Value)
			result = append(result, item)
			continue
		}

		tag := strings.SplitN(item.Value, ":", 2)[1]
		matched, err := match(t.pattern, tag)
		if err != nil {
			log.Errorf("failed to match pattern %s to value %s: %v, skip it", t.pattern, tag, err)
			continue
		}

		if matched {
			log.Debugf("pattern %s matched, add %s to the tag filter result list", t.pattern, item.Value)
			result = append(result, item)
		}
	}
	return result
}
