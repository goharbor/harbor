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

package system

import (
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

var (
	policies = []*types.Policy{
		{Resource: rbac.ResourceCatalog, Action: rbac.ActionRead},

		{Resource: rbac.ResourceAuditLog, Action: rbac.ActionList},

		{Resource: rbac.ResourceProject, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceProject, Action: rbac.ActionRead},
		{Resource: rbac.ResourceProject, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceProject, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceProject, Action: rbac.ActionList},

		{Resource: rbac.ResourceUser, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceUser, Action: rbac.ActionRead},
		{Resource: rbac.ResourceUser, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceUser, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceUser, Action: rbac.ActionList},

		{Resource: rbac.ResourceUserGroup, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceUserGroup, Action: rbac.ActionRead},
		{Resource: rbac.ResourceUserGroup, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceUserGroup, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceUserGroup, Action: rbac.ActionList},

		{Resource: rbac.ResourceRegistry, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceRegistry, Action: rbac.ActionRead},
		{Resource: rbac.ResourceRegistry, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceRegistry, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceRegistry, Action: rbac.ActionList},

		{Resource: rbac.ResourceReplication, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceReplication, Action: rbac.ActionRead},
		{Resource: rbac.ResourceReplication, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceReplication, Action: rbac.ActionList},
		{Resource: rbac.ResourceReplication, Action: rbac.ActionDelete},

		{Resource: rbac.ResourceDistribution, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceDistribution, Action: rbac.ActionRead},
		{Resource: rbac.ResourceDistribution, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceDistribution, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceDistribution, Action: rbac.ActionList},

		{Resource: rbac.ResourceGarbageCollection, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceGarbageCollection, Action: rbac.ActionRead},
		{Resource: rbac.ResourceGarbageCollection, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceGarbageCollection, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceGarbageCollection, Action: rbac.ActionList},

		{Resource: rbac.ResourceScanAll, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceScanAll, Action: rbac.ActionRead},
		{Resource: rbac.ResourceScanAll, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceScanAll, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceScanAll, Action: rbac.ActionList},
		{Resource: rbac.ResourceScanAll, Action: rbac.ActionStop},

		{Resource: rbac.ResourceSystemVolumes, Action: rbac.ActionRead},

		{Resource: rbac.ResourceLdapUser, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceLdapUser, Action: rbac.ActionList},
		{Resource: rbac.ResourceConfiguration, Action: rbac.ActionRead},
		{Resource: rbac.ResourceConfiguration, Action: rbac.ActionUpdate},

		{Resource: rbac.ResourceJobServiceMonitor, Action: rbac.ActionRead},
		{Resource: rbac.ResourceJobServiceMonitor, Action: rbac.ActionList},
		{Resource: rbac.ResourceJobServiceMonitor, Action: rbac.ActionStop},

		{Resource: rbac.ResourceSecurityHub, Action: rbac.ActionRead},
		{Resource: rbac.ResourceSecurityHub, Action: rbac.ActionList},
	}
)
