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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	replicationmodel "github.com/goharbor/harbor/src/pkg/replication/model"
	"github.com/robfig/cron"
)

// Policy defines the structure of a replication policy
type Policy struct {
	ID                        int64           `json:"id"`
	Name                      string          `json:"name"`
	Description               string          `json:"description"`
	Creator                   string          `json:"creator"`
	SrcRegistry               *model.Registry `json:"src_registry"`
	DestRegistry              *model.Registry `json:"dest_registry"`
	DestNamespace             string          `json:"dest_namespace"`
	DestNamespaceReplaceCount int8            `json:"dest_namespace_replace_count"`
	Filters                   []*model.Filter `json:"filters"`
	Trigger                   *model.Trigger  `json:"trigger"`
	ReplicateDeletion         bool            `json:"deletion"`
	Override                  bool            `json:"override"`
	Enabled                   bool            `json:"enabled"`
	CreationTime              time.Time       `json:"creation_time"`
	UpdateTime                time.Time       `json:"update_time"`
	Speed                     int32           `json:"speed"`
}

// IsScheduledTrigger returns true when the policy is scheduled trigger and enabled
func (p *Policy) IsScheduledTrigger() bool {
	if !p.Enabled {
		return false
	}
	if p.Trigger == nil {
		return false
	}
	return p.Trigger.Type == model.TriggerTypeScheduled
}

// Validate the policy
func (p *Policy) Validate() error {
	if len(p.Name) == 0 {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("empty name")
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
		return errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("either src_registry or dest_registry should be empty and the other one shouldn't be empty")
	}

	// valid the filters
	for _, f := range p.Filters {
		if err := f.Validate(); err != nil {
			return err
		}
	}

	// valid the destination namespace
	if len(p.DestNamespace) > 0 {
		if !lib.RepositoryNameRe.MatchString(p.DestNamespace) {
			return errors.New(nil).WithCode(errors.BadRequestCode).
				WithMessage("invalid destination namespace: %s", p.DestNamespace)
		}
	}

	// valid trigger
	if p.Trigger != nil {
		switch p.Trigger.Type {
		case model.TriggerTypeManual, model.TriggerTypeEventBased:
		case model.TriggerTypeScheduled:
			if p.Trigger.Settings == nil || len(p.Trigger.Settings.Cron) == 0 {
				return errors.New(nil).WithCode(errors.BadRequestCode).
					WithMessage("the cron string cannot be empty when the trigger type is %s", model.TriggerTypeScheduled)
			}
			if _, err := cron.Parse(p.Trigger.Settings.Cron); err != nil {
				return errors.New(nil).WithCode(errors.BadRequestCode).
					WithMessage("invalid cron string for scheduled trigger: %s", p.Trigger.Settings.Cron)
			}
		default:
			return errors.New(nil).WithCode(errors.BadRequestCode).
				WithMessage("invalid trigger type")
		}
	}
	return nil
}

// From converts the pkg model into the Policy
func (p *Policy) From(policy *replicationmodel.Policy) error {
	if policy == nil {
		return nil
	}
	p.ID = policy.ID
	p.Name = policy.Name
	p.Description = policy.Description
	p.Creator = policy.Creator
	p.DestNamespace = policy.DestNamespace
	p.DestNamespaceReplaceCount = policy.DestNamespaceReplaceCount
	p.ReplicateDeletion = policy.ReplicateDeletion
	p.Override = policy.Override
	p.Enabled = policy.Enabled
	p.CreationTime = policy.CreationTime
	p.UpdateTime = policy.UpdateTime
	p.Speed = policy.Speed

	if policy.SrcRegistryID > 0 {
		p.SrcRegistry = &model.Registry{
			ID: policy.SrcRegistryID,
		}
	}
	if policy.DestRegistryID > 0 {
		p.DestRegistry = &model.Registry{
			ID: policy.DestRegistryID,
		}
	}

	// parse Filters
	filters, err := parseFilters(policy.Filters)
	if err != nil {
		return err
	}
	p.Filters = filters

	// parse Trigger
	trigger, err := parseTrigger(policy.Trigger)
	if err != nil {
		return err
	}
	p.Trigger = trigger

	return nil
}

// To converts to pkg model
func (p *Policy) To() (*replicationmodel.Policy, error) {
	policy := &replicationmodel.Policy{
		ID:                        p.ID,
		Name:                      p.Name,
		Description:               p.Description,
		Creator:                   p.Creator,
		DestNamespace:             p.DestNamespace,
		DestNamespaceReplaceCount: p.DestNamespaceReplaceCount,
		Override:                  p.Override,
		Enabled:                   p.Enabled,
		ReplicateDeletion:         p.ReplicateDeletion,
		CreationTime:              p.CreationTime,
		UpdateTime:                p.UpdateTime,
		Speed:                     p.Speed,
	}
	if p.SrcRegistry != nil {
		policy.SrcRegistryID = p.SrcRegistry.ID
	}
	if p.DestRegistry != nil {
		policy.DestRegistryID = p.DestRegistry.ID
	}

	if p.Trigger != nil {
		trigger, err := json.Marshal(p.Trigger)
		if err != nil {
			return nil, err
		}
		policy.Trigger = string(trigger)
	}

	if len(p.Filters) > 0 {
		filters, err := json.Marshal(p.Filters)
		if err != nil {
			return nil, err
		}
		policy.Filters = string(filters)
	}

	return policy, nil
}

type filter struct {
	Type       string      `json:"type"`
	Value      interface{} `json:"value"`
	Decoration string      `json:"decoration"`
	Kind       string      `json:"kind"`
	Pattern    string      `json:"pattern"`
}

type trigger struct {
	Type          string                 `json:"type"`
	Settings      *model.TriggerSettings `json:"trigger_settings"`
	Kind          string                 `json:"kind"`
	ScheduleParam *scheduleParam         `json:"schedule_param"`
}

type scheduleParam struct {
	Type    string `json:"type"`
	Weekday int8   `json:"weekday"`
	Offtime int64  `json:"offtime"`
}

func parseFilters(str string) ([]*model.Filter, error) {
	if len(str) == 0 {
		return nil, nil
	}
	items := []*filter{}
	if err := json.Unmarshal([]byte(str), &items); err != nil {
		return nil, err
	}

	filters := []*model.Filter{}
	for _, item := range items {
		filter := &model.Filter{
			Type:       item.Type,
			Value:      item.Value,
			Decoration: item.Decoration,
		}
		// keep backwards compatibility
		if len(filter.Type) == 0 {
			if filter.Value == nil {
				filter.Value = item.Pattern
			}
			switch item.Kind {
			case "repository":
				// a name filter "project_name/**" must exist after running upgrade
				// if there is any repository filter, merge it into the name filter
				repository, ok := filter.Value.(string)
				if ok && len(repository) > 0 {
					for _, item := range items {
						if item.Type == model.FilterTypeName {
							name, ok := item.Value.(string)
							if ok && len(name) > 0 {
								item.Value = strings.Replace(name, "**", repository, 1)
							}
							break
						}
					}
				}
				continue
			case "tag":
				filter.Type = model.FilterTypeTag
			case "label":
				// drop all legend label filters
				continue
			default:
				log.Warningf("unknown filter type: %s", filter.Type)
				continue
			}
		}

		// convert the type of value from string to model.ResourceType if the filter
		// is a resource type filter
		if filter.Type == model.FilterTypeResource {
			filter.Value = filter.Value.(string)
		}
		if filter.Type == model.FilterTypeLabel {
			labels := []string{}
			for _, label := range filter.Value.([]interface{}) {
				labels = append(labels, label.(string))
			}
			filter.Value = labels
		}
		filters = append(filters, filter)
	}
	return filters, nil
}

func parseTrigger(str string) (*model.Trigger, error) {
	if len(str) == 0 {
		return nil, nil
	}
	item := &trigger{}
	if err := json.Unmarshal([]byte(str), item); err != nil {
		return nil, err
	}
	trigger := &model.Trigger{
		Type:     item.Type,
		Settings: item.Settings,
	}
	// keep backwards compatibility
	if len(trigger.Type) == 0 {
		switch item.Kind {
		case "Manual":
			trigger.Type = model.TriggerTypeManual
		case "Immediate":
			trigger.Type = model.TriggerTypeEventBased
		case "Scheduled":
			trigger.Type = model.TriggerTypeScheduled
			trigger.Settings = &model.TriggerSettings{
				Cron: parseScheduleParamToCron(item.ScheduleParam),
			}
		default:
			log.Warningf("unknown trigger type: %s", item.Kind)
			return nil, nil
		}
	}
	return trigger, nil
}

func parseScheduleParamToCron(param *scheduleParam) string {
	if param == nil {
		return ""
	}
	offtime := param.Offtime
	offtime = offtime % (3600 * 24)
	hour := int(offtime / 3600)
	offtime = offtime % 3600
	minute := int(offtime / 60)
	second := int(offtime % 60)
	if param.Type == "Weekly" {
		return fmt.Sprintf("%d %d %d * * %d", second, minute, hour, param.Weekday%7)
	}
	return fmt.Sprintf("%d %d %d * * *", second, minute, hour)
}
