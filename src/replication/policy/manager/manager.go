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

package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/dao"
	persist_models "github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/policy"
)

var errNilPolicyModel = errors.New("nil policy model")

func convertFromPersistModel(policy *persist_models.RepPolicy) (*model.Policy, error) {
	if policy == nil {
		return nil, nil
	}

	ply := model.Policy{
		ID:            policy.ID,
		Name:          policy.Name,
		Description:   policy.Description,
		Creator:       policy.Creator,
		DestNamespace: policy.DestNamespace,
		Deletion:      policy.ReplicateDeletion,
		Override:      policy.Override,
		Enabled:       policy.Enabled,
		CreationTime:  policy.CreationTime,
		UpdateTime:    policy.UpdateTime,
	}
	if policy.SrcRegistryID > 0 {
		ply.SrcRegistry = &model.Registry{
			ID: policy.SrcRegistryID,
		}
	}
	if policy.DestRegistryID > 0 {
		ply.DestRegistry = &model.Registry{
			ID: policy.DestRegistryID,
		}
	}

	// parse Filters
	filters, err := parseFilters(policy.Filters)
	if err != nil {
		return nil, err
	}
	ply.Filters = filters

	// parse Trigger
	trigger, err := parseTrigger(policy.Trigger)
	if err != nil {
		return nil, err
	}
	ply.Trigger = trigger

	return &ply, nil
}

func convertToPersistModel(policy *model.Policy) (*persist_models.RepPolicy, error) {
	if policy == nil {
		return nil, errNilPolicyModel
	}

	ply := &persist_models.RepPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		Creator:           policy.Creator,
		DestNamespace:     policy.DestNamespace,
		Override:          policy.Override,
		Enabled:           policy.Enabled,
		ReplicateDeletion: policy.Deletion,
		CreationTime:      policy.CreationTime,
		UpdateTime:        time.Now(),
	}
	if policy.SrcRegistry != nil {
		ply.SrcRegistryID = policy.SrcRegistry.ID
	}
	if policy.DestRegistry != nil {
		ply.DestRegistryID = policy.DestRegistry.ID
	}

	if policy.Trigger != nil {
		trigger, err := json.Marshal(policy.Trigger)
		if err != nil {
			return nil, err
		}
		ply.Trigger = string(trigger)
	}

	if len(policy.Filters) > 0 {
		filters, err := json.Marshal(policy.Filters)
		if err != nil {
			return nil, err
		}
		ply.Filters = string(filters)
	}

	return ply, nil
}

// DefaultManager provides replication policy CURD capabilities.
type DefaultManager struct{}

var _ policy.Controller = &DefaultManager{}

// NewDefaultManager is the constructor of DefaultManager.
func NewDefaultManager() *DefaultManager {
	return &DefaultManager{}
}

// Create creates a new policy with the provided data;
// If creating failed, error will be returned;
// If creating succeed, ID of the new created policy will be returned.
func (m *DefaultManager) Create(policy *model.Policy) (int64, error) {
	ply, err := convertToPersistModel(policy)
	if err != nil {
		return 0, err
	}
	return dao.AddRepPolicy(ply)
}

// List returns all the policies
func (m *DefaultManager) List(queries ...*model.PolicyQuery) (total int64, policies []*model.Policy, err error) {
	var persistPolicies []*persist_models.RepPolicy
	total, persistPolicies, err = dao.GetPolicies(queries...)
	if err != nil {
		return
	}

	for _, policy := range persistPolicies {
		ply, err := convertFromPersistModel(policy)
		if err != nil {
			return 0, nil, err
		}

		policies = append(policies, ply)
	}

	if policies == nil {
		policies = []*model.Policy{}
	}

	return
}

// Get returns the policy with the specified ID
func (m *DefaultManager) Get(policyID int64) (*model.Policy, error) {
	policy, err := dao.GetRepPolicy(policyID)
	if err != nil {
		return nil, err
	}

	return convertFromPersistModel(policy)
}

// GetByName returns the policy with the specified name
func (m *DefaultManager) GetByName(name string) (*model.Policy, error) {
	policy, err := dao.GetRepPolicyByName(name)
	if err != nil {
		return nil, err
	}

	return convertFromPersistModel(policy)
}

// Update Update the specified policy
func (m *DefaultManager) Update(policy *model.Policy) error {
	updatePolicy, err := convertToPersistModel(policy)
	if err != nil {
		return err
	}

	return dao.UpdateRepPolicy(updatePolicy)
}

// Remove Remove the specified policy
func (m *DefaultManager) Remove(policyID int64) error {
	return dao.DeleteRepPolicy(policyID)
}

type filter struct {
	Type    model.FilterType `json:"type"`
	Value   interface{}      `json:"value"`
	Kind    string           `json:"kind"`
	Pattern string           `json:"pattern"`
}

type trigger struct {
	Type          model.TriggerType      `json:"type"`
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
			Type:  item.Type,
			Value: item.Value,
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
			filter.Value = (model.ResourceType)(filter.Value.(string))
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
