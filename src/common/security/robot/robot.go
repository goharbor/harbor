package robot

import (
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/project"
)

// robot implement the rbac.User interface for project robot account
type robot struct {
	username  string
	namespace rbac.Namespace
	policies  []*rbac.Policy
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
	policies = append(policies, r.policies...)
	return policies
}

// GetRoles robot has no definition of role, always return nil here.
func (r *robot) GetRoles() []rbac.Role {
	return nil
}

// NewRobot ...
func NewRobot(username string, namespace rbac.Namespace, policies []*rbac.Policy) rbac.User {
	return &robot{
		username:  username,
		namespace: namespace,
		policies:  filterPolicies(namespace, policies),
	}
}

func filterPolicies(namespace rbac.Namespace, policies []*rbac.Policy) []*rbac.Policy {
	var results []*rbac.Policy
	if len(policies) == 0 {
		return results
	}

	mp := getAllPolicies(namespace)
	for _, policy := range policies {
		if mp[policy.String()] {
			results = append(results, policy)
		}
	}
	return results
}

// getAllPolicies gets all of supported policies supported in project and external policies supported for robot account
func getAllPolicies(namespace rbac.Namespace) map[string]bool {
	mp := map[string]bool{}
	for _, policy := range project.GetAllPolicies(namespace) {
		mp[policy.String()] = true
	}
	scannerPull := &rbac.Policy{Resource: namespace.Resource(rbac.ResourceRepository), Action: rbac.ActionScannerPull}
	mp[scannerPull.String()] = true
	return mp
}
