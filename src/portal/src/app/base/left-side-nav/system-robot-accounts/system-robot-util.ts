import { Robot } from '../../../../../ng-swagger-gen/models/robot';
import { Access } from '../../../../../ng-swagger-gen/models/access';
import { Project } from '../../../../../ng-swagger-gen/models/project';

export interface FrontRobot extends Robot {
    permissionScope?: {
        coverAll?: boolean;
        access?: Array<Access>;
    };
}

export interface FrontProjectForAdd extends Project {
    permissions?: Array<{
        kind?: string;
        namespace?: string;
        access?: Array<FrontAccess>;
    }>;
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

export const INITIAL_ACCESSES: FrontAccess[] = [
    {
        resource: 'repository',
        action: 'list',
        checked: true,
    },
    {
        resource: 'repository',
        action: 'pull',
        checked: true,
    },
    {
        resource: 'repository',
        action: 'push',
        checked: true,
    },
    {
        resource: 'repository',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'artifact',
        action: 'read',
        checked: true,
    },
    {
        resource: 'artifact',
        action: 'list',
        checked: true,
    },
    {
        resource: 'artifact',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'artifact-label',
        action: 'create',
        checked: true,
    },
    {
        resource: 'artifact-label',
        action: 'delete',
        checked: true,
    },
    
    {
        resource: 'tag',
        action: 'create',
        checked: true,
    },
    {
        resource: 'tag',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'tag',
        action: 'list',
        checked: true,
    },
    {
        resource: 'scan',
        action: 'create',
        checked: true,
    },
    {
        resource: 'scan',
        action: 'stop',
        checked: true,
    },
    //New Permissions to be seperated
    {
        resource: 'helm-chart',
        action: 'list',
        checked: true,
    },
    {
        resource: 'helm-chart',
        action: 'read',
        checked: true,
    },
    {
        resource: 'helm-chart',
        action: 'create',
        checked: true,
    },
    {
        resource: 'helm-chart',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'helm-chart-version',
        action: 'list',
        checked: true,
    },
    {
        resource: 'helm-chart-version',
        action: 'read',
        checked: true,
    },
    {
        resource: 'helm-chart-version',
        action: 'create',
        checked: true,
    },
    {
        resource: 'helm-chart-version',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'helm-chart-version-label',
        action: 'create',
        checked: true,
    },
    {
        resource: 'helm-chart-version-label',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'log',
        action: 'list',
        checked: true,
    },
    {
        resource: 'member',
        action: 'list',
        checked: true,
    },
    {
        resource: 'member',
        action: 'read',
        checked: true,
    },
    {
        resource: 'member',
        action: 'create',
        checked: true,
    },
    {
        resource: 'member',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'member',
        action: 'update',
        checked: true,
    },
    {
        resource: 'metadata',
        action: 'list',
        checked: true,
    },
    {
        resource: 'metadata',
        action: 'read',
        checked: true,
    },
    {
        resource: 'metadata',
        action: 'create',
        checked: true,
    },
    {
        resource: 'metadata',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'metadata',
        action: 'update',
        checked: true,
    },
    {
        resource: 'quota',
        action: 'read',
        checked: true,
    },
    {
        resource: 'tag-retention',
        action: 'list',
        checked: true,
    },
    {
        resource: 'tag-retention',
        action: 'read',
        checked: true,
    },
    {
        resource: 'tag-retention',
        action: 'create',
        checked: true,
    },
    {
        resource: 'tag-retention',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'tag-retention',
        action: 'update',
        checked: true,
    },
    {
        resource: 'immutable-tag',
        action: 'list',
        checked: true,
    },
    {
        resource: 'immutable-tag',
        action: 'create',
        checked: true,
    },
    {
        resource: 'immutable-tag',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'immutable-tag',
        action: 'update',
        checked: true,
    },
    {
        resource: 'robot',
        action: 'list',
        checked: true,
    },
    {
        resource: 'robot',
        action: 'read',
        checked: true,
    },
    {
        resource: 'robot',
        action: 'create',
        checked: true,
    },
    {
        resource: 'robot',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'robot',
        action: 'update',
        checked: true,
    },
    {
        resource: 'notification-policy',
        action: 'list',
        checked: true,
    },
    {
        resource: 'notification-policy',
        action: 'read',
        checked: true,
    },
    {
        resource: 'notification-policy',
        action: 'create',
        checked: true,
    },
    {
        resource: 'notification-policy',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'notification-policy',
        action: 'update',
        checked: true,
    },
    {
        resource: 'scanner',
        action: 'list',
        checked: true,
    },
    {
        resource: 'scanner',
        action: 'read',
        checked: true,
    },
    {
        resource: 'scanner',
        action: 'create',
        checked: true,
    },
    {
        resource: 'accessory',
        action: 'list',
        checked: true,
    },
    {
        resource: 'artifact-addition',
        action: 'read',
        checked: true,
    },
    {
        resource: 'preheat-policy',
        action: 'list',
        checked: true,
    },
    {
        resource: 'preheat-policy',
        action: 'read',
        checked: true,
    },
    {
        resource: 'preheat-policy',
        action: 'create',
        checked: true,
    },
    {
        resource: 'preheat-policy',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'preheat-policy',
        action: 'update',
        checked: true,
    },
    {
        resource: 'project',
        action: 'list',
        checked: true,
    },
    {
        resource: 'project',
        action: 'read',
        checked: true,
    },
    {
        resource: 'project',
        action: 'create',
        checked: true,
    },
    {
        resource: 'project',
        action: 'delete',
        checked: true,
    },
    {
        resource: 'project',
        action: 'update',
        checked: true,
    },
];

export const ACTION_RESOURCE_I18N_MAP = {
    push: 'SYSTEM_ROBOT.PUSH_AND_PULL', // push permission contains pull permission
    pull: 'ROBOT_ACCOUNT.PULL',
    read: 'SYSTEM_ROBOT.READ',
    create: 'SYSTEM_ROBOT.CREATE',
    delete: 'SYSTEM_ROBOT.DELETE',
    repository: 'SYSTEM_ROBOT.REPOSITORY',
    artifact: 'SYSTEM_ROBOT.ARTIFACT',
    tag: 'REPLICATION.TAG',
    'artifact-label': 'SYSTEM_ROBOT.ARTIFACT_LABEL',
    scan: 'SYSTEM_ROBOT.SCAN',
    'scanner-pull': 'SYSTEM_ROBOT.SCANNER_PULL',
    stop: 'SYSTEM_ROBOT.STOP',
    list: 'SYSTEM_ROBOT.LIST',
};

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
