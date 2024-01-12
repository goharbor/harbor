import { Robot } from '../../../../../ng-swagger-gen/models/robot';
import { Access } from '../../../../../ng-swagger-gen/models/access';
import { RobotPermission } from '../../../../../ng-swagger-gen/models/robot-permission';
import { Permission } from '../../../../../ng-swagger-gen/models/permission';

export interface FrontRobot extends Robot {
    permissionScope?: {
        coverAll?: boolean;
        access?: Array<Access>;
    };
}

export interface FrontAccess extends Access {
    checked?: boolean;
}

export enum PermissionsKinds {
    PROJECT = 'project',
    SYSTEM = 'system',
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
    push: 'SYSTEM_ROBOT.PUSH_AND_PULL', // push permission contains pull permission
    pull: 'ROBOT_ACCOUNT.PULL',
    read: 'SYSTEM_ROBOT.READ',
    create: 'SYSTEM_ROBOT.CREATE',
    delete: 'SYSTEM_ROBOT.DELETE',
    scan: 'SYSTEM_ROBOT.SCAN',
    stop: 'SYSTEM_ROBOT.STOP',
    list: 'SYSTEM_ROBOT.LIST',
    update: 'ROBOT_ACCOUNT.UPDATE',
    'audit-log': 'ROBOT_ACCOUNT.AUDIT_LOG',
    'preheat-instance': 'ROBOT_ACCOUNT.PREHEAT_INSTANCE',
    project: 'ROBOT_ACCOUNT.PROJECT',
    'replication-policy': 'ROBOT_ACCOUNT.REPLICATION_POLICY',
    replication: 'ROBOT_ACCOUNT.REPLICATION',
    'replication-adapter': 'ROBOT_ACCOUNT.REPLICATION_ADAPTER',
    registry: 'ROBOT_ACCOUNT.REGISTRY',
    'scan-all': 'ROBOT_ACCOUNT.SCAN_ALL',
    'system-volumes': 'ROBOT_ACCOUNT.SYSTEM_VOLUMES',
    'garbage-collection': 'ROBOT_ACCOUNT.GARBAGE_COLLECTION',
    'purge-audit': 'ROBOT_ACCOUNT.PURGE_AUDIT',
    'jobservice-monitor': 'ROBOT_ACCOUNT.JOBSERVICE_MONITOR',
    'tag-retention': 'ROBOT_ACCOUNT.TAG_RETENTION',
    scanner: 'ROBOT_ACCOUNT.SCANNER',
    label: 'ROBOT_ACCOUNT.LABEL',
    'export-cve': 'ROBOT_ACCOUNT.EXPORT_CVE',
    'security-hub': 'ROBOT_ACCOUNT.SECURITY_HUB',
    catalog: 'ROBOT_ACCOUNT.CATALOG',
    metadata: 'ROBOT_ACCOUNT.METADATA',
    repository: 'ROBOT_ACCOUNT.REPOSITORY',
    artifact: 'ROBOT_ACCOUNT.ARTIFACT',
    tag: 'ROBOT_ACCOUNT.TAG',
    accessory: 'ROBOT_ACCOUNT.ACCESSORY',
    'artifact-addition': 'ROBOT_ACCOUNT.ARTIFACT_ADDITION',
    'artifact-label': 'ROBOT_ACCOUNT.ARTIFACT_LABEL',
    'preheat-policy': 'ROBOT_ACCOUNT.PREHEAT_POLICY',
    'immutable-tag': 'ROBOT_ACCOUNT.IMMUTABLE_TAG',
    log: 'ROBOT_ACCOUNT.LOG',
    'notification-policy': 'ROBOT_ACCOUNT.NOTIFICATION_POLICY',
    quota: 'ROBOT_ACCOUNT.QUOTA',
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

export enum RobotTimeRemainColor {
    GREEN = 'green',
    WARNING = 'yellow',
    EXPIRED = 'red',
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

export const NEW_EMPTY_ROBOT: Robot = {
    permissions: [
        {
            access: [],
        },
    ],
};

export function getSystemAccess(r: Robot): Access[] {
    let systemPermissions: RobotPermission[] = [];
    if (r?.permissions?.length) {
        systemPermissions = r.permissions.filter(
            item => item.kind === PermissionsKinds.SYSTEM
        );
    }
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
