package policy

import (
	"encoding/json"
	"time"

	"github.com/vmware/harbor/src/common/dao"
	persist_models "github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/replication/models"
	"github.com/vmware/harbor/src/ui/config"
)

//Manager provides replication policy CURD capabilities.
type Manager struct{}

//NewManager is the constructor of Manager.
func NewManager() *Manager {
	return &Manager{}
}

//GetPolicies returns all the policies
func (m *Manager) GetPolicies(query models.QueryParameter) ([]models.ReplicationPolicy, error) {
	result := []models.ReplicationPolicy{}
	//TODO support more query conditions other than name and project ID
	policies, err := dao.FilterRepPolicies(query.Name, query.ProjectID)
	if err != nil {
		return result, err
	}

	for _, policy := range policies {
		ply, err := convertFromPersistModel(policy)
		if err != nil {
			return []models.ReplicationPolicy{}, err
		}
		result = append(result, ply)
	}

	return result, nil
}

//GetPolicy returns the policy with the specified ID
func (m *Manager) GetPolicy(policyID int64) (models.ReplicationPolicy, error) {
	policy, err := dao.GetRepPolicy(policyID)
	if err != nil {
		return models.ReplicationPolicy{}, err
	}

	return convertFromPersistModel(policy)
}

// TODO add UT
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

// TODO add ut
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

//CreatePolicy creates a new policy with the provided data;
//If creating failed, error will be returned;
//If creating succeed, ID of the new created policy will be returned.
func (m *Manager) CreatePolicy(policy models.ReplicationPolicy) (int64, error) {
	now := time.Now()
	policy.CreationTime = now
	policy.UpdateTime = now
	ply, err := convertToPersistModel(policy)
	if err != nil {
		return 0, err
	}
	return dao.AddRepPolicy(*ply)
}

//UpdatePolicy updates the policy;
//If updating failed, error will be returned.
func (m *Manager) UpdatePolicy(policy models.ReplicationPolicy) error {
	policy.UpdateTime = time.Now()
	ply, err := convertToPersistModel(policy)
	if err != nil {
		return err
	}
	return dao.UpdateRepPolicy(ply)
}

//RemovePolicy removes the specified policy;
//If removing failed, error will be returned.
func (m *Manager) RemovePolicy(policyID int64) error {
	return dao.DeleteRepPolicy(policyID)
}
