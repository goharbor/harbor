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

package policy

import (
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// shippedRoles is what the roles Harbor ships grant, as it stood in
// common/rbac/project before they became rows like every other role.
//
// Nothing asks this map a question. It is read once, the first time a Harbor
// starts against a database that has no grants for these roles, to write the
// rows that answer from then on. An operator who edits one afterwards keeps
// that edit: seeding skips a role that already has grants.
var shippedRoles = map[string][]*types.Policy{
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
		{Resource: rbac.ResourceMetadata, Action: rbac.ActionList},

		{Resource: rbac.ResourceLog, Action: rbac.ActionList},

		{Resource: rbac.ResourceLabel, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceLabel, Action: rbac.ActionRead},
		{Resource: rbac.ResourceLabel, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceLabel, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceLabel, Action: rbac.ActionList},

		{Resource: rbac.ResourceQuota, Action: rbac.ActionRead},

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

		{Resource: rbac.ResourceScan, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceScan, Action: rbac.ActionRead},
		{Resource: rbac.ResourceScan, Action: rbac.ActionStop},
		{Resource: rbac.ResourceSBOM, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceSBOM, Action: rbac.ActionStop},
		{Resource: rbac.ResourceSBOM, Action: rbac.ActionRead},

		{Resource: rbac.ResourceScanner, Action: rbac.ActionRead},
		{Resource: rbac.ResourceScanner, Action: rbac.ActionCreate},

		{Resource: rbac.ResourceArtifact, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceArtifact, Action: rbac.ActionRead},
		{Resource: rbac.ResourceArtifact, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceArtifact, Action: rbac.ActionList},
		{Resource: rbac.ResourceArtifactAddition, Action: rbac.ActionRead},

		{Resource: rbac.ResourceTag, Action: rbac.ActionList},
		{Resource: rbac.ResourceTag, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceTag, Action: rbac.ActionDelete},

		{Resource: rbac.ResourceAccessory, Action: rbac.ActionList},

		{Resource: rbac.ResourceArtifactLabel, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceArtifactLabel, Action: rbac.ActionDelete},

		{Resource: rbac.ResourcePreatPolicy, Action: rbac.ActionCreate},
		{Resource: rbac.ResourcePreatPolicy, Action: rbac.ActionRead},
		{Resource: rbac.ResourcePreatPolicy, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourcePreatPolicy, Action: rbac.ActionDelete},
		{Resource: rbac.ResourcePreatPolicy, Action: rbac.ActionList},

		{Resource: rbac.ResourceExportCVE, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceExportCVE, Action: rbac.ActionRead},
		{Resource: rbac.ResourceExportCVE, Action: rbac.ActionList},
	},

	"maintainer": {
		{Resource: rbac.ResourceSelf, Action: rbac.ActionRead},

		{Resource: rbac.ResourceMember, Action: rbac.ActionRead},
		{Resource: rbac.ResourceMember, Action: rbac.ActionList},

		{Resource: rbac.ResourceMetadata, Action: rbac.ActionRead},

		{Resource: rbac.ResourceQuota, Action: rbac.ActionRead},

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

		{Resource: rbac.ResourceAccessory, Action: rbac.ActionList},

		{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceImmutableTag, Action: rbac.ActionList},

		{Resource: rbac.ResourceConfiguration, Action: rbac.ActionRead},

		{Resource: rbac.ResourceRobot, Action: rbac.ActionRead},
		{Resource: rbac.ResourceRobot, Action: rbac.ActionList},

		{Resource: rbac.ResourceNotificationPolicy, Action: rbac.ActionRead},
		{Resource: rbac.ResourceNotificationPolicy, Action: rbac.ActionList},

		{Resource: rbac.ResourceScan, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceScan, Action: rbac.ActionRead},
		{Resource: rbac.ResourceScan, Action: rbac.ActionStop},
		{Resource: rbac.ResourceSBOM, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceSBOM, Action: rbac.ActionStop},
		{Resource: rbac.ResourceSBOM, Action: rbac.ActionRead},

		{Resource: rbac.ResourceScanner, Action: rbac.ActionRead},

		{Resource: rbac.ResourceArtifact, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceArtifact, Action: rbac.ActionRead},
		{Resource: rbac.ResourceArtifact, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceArtifact, Action: rbac.ActionList},
		{Resource: rbac.ResourceArtifactAddition, Action: rbac.ActionRead},

		{Resource: rbac.ResourceTag, Action: rbac.ActionList},
		{Resource: rbac.ResourceTag, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceTag, Action: rbac.ActionDelete},

		{Resource: rbac.ResourceArtifactLabel, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceArtifactLabel, Action: rbac.ActionDelete},

		{Resource: rbac.ResourceExportCVE, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceExportCVE, Action: rbac.ActionRead},
		{Resource: rbac.ResourceExportCVE, Action: rbac.ActionList},
	},

	"developer": {
		{Resource: rbac.ResourceSelf, Action: rbac.ActionRead},

		{Resource: rbac.ResourceMember, Action: rbac.ActionRead},
		{Resource: rbac.ResourceMember, Action: rbac.ActionList},

		{Resource: rbac.ResourceLabel, Action: rbac.ActionRead},
		{Resource: rbac.ResourceLabel, Action: rbac.ActionList},

		{Resource: rbac.ResourceQuota, Action: rbac.ActionRead},

		{Resource: rbac.ResourceRepository, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceRepository, Action: rbac.ActionRead},
		{Resource: rbac.ResourceRepository, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceRepository, Action: rbac.ActionList},
		{Resource: rbac.ResourceRepository, Action: rbac.ActionPush},
		{Resource: rbac.ResourceRepository, Action: rbac.ActionPull},

		{Resource: rbac.ResourceTagRetention, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceTagRetention, Action: rbac.ActionRead},
		{Resource: rbac.ResourceTagRetention, Action: rbac.ActionUpdate},
		{Resource: rbac.ResourceTagRetention, Action: rbac.ActionDelete},
		{Resource: rbac.ResourceTagRetention, Action: rbac.ActionList},
		{Resource: rbac.ResourceTagRetention, Action: rbac.ActionOperate},

		{Resource: rbac.ResourceConfiguration, Action: rbac.ActionRead},

		{Resource: rbac.ResourceRobot, Action: rbac.ActionRead},
		{Resource: rbac.ResourceRobot, Action: rbac.ActionList},

		{Resource: rbac.ResourceScan, Action: rbac.ActionRead},
		{Resource: rbac.ResourceSBOM, Action: rbac.ActionRead},

		{Resource: rbac.ResourceScanner, Action: rbac.ActionRead},

		{Resource: rbac.ResourceArtifact, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceArtifact, Action: rbac.ActionRead},
		{Resource: rbac.ResourceArtifact, Action: rbac.ActionList},
		{Resource: rbac.ResourceArtifactAddition, Action: rbac.ActionRead},

		{Resource: rbac.ResourceTag, Action: rbac.ActionList},
		{Resource: rbac.ResourceTag, Action: rbac.ActionCreate},

		{Resource: rbac.ResourceAccessory, Action: rbac.ActionList},

		{Resource: rbac.ResourceArtifactLabel, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceArtifactLabel, Action: rbac.ActionDelete},

		{Resource: rbac.ResourceExportCVE, Action: rbac.ActionCreate},
		{Resource: rbac.ResourceExportCVE, Action: rbac.ActionRead},
		{Resource: rbac.ResourceExportCVE, Action: rbac.ActionList},
	},

	"guest": {
		{Resource: rbac.ResourceSelf, Action: rbac.ActionRead},

		{Resource: rbac.ResourceMember, Action: rbac.ActionRead},
		{Resource: rbac.ResourceMember, Action: rbac.ActionList},

		{Resource: rbac.ResourceLabel, Action: rbac.ActionRead},
		{Resource: rbac.ResourceLabel, Action: rbac.ActionList},

		{Resource: rbac.ResourceQuota, Action: rbac.ActionRead},

		{Resource: rbac.ResourceRepository, Action: rbac.ActionRead},
		{Resource: rbac.ResourceRepository, Action: rbac.ActionList},
		{Resource: rbac.ResourceRepository, Action: rbac.ActionPull},

		{Resource: rbac.ResourceConfiguration, Action: rbac.ActionRead},

		{Resource: rbac.ResourceRobot, Action: rbac.ActionRead},
		{Resource: rbac.ResourceRobot, Action: rbac.ActionList},

		{Resource: rbac.ResourceScan, Action: rbac.ActionRead},
		{Resource: rbac.ResourceSBOM, Action: rbac.ActionRead},

		{Resource: rbac.ResourceScanner, Action: rbac.ActionRead},

		{Resource: rbac.ResourceTag, Action: rbac.ActionList},
		{Resource: rbac.ResourceAccessory, Action: rbac.ActionList},

		{Resource: rbac.ResourceArtifact, Action: rbac.ActionRead},
		{Resource: rbac.ResourceArtifact, Action: rbac.ActionList},
		{Resource: rbac.ResourceArtifactAddition, Action: rbac.ActionRead},
	},

	"limitedGuest": {
		{Resource: rbac.ResourceSelf, Action: rbac.ActionRead},

		{Resource: rbac.ResourceQuota, Action: rbac.ActionRead},

		{Resource: rbac.ResourceRepository, Action: rbac.ActionList},
		{Resource: rbac.ResourceRepository, Action: rbac.ActionRead},
		{Resource: rbac.ResourceRepository, Action: rbac.ActionPull},

		{Resource: rbac.ResourceConfiguration, Action: rbac.ActionRead},

		{Resource: rbac.ResourceScan, Action: rbac.ActionRead},
		{Resource: rbac.ResourceSBOM, Action: rbac.ActionRead},

		{Resource: rbac.ResourceScanner, Action: rbac.ActionRead},

		{Resource: rbac.ResourceTag, Action: rbac.ActionList},
		{Resource: rbac.ResourceAccessory, Action: rbac.ActionList},

		{Resource: rbac.ResourceArtifact, Action: rbac.ActionRead},
		{Resource: rbac.ResourceArtifact, Action: rbac.ActionList},
		{Resource: rbac.ResourceArtifactAddition, Action: rbac.ActionRead},
	},
}

// ProjectPolicies is every permission a project role can hold: the union of
// what the roles Harbor ships grant between them.
//
// It is the vocabulary, not a decision. Nothing on the permission path reads
// it; the APIs that report what is grantable in a project do.
func ProjectPolicies() []*types.Policy {
	seen := map[string]bool{}
	var out []*types.Policy
	for _, policies := range shippedRoles {
		for _, p := range policies {
			if seen[p.String()] {
				continue
			}
			seen[p.String()] = true
			out = append(out, p)
		}
	}
	return out
}
