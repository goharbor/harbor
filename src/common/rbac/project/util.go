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
		{Resource: ResourceImage, Action: ActionPull},
	}

	// subresource policies for system admin visitor
	systemAdminProjectPolicies = []*rbac.Policy{
		{Resource: ResourceAll, Action: ActionAll},
	}
)

func policiesForPublicProject(namespace rbac.Namespace) []*rbac.Policy {
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

func policiesForSystemAdmin(namespace rbac.Namespace) []*rbac.Policy {
	policies := []*rbac.Policy{}

	for _, policy := range systemAdminProjectPolicies {
		policies = append(policies, &rbac.Policy{
			Resource: namespace.Resource(policy.Resource),
			Action:   policy.Action,
			Effect:   policy.Effect,
		})
	}

	return policies
}
