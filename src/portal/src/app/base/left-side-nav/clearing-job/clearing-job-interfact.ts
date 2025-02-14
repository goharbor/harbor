export enum RetentionTimeUnit {
    HOURS = 'hours',
    DAYS = 'days',
}

export const RESOURCE_TYPES = [
    'create_artifact',
    'delete_artifact',
    'pull_artifact',
];

export const RESOURCE_TYPES_I18N_MAP = {
    artifact: 'AUDIT_LOG.ARTIFACT',
    user_login_logout: 'AUDIT_LOG.USER_LOGIN_LOGOUT',
    user: 'AUDIT_LOG.USER',
    project: 'AUDIT_LOG.PROJECT',
    configuration: 'AUDIT_LOG.CONFIGURATION',
    project_member: 'AUDIT_LOG.PROJECT_MEMBER',
};

export const JOB_STATUS = {
    PENDING: 'Pending',
    RUNNING: 'Running',
};

export const YES: string = 'TAG_RETENTION.YES';
export const NO: string = 'TAG_RETENTION.NO';

export const REFRESH_STATUS_TIME_DIFFERENCE: number = 5000;

export const WORKER_OPTIONS: number[] = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
