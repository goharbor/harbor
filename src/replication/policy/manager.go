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
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	persist_models "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/models"
)

// Manager defines the method a policy manger should implement
type Manager interface {
	GetPolicies(models.QueryParameter) (*models.ReplicationPolicyQueryResult, error)
	GetPolicy(int64) (models.ReplicationPolicy, error)
	CreatePolicy(models.ReplicationPolicy) (int64, error)
	UpdatePolicy(models.ReplicationPolicy) error
	RemovePolicy(int64) error
}

// DefaultManager provides replication policy CURD capabilities.
type DefaultManager struct{}

// NewDefaultManager is the constructor of DefaultManager.
func NewDefaultManager() *DefaultManager {
	return &DefaultManager{}
}

// GetPolicies returns all the policies
func (m *DefaultManager) GetPolicies(query models.QueryParameter) (*models.ReplicationPolicyQueryResult, error) {
	result := &models.ReplicationPolicyQueryResult{
		Policies: []*models.ReplicationPolicy{},
	}
	total, err := dao.GetTotalOfRepPolicies(query.Name, query.ProjectID)
	if err != nil {
		return nil, err
	}
	result.Total = total

	policies, err := dao.FilterRepPolicies(query.Name, query.ProjectID, query.Page, query.PageSize)
	if err != nil {
		return nil, err
	}

	for _, policy := range policies {
		ply, err := convertFromPersistModel(policy)
		if err != nil {
			return nil, err
		}

		result.Policies = append(result.Policies, &ply)
	}

	return result, nil
}

// GetPolicy returns the policy with the specified ID
func (m *DefaultManager) GetPolicy(policyID int64) (models.ReplicationPolicy, error) {
	policy, err := dao.GetRepPolicy(policyID)
	if err != nil {
		return models.ReplicationPolicy{}, err
	}

	return convertFromPersistModel(policy)
}

func convertFromPersistModel(policy *persist_models.RepPolicy) (models.ReplicationPolicy, error) {
	if policy == nil {
		return models.ReplicationPolicy{}, nil
	}

	ply := models.ReplicationPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		ReplicateDeletion: policy.ReplicateDeletion,
		ProjectIDs:        []int64{policy.ProjectID},
		TargetIDs:         []int64{policy.TargetID},
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
	}

	project, err := config.GlobalProjectMgr.Get(policy.ProjectID)
	if err != nil {
		return models.ReplicationPolicy{}, err
	}
	ply.Namespaces = []string{project.Name}

	if len(policy.Filters) > 0 {
		filters := []models.Filter{}
		if err := json.Unmarshal([]byte(policy.Filters), &filters); err != nil {
			return models.ReplicationPolicy{}, err
		}
		for i := range filters {
			if filters[i].Value == nil && len(filters[i].Pattern) > 0 {
				filters[i].Value = filters[i].Pattern
			}
			// convert the type of Value to int64 as the default type of
			// json Unmarshal for number is float64
			if filters[i].Kind == replication.FilterItemKindLabel {
				filters[i].Value = int64(filters[i].Value.(float64))
			}
		}
		ply.Filters = filters
	}

	if len(policy.Trigger) > 0 {
		trigger := &models.Trigger{}
		if err := json.Unmarshal([]byte(policy.Trigger), trigger); err != nil {
			return models.ReplicationPolicy{}, err
		}
		ply.Trigger = trigger
	}

	return ply, nil
}

func convertToPersistModel(policy models.ReplicationPolicy) (*persist_models.RepPolicy, error) {
	ply := &persist_models.RepPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		ReplicateDeletion: policy.ReplicateDeletion,
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
	}

	if len(policy.ProjectIDs) > 0 {
		ply.ProjectID = policy.ProjectIDs[0]
	}

	if len(policy.TargetIDs) > 0 {
		ply.TargetID = policy.TargetIDs[0]
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

// CreatePolicy creates a new policy with the provided data;
// If creating failed, error will be returned;
// If creating succeed, ID of the new created policy will be returned.
func (m *DefaultManager) CreatePolicy(policy models.ReplicationPolicy) (int64, error) {
	now := time.Now()
	policy.CreationTime = now
	policy.UpdateTime = now
	ply, err := convertToPersistModel(policy)
	if err != nil {
		return 0, err
	}
	return dao.AddRepPolicy(*ply)
}

// UpdatePolicy updates the policy;
// If updating failed, error will be returned.
func (m *DefaultManager) UpdatePolicy(policy models.ReplicationPolicy) error {
	policy.UpdateTime = time.Now()
	ply, err := convertToPersistModel(policy)
	if err != nil {
		return err
	}
	return dao.UpdateRepPolicy(ply)
}

// RemovePolicy removes the specified policy;
// If removing failed, error will be returned.
func (m *DefaultManager) RemovePolicy(policyID int64) error {
	// delete replication jobs
	if err := dao.DeleteRepJobs(policyID); err != nil {
		return err
	}
	// delete the replication policy
	return dao.DeleteRepPolicy(policyID)
}
