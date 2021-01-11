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
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

var (
	// subresource policies for public project
	publicProjectPolicies = []*types.Policy{
		{Resource: rbac.ResourceSelf, Action: rbac.ActionRead},

		{Resource: rbac.ResourceLabel, Action: rbac.ActionRead},
		{Resource: rbac.ResourceLabel, Action: rbac.ActionList},

		{Resource: rbac.ResourceRepository, Action: rbac.ActionList},
		{Resource: rbac.ResourceRepository, Action: rbac.ActionPull},

		{Resource: rbac.ResourceHelmChart, Action: rbac.ActionRead},
		{Resource: rbac.ResourceHelmChart, Action: rbac.ActionList},

		{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionRead},
		{Resource: rbac.ResourceHelmChartVersion, Action: rbac.ActionList},

		{Resource: rbac.ResourceScan, Action: rbac.ActionRead},
		{Resource: rbac.ResourceScanner, Action: rbac.ActionRead},

		{Resource: rbac.ResourceTag, Action: rbac.ActionList},

		{Resource: rbac.ResourceArtifact, Action: rbac.ActionRead},
		{Resource: rbac.ResourceArtifact, Action: rbac.ActionList},
		{Resource: rbac.ResourceArtifactAddition, Action: rbac.ActionRead},
	}

	// sub policies for the projects
	subPoliciesForProject = computeSubPoliciesForProject()
)

func getPoliciesForPublicProject(projectID int64) []*types.Policy {
	policies := []*types.Policy{}

	namespace := NewNamespace(projectID)
	for _, policy := range publicProjectPolicies {
		policies = append(policies, &types.Policy{
			Resource: namespace.Resource(policy.Resource),
			Action:   policy.Action,
			Effect:   policy.Effect,
		})
	}

	return policies
}

// GetPoliciesOfProject returns all policies for projectNamespace of the project
func GetPoliciesOfProject(projectID int64) []*types.Policy {
	policies := []*types.Policy{}

	namespace := NewNamespace(projectID)
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
