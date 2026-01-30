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
import { Role } from '../../../../../ng-swagger-gen/models/role';
import { Access } from '../../../../../ng-swagger-gen/models/access';
import { RolePermission } from '../../../../../ng-swagger-gen/models/role-permission';
import { Permission } from '../../../../../ng-swagger-gen/models/permission';

export interface FrontRole extends Role {
    permissionScope?: {
        coverAll?: boolean;
        access?: Array<Access>;
    };
}

export interface FrontAccess extends Access {
    checked?: boolean;
}

export enum PermissionsKinds {
    ROLE = 'project-role',
}

export enum Resource {
    REPO = 'repository',
    ARTIFACT = 'artifact',
}

export enum Action {
    PUSH = 'push',
    PULL = 'pull',
    READ = 'read',
    CREATE = 'create',
    LIST = 'list',
    STOP = 'stop',
    DELETE = 'delete',
}

export const NAMESPACE_ALL_PROJECTS: string = '*';

export const NAMESPACE_SYSTEM: string = '/';

export const ACTION_RESOURCE_I18N_MAP = {
    push: 'ROLE.PUSH_AND_PULL', // push permission contains pull permission
    pull: 'ROLE.PULL',
    read: 'ROLE.READ',
    create: 'ROLE.CREATE',
    delete: 'ROLE.DELETE',
    scan: 'ROLE.SCAN',
    stop: 'ROLE.STOP',
    list: 'ROLE.LIST',
    update: 'ROLE.UPDATE',
    'audit-log': 'ROLE.AUDIT_LOG',
    'preheat-instance': 'ROLE.PREHEAT_INSTANCE',
    project: 'ROLE.PROJECT',
    'replication-policy': 'ROLE.REPLICATION_POLICY',
    replication: 'ROLE.REPLICATION',
    'replication-adapter': 'ROLE.REPLICATION_ADAPTER',
    registry: 'ROLE.REGISTRY',
    'scan-all': 'ROLE.SCAN_ALL',
    'system-volumes': 'ROLE.SYSTEM_VOLUMES',
    'garbage-collection': 'ROLE.GARBAGE_COLLECTION',
    'purge-audit': 'ROLE.PURGE_AUDIT',
    'jobservice-monitor': 'ROLE.JOBSERVICE_MONITOR',
    'tag-retention': 'ROLE.TAG_RETENTION',
    scanner: 'ROLE.SCANNER',
    label: 'ROLE.LABEL',
    'export-cve': 'ROLE.EXPORT_CVE',
    'security-hub': 'ROLE.SECURITY_HUB',
    catalog: 'ROLE.CATALOG',
    metadata: 'ROLE.METADATA',
    repository: 'ROLE.REPOSITORY',
    artifact: 'ROLE.ARTIFACT',
    tag: 'ROLE.TAG',
    accessory: 'ROLE.ACCESSORY',
    'artifact-addition': 'ROLE.ARTIFACT_ADDITION',
    'artifact-label': 'ROLE.ARTIFACT_LABEL',
    'preheat-policy': 'ROLE.PREHEAT_POLICY',
    'immutable-tag': 'ROLE.IMMUTABLE_TAG',
    log: 'ROLE.LOG',
    'notification-policy': 'ROLE.NOTIFICATION_POLICY',
    quota: 'ROLE.QUOTA',
    sbom: 'ROLE.SBOM',
    role: 'ROLE.ROLE',
    user: 'ROLE.USER',
    'user-group': 'ROLE.GROUP',
    'ldap-user': 'ROLE.LDAPUSER',
    member: 'ROLE.MEMBER',
};

export function convertKey(key: string) {
    return ACTION_RESOURCE_I18N_MAP[key] ? ACTION_RESOURCE_I18N_MAP[key] : key;
}

export enum ExpirationType {
    DEFAULT = 'default',
    DAYS = 'days',
    NEVER = 'never',
}

export function onlyHasPushPermission(access: Access[]): boolean {
    if (access && access.length) {
        let hasPushPermission: boolean = false;
        let hasPullPermission: boolean = false;
        access.forEach(item => {
            if (
                item.action === Action.PUSH &&
                item.resource === Resource.REPO
            ) {
                hasPushPermission = true;
            }
            if (
                item.action === Action.PULL &&
                item.resource === Resource.REPO
            ) {
                hasPullPermission = true;
            }
        });
        if (hasPushPermission && !hasPullPermission) {
            return true;
        }
    }
    return false;
}



export function isCandidate(
    candidatePermissions: Permission[],
    permission: Access
): boolean {
    if (candidatePermissions?.length) {
        for (let i = 0; i < candidatePermissions.length; i++) {
            if (
                candidatePermissions[i].resource === permission.resource &&
                candidatePermissions[i].action === permission.action
            ) {
                return true;
            }
        }
    }
    return false;
}

export function hasPermission(
    permissions: Access[],
    permission: Access
): boolean {
    if (permissions?.length) {
        for (let i = 0; i < permissions.length; i++) {
            if (
                permissions[i].resource === permission.resource &&
                permissions[i].action === permission.action
            ) {
                return true;
            }
        }
    }
    return false;
}

export const NEW_EMPTY_ROLE: Role = {
    permissions: [
        {
            access: [],
        },
    ],
};

export function getRoleAccess(r: Role): Access[] {
    let systemPermissions: RolePermission[] = [];
    systemPermissions = r.permissions;
/*    if (r?.permissions?.length) {
        systemPermissions = r.permissions.filter(
            item => item.kind === PermissionsKinds.ROLE
        );
    }
*/
    if (systemPermissions?.length) {
        const map = {};
        systemPermissions.forEach(p => {
            if (p?.access?.length) {
                p.access.forEach(item => {
                    map[`${item.resource}@${item.action}`] = item;
                });
            }
        });
        return Object.values(map);
    }
    return [];
}
