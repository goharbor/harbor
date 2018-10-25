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
	"fmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/models"
)

// LabelFilter filter resources according to label
type LabelFilter struct {
	labelID int64
}

// Init ...
func (l *LabelFilter) Init() error {
	return nil
}

// GetConverter ...
func (l *LabelFilter) GetConverter() Converter {
	return nil
}

// NewLabelFilter returns an instance of LabelFilter
func NewLabelFilter(labelID int64) *LabelFilter {
	return &LabelFilter{
		labelID: labelID,
	}
}

// DoFilter filter the resources according to the label
func (l *LabelFilter) DoFilter(items []models.FilterItem) []models.FilterItem {
	candidates := []string{}
	for _, item := range items {
		candidates = append(candidates, item.Value)
	}
	log.Debugf("label filter candidates: %v", candidates)
	result := []models.FilterItem{}
	for _, item := range items {
		hasLabel, err := hasLabel(item, l.labelID)
		if err != nil {
			log.Errorf("failed to check the label of resouce %v: %v, skip it", item, err)
			continue
		}
		if hasLabel {
			log.Debugf("has label %d, add %s to the label filter result list", l.labelID, item.Value)
			result = append(result, item)
		}
	}
	return result
}

func hasLabel(resource models.FilterItem, labelID int64) (bool, error) {
	rType := ""
	switch resource.Kind {
	case replication.FilterItemKindProject:
		rType = common.ResourceTypeProject
	case replication.FilterItemKindRepository:
		rType = common.ResourceTypeRepository
	case replication.FilterItemKindTag:
		rType = common.ResourceTypeImage
	default:
		return false, fmt.Errorf("invalid resource type: %s", resource.Kind)
	}
	rl, err := dao.GetResourceLabel(rType, resource.Value, labelID)
	if err != nil {
		return false, err
	}
	return rl != nil, nil
}
