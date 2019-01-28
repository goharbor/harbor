package robot

import (
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/project"
)

// robot implement the rbac.User interface for project robot account
type robot struct {
	username  string
	namespace rbac.Namespace
	policy    []*rbac.Policy
}

// GetUserName get the robot name.
func (r *robot) GetUserName() string {
	return r.username
}

// GetPolicies ...
func (r *robot) GetPolicies() []*rbac.Policy {
	policies := []*rbac.Policy{}
	if r.namespace.IsPublic() {
		policies = append(policies, project.PoliciesForPublicProject(r.namespace)...)
	}
	policies = append(policies, r.policy...)
	return policies
}

// GetRoles robot has no definition of role, always return nil here.
func (r *robot) GetRoles() []rbac.Role {
	return nil
}

// NewRobot ...
func NewRobot(username string, namespace rbac.Namespace, policy []*rbac.Policy) rbac.User {
	return &robot{
		username:  username,
		namespace: namespace,
		policy:    policy,
	}
}
