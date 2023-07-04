export enum RetentionTimeUnit {
    HOURS = 'hours',
    DAYS = 'days',
}

export const RETENTION_OPERATIONS = ['create', 'delete', 'pull'];

export const RETENTION_OPERATIONS_I18N_MAP = {
    pull: 'AUDIT_LOG.PULL',
    create: 'AUDIT_LOG.CREATE',
    delete: 'AUDIT_LOG.DELETE',
};

export const JOB_STATUS = {
    PENDING: 'Pending',
    RUNNING: 'Running',
};

export const YES: string = 'TAG_RETENTION.YES';
export const NO: string = 'TAG_RETENTION.NO';

export const REFRESH_STATUS_TIME_DIFFERENCE: number = 5000;

export const WORKER_OPTIONS: number[] = [1, 2, 3, 4, 5];
