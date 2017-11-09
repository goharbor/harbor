package policy

import (
	"github.com/vmware/harbor/src/replication/models"
)

//Manager provides replication policy CURD capabilities.
type Manager struct{}

//NewManager is the constructor of Manager.
func NewManager() *Manager {
	return &Manager{}
}

//GetPolicies returns all the policies
func (m *Manager) GetPolicies(query models.QueryParameter) []models.ReplicationPolicy {
	return []models.ReplicationPolicy{}
}

//GetPolicy returns the policy with the specified ID
func (m *Manager) GetPolicy(policyID int) models.ReplicationPolicy {
	return models.ReplicationPolicy{}
}

//CreatePolicy creates a new policy with the provided data;
//If creating failed, error will be returned;
//If creating succeed, ID of the new created policy will be returned.
func (m *Manager) CreatePolicy(policy models.ReplicationPolicy) (int, error) {
	return 0, nil
}

//UpdatePolicy updates the policy;
//If updating failed, error will be returned.
func (m *Manager) UpdatePolicy(policy models.ReplicationPolicy) error {
	return nil
}

//RemovePolicy removes the specified policy;
//If removing failed, error will be returned.
func (m *Manager) RemovePolicy(policyID int) error {
	return nil
}
