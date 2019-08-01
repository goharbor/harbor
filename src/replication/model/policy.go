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

package model

import (
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/replication/filter"

	"github.com/astaxie/beego/validation"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/robfig/cron"
)

// const definition
const (
	FilterTypeResource FilterType = "resource"
	FilterTypeName     FilterType = "name"
	FilterTypeTag      FilterType = "tag"
	FilterTypeLabel    FilterType = "label"

	TriggerTypeManual     TriggerType = "manual"
	TriggerTypeScheduled  TriggerType = "scheduled"
	TriggerTypeEventBased TriggerType = "event_based"
)

// Policy defines the structure of a replication policy
type Policy struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Creator     string `json:"creator"`
	// source
	SrcRegistry *Registry `json:"src_registry"`
	// destination
	DestRegistry *Registry `json:"dest_registry"`
	// Only support two dest namespace modes:
	// Put all the src resources to the one single dest namespace
	// or keep namespaces same with the source ones (under this case,
	// the DestNamespace should be set to empty)
	DestNamespace string `json:"dest_namespace"`
	// Filters
	Filters []*Filter `json:"filters"`
	// Trigger
	Trigger *Trigger `json:"trigger"`
	// Settings
	// TODO: rename the property name
	Deletion bool `json:"deletion"`
	// If override the image tag
	Override bool `json:"override"`
	// Operations
	Enabled      bool      `json:"enabled"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}

// Valid the policy
func (p *Policy) Valid(v *validation.Validation) {
	if len(p.Name) == 0 {
		v.SetError("name", "cannot be empty")
	}
	var srcRegistryID, dstRegistryID int64
	if p.SrcRegistry != nil {
		srcRegistryID = p.SrcRegistry.ID
	}
	if p.DestRegistry != nil {
		dstRegistryID = p.DestRegistry.ID
	}

	// one of the source registry and destination registry must be Harbor itself
	if srcRegistryID != 0 && dstRegistryID != 0 ||
		srcRegistryID == 0 && dstRegistryID == 0 {
		v.SetError("src_registry, dest_registry", "one of them should be empty and the other one shouldn't be empty")
	}

	// valid the filters
	for _, filter := range p.Filters {
		switch filter.Type {
		case FilterTypeResource, FilterTypeName, FilterTypeTag:
			value, ok := filter.Value.(string)
			if !ok {
				v.SetError("filters", "the type of filter value isn't string")
				break
			}
			if filter.Type == FilterTypeResource {
				rt := ResourceType(value)
				if !(rt == ResourceTypeImage || rt == ResourceTypeChart) {
					v.SetError("filters", fmt.Sprintf("invalid resource filter: %s", value))
					break
				}
			}
		case FilterTypeLabel:
			labels, ok := filter.Value.([]interface{})
			if !ok {
				v.SetError("filters", "the type of label filter value isn't string slice")
				break
			}
			for _, label := range labels {
				_, ok := label.(string)
				if !ok {
					v.SetError("filters", "the type of label filter value isn't string slice")
					break
				}
			}
		default:
			v.SetError("filters", "invalid filter type")
			break
		}
	}

	// valid trigger
	if p.Trigger != nil {
		switch p.Trigger.Type {
		case TriggerTypeManual, TriggerTypeEventBased:
		case TriggerTypeScheduled:
			if p.Trigger.Settings == nil || len(p.Trigger.Settings.Cron) == 0 {
				v.SetError("trigger", fmt.Sprintf("the cron string cannot be empty when the trigger type is %s", TriggerTypeScheduled))
			} else {
				_, err := cron.Parse(p.Trigger.Settings.Cron)
				if err != nil {
					v.SetError("trigger", fmt.Sprintf("invalid cron string for scheduled trigger: %s", p.Trigger.Settings.Cron))
				}
			}
		default:
			v.SetError("trigger", "invalid trigger type")
		}
	}
}

// FilterType represents the type info of the filter.
type FilterType string

// Filter holds the info of the filter
type Filter struct {
	Type  FilterType  `json:"type"`
	Value interface{} `json:"value"`
}

// DoFilter filter the filterables
// The parameter "filterables" must be a pointer points to a slice
// whose elements must be Filterable. After applying the filter
// to the "filterables", the result is put back into the variable
// "filterables"
func (f *Filter) DoFilter(filterables interface{}) error {
	var ft filter.Filter
	switch f.Type {
	case FilterTypeName:
		ft = filter.NewRepositoryNameFilter(f.Value.(string))
	case FilterTypeTag:
		ft = filter.NewVTagNameFilter(f.Value.(string))
	case FilterTypeLabel:
		labels, ok := f.Value.([]string)
		if ok {
			ft = filter.NewVTagLabelFilter(labels)
		}
	case FilterTypeResource:
		ft = filter.NewResourceTypeFilter(f.Value.(string))
	default:
		return fmt.Errorf("unsupported filter type: %s", f.Type)
	}

	return filter.DoFilter(filterables, ft)
}

// TriggerType represents the type of trigger.
type TriggerType string

// Trigger holds info for a trigger
type Trigger struct {
	Type     TriggerType      `json:"type"`
	Settings *TriggerSettings `json:"trigger_settings"`
}

// TriggerSettings is the setting about the trigger
type TriggerSettings struct {
	Cron string `json:"cron"`
}

// PolicyQuery defines the query conditions for listing policies
type PolicyQuery struct {
	Name string
	// TODO: need to consider how to support listing the policies
	// of one namespace in both pull and push modes
	Namespace    string
	SrcRegistry  int64
	DestRegistry int64
	models.Pagination
}
