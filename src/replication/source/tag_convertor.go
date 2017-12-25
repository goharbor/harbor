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
	"github.com/vmware/harbor/src/replication"
	"github.com/vmware/harbor/src/replication/models"
	"github.com/vmware/harbor/src/replication/registry"
)

// TagConvertor implement Convertor interface, convert repositories to tags
type TagConvertor struct {
	registry registry.Adaptor
}

// NewTagConvertor returns an instance of TagConvertor
func NewTagConvertor(registry registry.Adaptor) *TagConvertor {
	return &TagConvertor{
		registry: registry,
	}
}

//Convert repositories to tags
func (t *TagConvertor) Convert(items []models.FilterItem) []models.FilterItem {
	result := []models.FilterItem{}
	for _, item := range items {
		if item.Kind != replication.FilterItemKindRepository {
			// just put it to the result list if the item is not a repository
			result = append(result, item)
			continue
		}

		tags := t.registry.GetTags(item.Value, "")
		for _, tag := range tags {
			result = append(result, models.FilterItem{
				Kind:      replication.FilterItemKindTag,
				Value:     item.Value + ":" + tag.Name,
				Operation: item.Operation,
			})
		}
	}
	return result
}
