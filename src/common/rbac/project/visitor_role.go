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
			{Resource: rbac.ResourceSelf, Action: rbac.ActionRead},
			{Resource: rbac.ResourceSelf, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceSelf, Action: rbac.ActionDelete},

			{Resource: rbac.ResourceMember, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceMember, Action: rbac.ActionRead},
			{Resource: rbac.ResourceMember, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceMember, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceMember, Action: rbac.ActionList},

			{Resource: rbac.ResourceMetadata, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceMetadata, Action: rbac.ActionRead},
			{Resource: rbac.ResourceMetadata, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceMetadata, Action: rbac.ActionDelete},

			{Resource: rbac.ResourceLog, Action: rbac.ActionList},

			{Resource: rbac.ResourceReplication, Action: rbac.ActionRead},
			{Resource: rbac.ResourceReplication, Action: rbac.ActionList},

			{Resource: rbac.ResourceReplicationJob, Action: rbac.ActionRead},
			{Resource: rbac.ResourceReplicationJob, Action: rbac.ActionList},

			{Resource: rbac.ResourceLabel, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceLabel, Action: rbac.ActionRead},
			{Resource: rbac.ResourceLabel, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceLabel, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceLabelResource, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepository, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionList},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionPull},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionPush},

			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionRead},
			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionList},
			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionOperate},

			{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryLabel, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRepositoryLabel, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceRepositoryLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTagScanJob, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRepositoryTagScanJob, Action: rbac.ActionRead},

			{Resource: rbac.ResourceRepositoryTagVulnerability, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTagManifest, Action: rbac.ActionRead},

			{Resource: rbac.ResourceRepositoryTagLabel, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRepositoryTagLabel, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceRepositoryTagLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionCreate}, // upload helm chart
			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionRead},   // download helm chart
			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionList},

			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionCreate}, // upload helm chart version
			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionRead},   // read and download helm chart version
			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionList},

			{Resource: rbac.ResourceHelmChartVersionLabel, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceHelmChartVersionLabel, Action: rbac.ActionDelete},

			{Resource: rbac.ResourceConfiguration, Action: rbac.ActionRead},
			{Resource: rbac.ResourceConfiguration, Action: rbac.ActionUpdate},

			{Resource: rbac.ResourceRobot, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRobot, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRobot, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceRobot, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceRobot, Action: rbac.ActionList},

			{Resource: rbac.ResourceNotificationPolicy, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceNotificationPolicy, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceNotificationPolicy, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceNotificationPolicy, Action: rbac.ActionList},
			{Resource: rbac.ResourceNotificationPolicy, Action: rbac.ActionRead},
		},

		"master": {
			{Resource: rbac.ResourceSelf, Action: rbac.ActionRead},

			{Resource: rbac.ResourceMember, Action: rbac.ActionRead},
			{Resource: rbac.ResourceMember, Action: rbac.ActionList},

			{Resource: rbac.ResourceMetadata, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceMetadata, Action: rbac.ActionRead},
			{Resource: rbac.ResourceMetadata, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceMetadata, Action: rbac.ActionDelete},

			{Resource: rbac.ResourceLog, Action: rbac.ActionList},

			{Resource: rbac.ResourceReplication, Action: rbac.ActionRead},
			{Resource: rbac.ResourceReplication, Action: rbac.ActionList},

			{Resource: rbac.ResourceLabel, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceLabel, Action: rbac.ActionRead},
			{Resource: rbac.ResourceLabel, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceLabel, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepository, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionList},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionPush},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionPull},

			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionRead},
			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionList},
			{Resource: rbac.ResourceTagRetention, Action: rbac.ActionOperate},

			{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryLabel, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRepositoryLabel, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceRepositoryLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTagScanJob, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRepositoryTagScanJob, Action: rbac.ActionRead},

			{Resource: rbac.ResourceRepositoryTagVulnerability, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTagManifest, Action: rbac.ActionRead},

			{Resource: rbac.ResourceRepositoryTagLabel, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRepositoryTagLabel, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceRepositoryTagLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionRead},
			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionList},

			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionRead},
			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionList},

			{Resource: rbac.ResourceHelmChartVersionLabel, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceHelmChartVersionLabel, Action: rbac.ActionDelete},

			{Resource: rbac.ResourceConfiguration, Action: rbac.ActionRead},

			{Resource: rbac.ResourceRobot, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRobot, Action: rbac.ActionList},

			{Resource: rbac.ResourceNotificationPolicy, Action: rbac.ActionList},
		},

		"developer": {
			{Resource: rbac.ResourceSelf, Action: rbac.ActionRead},

			{Resource: rbac.ResourceMember, Action: rbac.ActionRead},
			{Resource: rbac.ResourceMember, Action: rbac.ActionList},

			{Resource: rbac.ResourceLog, Action: rbac.ActionList},

			{Resource: rbac.ResourceLabel, Action: rbac.ActionRead},
			{Resource: rbac.ResourceLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepository, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionUpdate},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionList},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionPush},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionPull},

			{Resource: rbac.ResourceRepositoryLabel, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRepositoryLabel, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceRepositoryLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTagVulnerability, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTagManifest, Action: rbac.ActionRead},

			{Resource: rbac.ResourceRepositoryTagLabel, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceRepositoryTagLabel, Action: rbac.ActionDelete},
			{Resource: rbac.ResourceRepositoryTagLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionRead},
			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionList},

			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionRead},
			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionList},

			{Resource: rbac.ResourceHelmChartVersionLabel, Action: rbac.ActionCreate},
			{Resource: rbac.ResourceHelmChartVersionLabel, Action: rbac.ActionDelete},

			{Resource: rbac.ResourceConfiguration, Action: rbac.ActionRead},

			{Resource: rbac.ResourceRobot, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRobot, Action: rbac.ActionList},
		},

		"guest": {
			{Resource: rbac.ResourceSelf, Action: rbac.ActionRead},

			{Resource: rbac.ResourceMember, Action: rbac.ActionRead},
			{Resource: rbac.ResourceMember, Action: rbac.ActionList},

			{Resource: rbac.ResourceLog, Action: rbac.ActionList},

			{Resource: rbac.ResourceLabel, Action: rbac.ActionRead},
			{Resource: rbac.ResourceLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepository, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionList},
			{Resource: rbac.ResourceRepository, Action: rbac.ActionPull},

			{Resource: rbac.ResourceRepositoryLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRepositoryTag, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTagLabel, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTagVulnerability, Action: rbac.ActionList},

			{Resource: rbac.ResourceRepositoryTagManifest, Action: rbac.ActionRead},

			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionRead},
			{Resource: rbac.ResourceHelmChart, Action: rbac.ActionList},

			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionRead},
			{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionList},

			{Resource: rbac.ResourceConfiguration, Action: rbac.ActionRead},

			{Resource: rbac.ResourceRobot, Action: rbac.ActionRead},
			{Resource: rbac.ResourceRobot, Action: rbac.ActionList},
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
