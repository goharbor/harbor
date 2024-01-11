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

import "github.com/goharbor/harbor/src/pkg/permission/types"

// const action variables
const (
	ActionAll = Action("*") // action match any other actions

	ActionPull = Action("pull") // pull repository tag
	ActionPush = Action("push") // push repository tag

	// create, read, update, delete, list actions compatible with restful api methods
	ActionCreate = Action("create")
	ActionRead   = Action("read")
	ActionUpdate = Action("update")
	ActionDelete = Action("delete")
	ActionList   = Action("list")

	ActionOperate     = Action("operate")
	ActionScannerPull = Action("scanner-pull") // for robot account created by scanner to pull image, bypass the policy check
	ActionStop        = Action("stop")         // for stop scan/scan-all execution
)

// const resource variables
const (
	ResourceAll                = Resource("*")             // resource match any other resources
	ResourceConfiguration      = Resource("configuration") // project configuration compatible for portal only
	ResourceLabel              = Resource("label")
	ResourceLog                = Resource("log")
	ResourceLdapUser           = Resource("ldap-user")
	ResourceMember             = Resource("member")
	ResourceMetadata           = Resource("metadata")
	ResourceQuota              = Resource("quota")
	ResourceRepository         = Resource("repository")
	ResourceTagRetention       = Resource("tag-retention")
	ResourceImmutableTag       = Resource("immutable-tag")
	ResourceRobot              = Resource("robot")
	ResourceNotificationPolicy = Resource("notification-policy")
	ResourceScan               = Resource("scan")
	ResourceScanner            = Resource("scanner")
	ResourceArtifact           = Resource("artifact")
	ResourceTag                = Resource("tag")
	ResourceAccessory          = Resource("accessory")
	ResourceArtifactAddition   = Resource("artifact-addition")
	ResourceArtifactLabel      = Resource("artifact-label")
	ResourcePreatPolicy        = Resource("preheat-policy")
	ResourcePreatInstance      = Resource("preheat-instance")
	ResourceSelf               = Resource("") // subresource for self

	ResourceAuditLog           = Resource("audit-log")
	ResourceCatalog            = Resource("catalog")
	ResourceProject            = Resource("project")
	ResourceUser               = Resource("user")
	ResourceUserGroup          = Resource("user-group")
	ResourceRegistry           = Resource("registry")
	ResourceReplication        = Resource("replication")
	ResourceDistribution       = Resource("distribution")
	ResourceGarbageCollection  = Resource("garbage-collection")
	ResourceReplicationAdapter = Resource("replication-adapter")
	ResourceReplicationPolicy  = Resource("replication-policy")
	ResourceScanAll            = Resource("scan-all")
	ResourceSystemVolumes      = Resource("system-volumes")
	ResourcePurgeAuditLog      = Resource("purge-audit")
	ResourceExportCVE          = Resource("export-cve")
	ResourceJobServiceMonitor  = Resource("jobservice-monitor")
	ResourceSecurityHub        = Resource("security-hub")
)

var (
	PoliciesMap = map[string][]*types.Policy{
		"System": {
			{Resource: ResourceAuditLog, Action: ActionList},

			{Resource: ResourcePreatInstance, Action: ActionRead},
			{Resource: ResourcePreatInstance, Action: ActionCreate},
			{Resource: ResourcePreatInstance, Action: ActionDelete},
			{Resource: ResourcePreatInstance, Action: ActionList},
			{Resource: ResourcePreatInstance, Action: ActionUpdate},

			{Resource: ResourceProject, Action: ActionList},
			{Resource: ResourceProject, Action: ActionCreate},

			{Resource: ResourceReplicationPolicy, Action: ActionRead},
			{Resource: ResourceReplicationPolicy, Action: ActionCreate},
			{Resource: ResourceReplicationPolicy, Action: ActionDelete},
			{Resource: ResourceReplicationPolicy, Action: ActionList},
			{Resource: ResourceReplicationPolicy, Action: ActionUpdate},

			{Resource: ResourceReplication, Action: ActionRead},
			{Resource: ResourceReplication, Action: ActionCreate},
			{Resource: ResourceReplication, Action: ActionList},

			{Resource: ResourceReplicationAdapter, Action: ActionList},

			{Resource: ResourceRegistry, Action: ActionRead},
			{Resource: ResourceRegistry, Action: ActionCreate},
			{Resource: ResourceRegistry, Action: ActionDelete},
			{Resource: ResourceRegistry, Action: ActionList},
			{Resource: ResourceRegistry, Action: ActionUpdate},

			{Resource: ResourceScanAll, Action: ActionRead},
			{Resource: ResourceScanAll, Action: ActionUpdate},
			{Resource: ResourceScanAll, Action: ActionStop},
			{Resource: ResourceScanAll, Action: ActionCreate},

			{Resource: ResourceSystemVolumes, Action: ActionRead},

			{Resource: ResourceGarbageCollection, Action: ActionRead},
			{Resource: ResourceGarbageCollection, Action: ActionCreate},
			{Resource: ResourceGarbageCollection, Action: ActionList},
			{Resource: ResourceGarbageCollection, Action: ActionUpdate},
			{Resource: ResourceGarbageCollection, Action: ActionStop},

			{Resource: ResourcePurgeAuditLog, Action: ActionRead},
			{Resource: ResourcePurgeAuditLog, Action: ActionCreate},
			{Resource: ResourcePurgeAuditLog, Action: ActionList},
			{Resource: ResourcePurgeAuditLog, Action: ActionUpdate},
			{Resource: ResourcePurgeAuditLog, Action: ActionStop},

			{Resource: ResourceJobServiceMonitor, Action: ActionList},
			{Resource: ResourceJobServiceMonitor, Action: ActionStop},

			{Resource: ResourceScanner, Action: ActionRead},
			{Resource: ResourceScanner, Action: ActionCreate},
			{Resource: ResourceScanner, Action: ActionDelete},
			{Resource: ResourceScanner, Action: ActionList},
			{Resource: ResourceScanner, Action: ActionUpdate},

			{Resource: ResourceLabel, Action: ActionRead},
			{Resource: ResourceLabel, Action: ActionCreate},
			{Resource: ResourceLabel, Action: ActionDelete},
			{Resource: ResourceLabel, Action: ActionUpdate},

			{Resource: ResourceSecurityHub, Action: ActionRead},
			{Resource: ResourceSecurityHub, Action: ActionList},

			{Resource: ResourceCatalog, Action: ActionRead},

			{Resource: ResourceQuota, Action: ActionRead},
			{Resource: ResourceQuota, Action: ActionList},
		},
		"Project": {
			{Resource: ResourceLog, Action: ActionList},

			{Resource: ResourceProject, Action: ActionRead},
			{Resource: ResourceProject, Action: ActionDelete},
			{Resource: ResourceProject, Action: ActionUpdate},

			{Resource: ResourceMetadata, Action: ActionRead},
			{Resource: ResourceMetadata, Action: ActionCreate},
			{Resource: ResourceMetadata, Action: ActionDelete},
			{Resource: ResourceMetadata, Action: ActionList},
			{Resource: ResourceMetadata, Action: ActionUpdate},

			{Resource: ResourceRepository, Action: ActionRead},
			{Resource: ResourceRepository, Action: ActionUpdate},
			{Resource: ResourceRepository, Action: ActionDelete},
			{Resource: ResourceRepository, Action: ActionList},
			{Resource: ResourceRepository, Action: ActionPull},
			{Resource: ResourceRepository, Action: ActionPush},

			{Resource: ResourceArtifact, Action: ActionRead},
			{Resource: ResourceArtifact, Action: ActionCreate},
			{Resource: ResourceArtifact, Action: ActionList},
			{Resource: ResourceArtifact, Action: ActionDelete},

			{Resource: ResourceScan, Action: ActionCreate},
			{Resource: ResourceScan, Action: ActionRead},
			{Resource: ResourceScan, Action: ActionStop},

			{Resource: ResourceTag, Action: ActionCreate},
			{Resource: ResourceTag, Action: ActionList},
			{Resource: ResourceTag, Action: ActionDelete},

			{Resource: ResourceAccessory, Action: ActionList},

			{Resource: ResourceArtifactAddition, Action: ActionRead},

			{Resource: ResourceArtifactLabel, Action: ActionCreate},
			{Resource: ResourceArtifactLabel, Action: ActionDelete},

			{Resource: ResourceScanner, Action: ActionCreate},
			{Resource: ResourceScanner, Action: ActionRead},

			{Resource: ResourcePreatPolicy, Action: ActionRead},
			{Resource: ResourcePreatPolicy, Action: ActionCreate},
			{Resource: ResourcePreatPolicy, Action: ActionDelete},
			{Resource: ResourcePreatPolicy, Action: ActionList},
			{Resource: ResourcePreatPolicy, Action: ActionUpdate},

			{Resource: ResourceImmutableTag, Action: ActionCreate},
			{Resource: ResourceImmutableTag, Action: ActionDelete},
			{Resource: ResourceImmutableTag, Action: ActionList},
			{Resource: ResourceImmutableTag, Action: ActionUpdate},

			{Resource: ResourceNotificationPolicy, Action: ActionRead},
			{Resource: ResourceNotificationPolicy, Action: ActionCreate},
			{Resource: ResourceNotificationPolicy, Action: ActionDelete},
			{Resource: ResourceNotificationPolicy, Action: ActionList},
			{Resource: ResourceNotificationPolicy, Action: ActionUpdate},

			{Resource: ResourceTagRetention, Action: ActionRead},
			{Resource: ResourceTagRetention, Action: ActionCreate},
			{Resource: ResourceTagRetention, Action: ActionDelete},
			{Resource: ResourceTagRetention, Action: ActionList},
			{Resource: ResourceTagRetention, Action: ActionUpdate},

			{Resource: ResourceLabel, Action: ActionRead},
			{Resource: ResourceLabel, Action: ActionCreate},
			{Resource: ResourceLabel, Action: ActionDelete},
			{Resource: ResourceLabel, Action: ActionList},
			{Resource: ResourceLabel, Action: ActionUpdate},

			{Resource: ResourceQuota, Action: ActionRead},
		},
	}
)
