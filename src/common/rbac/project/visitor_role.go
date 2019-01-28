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

package project

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
)

var (
	rolePoliciesMap = map[string][]*rbac.Policy{
		"projectAdmin": {
			{Resource: ResourceSelf, Action: ActionRead},
			{Resource: ResourceSelf, Action: ActionUpdate},
			{Resource: ResourceSelf, Action: ActionDelete},

			{Resource: ResourceMember, Action: ActionCreate},
			{Resource: ResourceMember, Action: ActionUpdate},
			{Resource: ResourceMember, Action: ActionDelete},
			{Resource: ResourceMember, Action: ActionList},

			{Resource: ResourceLog, Action: ActionList},

			{Resource: ResourceReplication, Action: ActionRead},
			{Resource: ResourceReplication, Action: ActionList},

			{Resource: ResourceLabel, Action: ActionCreate},
			{Resource: ResourceLabel, Action: ActionUpdate},
			{Resource: ResourceLabel, Action: ActionDelete},
			{Resource: ResourceLabel, Action: ActionList},

			{Resource: ResourceRepository, Action: ActionCreate},
			{Resource: ResourceRepository, Action: ActionUpdate},
			{Resource: ResourceRepository, Action: ActionDelete},
			{Resource: ResourceRepository, Action: ActionList},
			{Resource: ResourceRepository, Action: ActionPushPull}, // compatible with security all perm of project
			{Resource: ResourceRepository, Action: ActionPush},
			{Resource: ResourceRepository, Action: ActionPull},

			{Resource: ResourceRepositoryTag, Action: ActionDelete},
			{Resource: ResourceRepositoryTag, Action: ActionList},
			{Resource: ResourceRepositoryTag, Action: ActionScan},

			{Resource: ResourceRepositoryTagVulnerability, Action: ActionList},

			{Resource: ResourceRepositoryTagManifest, Action: ActionRead},

			{Resource: ResourceRepositoryTagLabel, Action: ActionCreate},
			{Resource: ResourceRepositoryTagLabel, Action: ActionDelete},

			{Resource: ResourceHelmChart, Action: ActionCreate}, // upload helm chart
			{Resource: ResourceHelmChart, Action: ActionRead},   // download helm chart
			{Resource: ResourceHelmChart, Action: ActionDelete},
			{Resource: ResourceHelmChart, Action: ActionList},

			{Resource: ResourceHelmChartVersion, Action: ActionCreate}, // upload helm chart version
			{Resource: ResourceHelmChartVersion, Action: ActionRead},   // read and download helm chart version
			{Resource: ResourceHelmChartVersion, Action: ActionDelete},
			{Resource: ResourceHelmChartVersion, Action: ActionList},

			{Resource: ResourceHelmChartVersionLabel, Action: ActionCreate},
			{Resource: ResourceHelmChartVersionLabel, Action: ActionDelete},

			{Resource: ResourceConfiguration, Action: ActionRead},
			{Resource: ResourceConfiguration, Action: ActionUpdate},

			{Resource: ResourceRobot, Action: ActionCreate},
			{Resource: ResourceRobot, Action: ActionRead},
			{Resource: ResourceRobot, Action: ActionUpdate},
			{Resource: ResourceRobot, Action: ActionDelete},
			{Resource: ResourceRobot, Action: ActionList},
		},

		"master": {
			{Resource: ResourceSelf, Action: ActionRead},

			{Resource: ResourceMember, Action: ActionList},

			{Resource: ResourceLog, Action: ActionList},

			{Resource: ResourceReplication, Action: ActionRead},
			{Resource: ResourceReplication, Action: ActionList},

			{Resource: ResourceLabel, Action: ActionCreate},
			{Resource: ResourceLabel, Action: ActionUpdate},
			{Resource: ResourceLabel, Action: ActionDelete},
			{Resource: ResourceLabel, Action: ActionList},

			{Resource: ResourceRepository, Action: ActionCreate},
			{Resource: ResourceRepository, Action: ActionUpdate},
			{Resource: ResourceRepository, Action: ActionDelete},
			{Resource: ResourceRepository, Action: ActionList},
			{Resource: ResourceRepository, Action: ActionPush},
			{Resource: ResourceRepository, Action: ActionPull},

			{Resource: ResourceRepositoryTag, Action: ActionDelete},
			{Resource: ResourceRepositoryTag, Action: ActionList},
			{Resource: ResourceRepositoryTag, Action: ActionScan},

			{Resource: ResourceRepositoryTagVulnerability, Action: ActionList},

			{Resource: ResourceRepositoryTagManifest, Action: ActionRead},

			{Resource: ResourceRepositoryTagLabel, Action: ActionCreate},
			{Resource: ResourceRepositoryTagLabel, Action: ActionDelete},

			{Resource: ResourceHelmChart, Action: ActionCreate},
			{Resource: ResourceHelmChart, Action: ActionRead},
			{Resource: ResourceHelmChart, Action: ActionDelete},
			{Resource: ResourceHelmChart, Action: ActionList},

			{Resource: ResourceHelmChartVersion, Action: ActionCreate},
			{Resource: ResourceHelmChartVersion, Action: ActionRead},
			{Resource: ResourceHelmChartVersion, Action: ActionDelete},
			{Resource: ResourceHelmChartVersion, Action: ActionList},

			{Resource: ResourceHelmChartVersionLabel, Action: ActionCreate},
			{Resource: ResourceHelmChartVersionLabel, Action: ActionDelete},

			{Resource: ResourceConfiguration, Action: ActionRead},
			{Resource: ResourceConfiguration, Action: ActionUpdate},
		},

		"developer": {
			{Resource: ResourceSelf, Action: ActionRead},

			{Resource: ResourceMember, Action: ActionList},

			{Resource: ResourceLog, Action: ActionList},

			{Resource: ResourceRepository, Action: ActionCreate},
			{Resource: ResourceRepository, Action: ActionList},
			{Resource: ResourceRepository, Action: ActionPush},
			{Resource: ResourceRepository, Action: ActionPull},

			{Resource: ResourceRepositoryTag, Action: ActionList},

			{Resource: ResourceRepositoryTagVulnerability, Action: ActionList},

			{Resource: ResourceRepositoryTagManifest, Action: ActionRead},

			{Resource: ResourceRepositoryTagLabel, Action: ActionCreate},
			{Resource: ResourceRepositoryTagLabel, Action: ActionDelete},

			{Resource: ResourceHelmChart, Action: ActionCreate},
			{Resource: ResourceHelmChart, Action: ActionRead},
			{Resource: ResourceHelmChart, Action: ActionList},

			{Resource: ResourceHelmChartVersion, Action: ActionCreate},
			{Resource: ResourceHelmChartVersion, Action: ActionRead},
			{Resource: ResourceHelmChartVersion, Action: ActionList},

			{Resource: ResourceHelmChartVersionLabel, Action: ActionCreate},
			{Resource: ResourceHelmChartVersionLabel, Action: ActionDelete},

			{Resource: ResourceConfiguration, Action: ActionRead},
		},

		"guest": {
			{Resource: ResourceSelf, Action: ActionRead},

			{Resource: ResourceMember, Action: ActionList},

			{Resource: ResourceLog, Action: ActionList},

			{Resource: ResourceRepository, Action: ActionList},
			{Resource: ResourceRepository, Action: ActionPull},

			{Resource: ResourceRepositoryTag, Action: ActionList},

			{Resource: ResourceRepositoryTagVulnerability, Action: ActionList},

			{Resource: ResourceRepositoryTagManifest, Action: ActionRead},

			{Resource: ResourceHelmChart, Action: ActionRead},
			{Resource: ResourceHelmChart, Action: ActionList},

			{Resource: ResourceHelmChartVersion, Action: ActionRead},
			{Resource: ResourceHelmChartVersion, Action: ActionList},

			{Resource: ResourceConfiguration, Action: ActionRead},
		},
	}
)

// visitorRole implement the rbac.Role interface
type visitorRole struct {
	namespace rbac.Namespace
	roleID    int
}

// GetRoleName returns role name for the visitor role
func (role *visitorRole) GetRoleName() string {
	switch role.roleID {
	case common.RoleProjectAdmin:
		return "projectAdmin"
	case common.RoleMaster:
		return "master"
	case common.RoleDeveloper:
		return "developer"
	case common.RoleGuest:
		return "guest"
	default:
		return ""
	}
}

// GetPolicies returns policies for the visitor role
func (role *visitorRole) GetPolicies() []*rbac.Policy {
	policies := []*rbac.Policy{}

	roleName := role.GetRoleName()
	if roleName == "" {
		return policies
	}

	for _, policy := range rolePoliciesMap[roleName] {
		policies = append(policies, &rbac.Policy{
			Resource: role.namespace.Resource(policy.Resource),
			Action:   policy.Action,
			Effect:   policy.Effect,
		})
	}

	return policies
}
