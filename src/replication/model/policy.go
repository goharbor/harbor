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

	"github.com/astaxie/beego/validation"
	"github.com/goharbor/harbor/src/common/models"
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
	// TODO consider to remove this property?
	Creator string `json:"creator"`
	// source
	SrcRegistry *Registry `json:"src_registry"`
	// destination
	// TODO rename to DstRegistry
	DestRegistry *Registry `json:"dest_registry"`
	// Only support two dest namespace modes:
	// Put all the src resources to the one single dest namespace
	// or keep namespaces same with the source ones (under this case,
	// the DestNamespace should be set to empty)
	// TODO rename to DstNamespace
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
	if p.SrcRegistry == nil || p.SrcRegistry.ID == 0 {
		v.SetError("src_registry", "cannot be empty")
	}
	if p.DestRegistry == nil || p.DestRegistry.ID == 0 {
		v.SetError("dest_registry", "cannot be empty")
	}
	// valid the filters
	for _, filter := range p.Filters {
		if filter.Type != FilterTypeResource &&
			filter.Type != FilterTypeName &&
			filter.Type != FilterTypeTag &&
			filter.Type != FilterTypeLabel {
			v.SetError("filters", "invalid filter type")
			break
		}
	}
	// valid trigger
	if p.Trigger != nil {
		if p.Trigger.Type != TriggerTypeManual &&
			p.Trigger.Type != TriggerTypeScheduled &&
			p.Trigger.Type != TriggerTypeEventBased {
			v.SetError("trigger", "invalid trigger type")
		}
		if p.Trigger.Type == TriggerTypeScheduled &&
			(p.Trigger.Settings == nil || len(p.Trigger.Settings.Cron) == 0) {
			v.SetError("trigger", fmt.Sprintf("the cron string cannot be empty when the trigger type is %s", TriggerTypeScheduled))
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
