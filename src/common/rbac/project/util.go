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

		{Resource: rbac.ResourceScan, Action: rbac.ActionRead},
		{Resource: rbac.ResourceScanner, Action: rbac.ActionRead},
	}

	// all policies for the projects
	allPolicies = computeAllPolicies()
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

func computeAllPolicies() []*rbac.Policy {
	var results []*rbac.Policy

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
