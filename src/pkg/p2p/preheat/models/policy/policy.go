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

package policy

import (
	"encoding/json"
	"strconv"
	"time"

	beego_orm "github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
)

func init() {
	beego_orm.RegisterModel(&Schema{})
}

// ScopeType represents the preheat scope type.
type ScopeType = string

const (
	// Filters:
	//  Repository : type=Repository value=name text (double star pattern used)
	//  Tag: type=Tag value=tag text (double star pattern used)
	//  Signature: type=Signature value=bool (True/False)
	//  Vulnerability: type=Vulnerability value=Severity (int) (expected bar)
	//  Label: type=Label value=label array (with format: lb1,lb2,lb3)

	// FilterTypeRepository represents the repository filter type
	FilterTypeRepository FilterType = "repository"
	// FilterTypeTag represents the tag filter type
	FilterTypeTag FilterType = "tag"
	// FilterTypeSignature represents the signature filter type
	FilterTypeSignature FilterType = "signature"
	// FilterTypeVulnerability represents the vulnerability filter type
	FilterTypeVulnerability FilterType = "vulnerability"
	// FilterTypeLabel represents the label filter type
	FilterTypeLabel FilterType = "label"

	// TriggerTypeManual represents the manual trigger type
	TriggerTypeManual TriggerType = "manual"
	// TriggerTypeScheduled represents the scheduled trigger type
	TriggerTypeScheduled TriggerType = "scheduled"
	// TriggerTypeEventBased represents the event_based trigger type
	TriggerTypeEventBased TriggerType = "event_based"

	// ScopeTypeSinglePeer represents preheat image to single peer in p2p cluster.
	ScopeTypeSinglePeer ScopeType = "single_peer"
	// ScopeTypeAllPeers represents preheat image to all peers in p2p cluster.
	ScopeTypeAllPeers ScopeType = "all_peers"
)

// Schema defines p2p preheat policy schema
type Schema struct {
	ID          int64  `orm:"column(id)" json:"id"`
	Name        string `orm:"column(name)" json:"name"`
	Description string `orm:"column(description)" json:"description"`
	// use project id
	ProjectID  int64     `orm:"column(project_id)" json:"project_id"`
	ProviderID int64     `orm:"column(provider_id)" json:"provider_id"`
	Filters    []*Filter `orm:"-" json:"filters"`
	// Use JSON data format （query by filter type should be supported)
	FiltersStr string   `orm:"column(filters)" json:"-"`
	Trigger    *Trigger `orm:"-" json:"trigger"`
	// Use JSON data format (query by trigger type should be supported)
	TriggerStr string `orm:"column(trigger)" json:"-"`
	Enabled    bool   `orm:"column(enabled)" json:"enabled"`
	// Scope decides the preheat scope.
	Scope string `orm:"column(scope)" json:"scope"`
	// ExtraAttrs is used to store extra attributes provided by vendor.
	ExtraAttrsStr string                 `orm:"column(extra_attrs)" json:"-"`
	ExtraAttrs    map[string]interface{} `orm:"-" json:"extra_attrs"`
	CreatedAt     time.Time              `orm:"column(creation_time)" json:"creation_time"`
	UpdatedTime   time.Time              `orm:"column(update_time)" json:"update_time"`
}

// TableName specifies the policy schema table name.
func (s *Schema) TableName() string {
	return "p2p_preheat_policy"
}

// GetDefaultSorts specifies the default sorts
func (s *Schema) GetDefaultSorts() []*q.Sort {
	return []*q.Sort{
		{
			Key: "UpdatedTime",
		},
		{
			Key:  "ID",
			DESC: true,
		},
	}
}

// FilterType represents the type info of the filter.
type FilterType = string

// Filter holds the info of the filter
type Filter struct {
	Type  FilterType  `json:"type"`
	Value interface{} `json:"value"`
}

// TriggerType represents the type of trigger.
type TriggerType = string

// Trigger holds the trigger info.
type Trigger struct {
	// The preheat policy trigger type. The valid values ar manual, scheduled.
	Type     TriggerType `json:"type"`
	Settings struct {
		// The cron string for scheduled trigger.
		Cron string `json:"cron,omitempty"`
	} `json:"trigger_setting,omitempty"`
}

// ValidatePreheatPolicy validate preheat policy
func (s *Schema) ValidatePreheatPolicy() error {
	// currently only validate cron string of preheat policy
	if s.Trigger != nil && s.Trigger.Type == TriggerTypeScheduled && len(s.Trigger.Settings.Cron) > 0 {
		if err := utils.ValidateCronString(s.Trigger.Settings.Cron); err != nil {
			return errors.New(nil).WithCode(errors.BadRequestCode).
				WithMessagef("invalid cron string for scheduled preheat: %s, error: %v", s.Trigger.Settings.Cron, err)
		}
	}

	// validate preheat scope
	if s.Scope != "" && s.Scope != ScopeTypeSinglePeer && s.Scope != ScopeTypeAllPeers {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("invalid scope for preheat policy: %s", s.Scope)
	}

	return nil
}

// Encode encodes policy schema.
func (s *Schema) Encode() error {
	if s.Filters != nil {
		filterStr, err := json.Marshal(s.Filters)
		if err != nil {
			return err
		}
		s.FiltersStr = string(filterStr)
	}

	if s.Trigger != nil {
		triggerStr, err := json.Marshal(s.Trigger)
		if err != nil {
			return err
		}
		s.TriggerStr = string(triggerStr)
	}

	if s.ExtraAttrs != nil {
		extraAttrsStr, err := json.Marshal(s.ExtraAttrs)
		if err != nil {
			return err
		}
		s.ExtraAttrsStr = string(extraAttrsStr)
	}

	return nil
}

// Decode decodes policy schema.
func (s *Schema) Decode() error {
	// parse filters
	filters, err := decodeFilters(s.FiltersStr)
	if err != nil {
		return err
	}
	s.Filters = filters

	// parse trigger
	trigger, err := decodeTrigger(s.TriggerStr)
	if err != nil {
		return err
	}
	s.Trigger = trigger

	// parse extra attributes
	extraAttrs, err := decodeExtraAttrs(s.ExtraAttrsStr)
	if err != nil {
		return err
	}
	s.ExtraAttrs = extraAttrs

	return nil
}

// decodeFilters parse filterStr to filter.
func decodeFilters(filterStr string) ([]*Filter, error) {
	// Filters are required
	if len(filterStr) == 0 {
		return nil, errors.New("missing filters in preheat policy schema")
	}

	var filters []*Filter
	if err := json.Unmarshal([]byte(filterStr), &filters); err != nil {
		return nil, err
	}

	// Convert value type
	// TODO: remove switch after UI bug #12579 fixed
	for _, f := range filters {
		if f.Type == FilterTypeVulnerability {
			switch f.Value.(type) {
			case string:
				sev, err := strconv.ParseInt(f.Value.(string), 10, 32)
				if err != nil {
					return nil, errors.Wrapf(err, "parse filters")
				}
				f.Value = (int)(sev)
			case float64:
				f.Value = (int)(f.Value.(float64))
			}
		}
	}

	return filters, nil
}

// decodeTrigger parse triggerStr to trigger.
func decodeTrigger(triggerStr string) (*Trigger, error) {
	// trigger must be existing, at least is a "manual" trigger.
	if len(triggerStr) == 0 {
		return nil, errors.New("missing trigger settings in preheat policy schema")
	}

	trigger := &Trigger{}
	if err := json.Unmarshal([]byte(triggerStr), trigger); err != nil {
		return nil, err
	}

	return trigger, nil
}

// decodeExtraAttrs parse extraAttrsStr to extraAttrs.
func decodeExtraAttrs(extraAttrsStr string) (map[string]interface{}, error) {
	if len(extraAttrsStr) == 0 {
		return nil, nil
	}

	extraAttrs := make(map[string]interface{})
	if err := json.Unmarshal([]byte(extraAttrsStr), &extraAttrs); err != nil {
		return nil, err
	}

	return extraAttrs, nil
}
