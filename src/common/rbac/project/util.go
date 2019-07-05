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
	"github.com/goharbor/harbor/src/common/rbac"
)

var (
	// subresource policies for public project
	publicProjectPolicies = []*rbac.Policy{
		{Resource: rbac.ResourceSelf, Action: rbac.ActionRead},

		{Resource: rbac.ResourceLabel, Action: rbac.ActionRead},
		{Resource: rbac.ResourceLabel, Action: rbac.ActionList},

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
	}

	// all policies for the projects
	allPolicies = []*rbac.Policy{
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

		{Resource: rbac.ResourceReplication, Action: rbac.ActionList},
		{Resource: rbac.ResourceReplication, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceReplication, Action: rbac.ActionRead},
		{Resource: rbac.ResourceReplication, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceReplication, Action: rbac.ActionDelete},

		{Resource: rbac.ResourceReplicationJob, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceReplicationJob, Action: rbac.ActionRead},
		{Resource: rbac.ResourceReplicationJob, Action: rbac.ActionList},

		{Resource: rbac.ResourceReplicationExecution, Action: rbac.ActionRead},
		{Resource: rbac.ResourceReplicationExecution, Action: rbac.ActionList},
		{Resource: rbac.ResourceReplicationExecution, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceReplicationExecution, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceReplicationExecution, Action: rbac.ActionDelete},

		{Resource: rbac.ResourceReplicationTask, Action: rbac.ActionRead},
		{Resource: rbac.ResourceReplicationTask, Action: rbac.ActionList},
		{Resource: rbac.ResourceReplicationTask, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceReplicationTask, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceReplicationTask, Action: rbac.ActionDelete},

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
		{Resource: rbac.ResourceConfiguration, Action: rbac.ActionUpdate},

		{Resource: rbac.ResourceRobot, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceRobot, Action: rbac.ActionRead},
		{Resource: rbac.ResourceRobot, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceRobot, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceRobot, Action: rbac.ActionList},
	}
)

// PoliciesForPublicProject ...
func PoliciesForPublicProject(namespace rbac.Namespace) []*rbac.Policy {
	policies := []*rbac.Policy{}

	for _, policy := range publicProjectPolicies {
		policies = append(policies, &rbac.Policy{
			Resource: namespace.Resource(policy.Resource),
			Action:   policy.Action,
			Effect:   policy.Effect,
		})
	}

	return policies
}

// GetAllPolicies returns all policies for namespace of the project
func GetAllPolicies(namespace rbac.Namespace) []*rbac.Policy {
	policies := []*rbac.Policy{}

	for _, policy := range allPolicies {
		policies = append(policies, &rbac.Policy{
			Resource: namespace.Resource(policy.Resource),
			Action:   policy.Action,
			Effect:   policy.Effect,
		})
	}

	return policies
}
