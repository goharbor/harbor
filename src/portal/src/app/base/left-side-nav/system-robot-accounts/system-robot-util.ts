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
    HELM_CHART = 'helm-chart',
    HELM_CHART_VERSION = 'helm-chart-version',
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
    {
        resource: 'helm-chart',
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
];

export const ACTION_RESOURCE_I18N_MAP = {
    push: 'SYSTEM_ROBOT.PUSH_AND_PULL', // push permission contains pull permission
    pull: 'ROBOT_ACCOUNT.PULL',
    read: 'SYSTEM_ROBOT.READ',
    create: 'SYSTEM_ROBOT.CREATE',
    delete: 'SYSTEM_ROBOT.DELETE',
    repository: 'SYSTEM_ROBOT.REPOSITORY',
    artifact: 'SYSTEM_ROBOT.ARTIFACT',
    'helm-chart': 'SYSTEM_ROBOT.HELM',
    'helm-chart-version': 'SYSTEM_ROBOT.HELM_VERSION',
    tag: 'REPLICATION.TAG',
    'artifact-label': 'SYSTEM_ROBOT.ARTIFACT_LABEL',
    'helm-chart-version-label': 'SYSTEM_ROBOT.HELM_LABEL',
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
