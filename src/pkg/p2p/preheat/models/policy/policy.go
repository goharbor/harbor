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
	"fmt"
	"strconv"
	"time"

	beego_orm "github.com/beego/beego/orm"
	"github.com/beego/beego/validation"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/robfig/cron"
)

func init() {
	beego_orm.RegisterModel(&Schema{})
}

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
	TriggerStr  string    `orm:"column(trigger)" json:"-"`
	Enabled     bool      `orm:"column(enabled)" json:"enabled"`
	CreatedAt   time.Time `orm:"column(creation_time)" json:"creation_time"`
	UpdatedTime time.Time `orm:"column(update_time)" json:"update_time"`
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

// Valid the policy
func (s *Schema) Valid(v *validation.Validation) {
	if len(s.Name) == 0 {
		v.SetError("name", "cannot be empty")
	}

	// valid the filters
	for _, filter := range s.Filters {
		switch filter.Type {
		case FilterTypeRepository, FilterTypeTag, FilterTypeVulnerability:
			_, ok := filter.Value.(string)
			if !ok {
				v.SetError("filters", "the type of filter value isn't string")
				break
			}
		case FilterTypeSignature:
			_, ok := filter.Value.(bool)
			if !ok {
				v.SetError("filers", "the type of signature filter value isn't bool")
				break
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
	if s.Trigger != nil {
		switch s.Trigger.Type {
		case TriggerTypeManual, TriggerTypeEventBased:
		case TriggerTypeScheduled:
			if len(s.Trigger.Settings.Cron) == 0 {
				v.SetError("trigger", fmt.Sprintf("the cron string cannot be empty when the trigger type is %s", TriggerTypeScheduled))
			} else {
				_, err := cron.Parse(s.Trigger.Settings.Cron)
				if err != nil {
					v.SetError("trigger", fmt.Sprintf("invalid cron string for scheduled trigger: %s", s.Trigger.Settings.Cron))
				}
			}
		default:
			v.SetError("trigger", "invalid trigger type")
		}
	}
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
