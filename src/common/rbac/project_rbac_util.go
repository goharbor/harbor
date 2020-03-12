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

package rbac

import (
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

var (
	// subresource policies for public project
	publicProjectPolicies = []*types.Policy{
		{Resource: ResourceSelf, Action: ActionRead},

		{Resource: ResourceLabel, Action: ActionRead},
		{Resource: ResourceLabel, Action: ActionList},

		{Resource: ResourceRepository, Action: ActionList},
		{Resource: ResourceRepository, Action: ActionPull},

		{Resource: ResourceRepositoryLabel, Action: ActionList},

		{Resource: ResourceRepositoryTag, Action: ActionRead},
		{Resource: ResourceRepositoryTag, Action: ActionList},

		{Resource: ResourceRepositoryTagLabel, Action: ActionList},

		{Resource: ResourceRepositoryTagVulnerability, Action: ActionList},

		{Resource: ResourceRepositoryTagManifest, Action: ActionRead},

		{Resource: ResourceHelmChart, Action: ActionRead},
		{Resource: ResourceHelmChart, Action: ActionList},

		{Resource: ResourceHelmChartVersion, Action: ActionRead},
		{Resource: ResourceHelmChartVersion, Action: ActionList},

		{Resource: ResourceScan, Action: ActionRead},
		{Resource: ResourceScanner, Action: ActionRead},

		{Resource: ResourceArtifact, Action: ActionRead},
		{Resource: ResourceArtifact, Action: ActionList},
		{Resource: ResourceArtifactAddition, Action: ActionRead},
	}

	// sub policies for the projects
	subPoliciesForProject = computeSubPoliciesForProject()
)

func getPoliciesForPublicProject(projectID int64) []*types.Policy {
	policies := []*types.Policy{}

	namespace := NewProjectNamespace(projectID)
	for _, policy := range publicProjectPolicies {
		policies = append(policies, &types.Policy{
			Resource: namespace.Resource(policy.Resource),
			Action:   policy.Action,
			Effect:   policy.Effect,
		})
	}

	return policies
}

// GetPoliciesOfProject returns all policies for namespace of the project
func GetPoliciesOfProject(projectID int64) []*types.Policy {
	policies := []*types.Policy{}

	namespace := NewProjectNamespace(projectID)
	for _, policy := range subPoliciesForProject {
		policies = append(policies, &types.Policy{
			Resource: namespace.Resource(policy.Resource),
			Action:   policy.Action,
			Effect:   policy.Effect,
		})
	}

	return policies
}

func computeSubPoliciesForProject() []*types.Policy {
	var results []*types.Policy

	mp := map[string]bool{}
	for _, policies := range rolePoliciesMap {
		for _, policy := range policies {
			if !mp[policy.String()] {
				results = append(results, policy)
				mp[policy.String()] = true
			}
		}
	}

	return results
}
