package robot

import (
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// robot implement the rbac.User interface for project robot account
type robot struct {
	username  string
	namespace types.Namespace
	policies  []*types.Policy
}

// GetUserName get the robot name.
func (r *robot) GetUserName() string {
	return r.username
}

// GetPolicies ...
func (r *robot) GetPolicies() []*types.Policy {
	return r.policies
}

// GetRoles robot has no definition of role, always return nil here.
func (r *robot) GetRoles() []types.RBACRole {
	return nil
}

// NewRobot ...
func NewRobot(username string, namespace types.Namespace, policies []*types.Policy) types.RBACUser {
	return &robot{
		username:  username,
		namespace: namespace,
		policies:  filterPolicies(namespace, policies),
	}
}

func filterPolicies(namespace types.Namespace, policies []*types.Policy) []*types.Policy {
	var results []*types.Policy
	for _, policy := range policies {
		if types.ResourceAllowedInNamespace(policy.Resource, namespace) {
			results = append(results, policy)
		}
	}

	return results
}
