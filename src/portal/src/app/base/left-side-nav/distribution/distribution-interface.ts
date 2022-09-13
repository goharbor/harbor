import { Instance } from '../../../../../ng-swagger-gen/models/instance';

export class AuthMode {
    static NONE = 'NONE';
    static BASIC = 'BASIC';
    static OAUTH = 'OAUTH';
    static CUSTOM = 'CUSTOM';
}

export enum PreheatingStatusEnum {
    // front status
    NOT_PREHEATED = 'NOT_PREHEATED',
    // back-end status
    PENDING = 'PENDING',
    RUNNING = 'RUNNING',
    SUCCESS = 'SUCCESS',
    FAIL = 'FAIL',
}

export interface FrontInstance extends Instance {
    hasCheckHealth?: boolean;
    pingStatus?: string;
}

export const HEALTHY: string = 'Healthy';
export const UNHEALTHY: string = 'Unhealthy';
