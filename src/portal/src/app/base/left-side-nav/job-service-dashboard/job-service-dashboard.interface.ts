import { ScheduleTask } from '../../../../../ng-swagger-gen/models/schedule-task';

export enum All {
    ALL_WORKERS = 'all',
}

export enum PendingJobsActions {
    PAUSE = 'pause',
    RESUME = 'resume',
    STOP = 'stop',
}

export const INTERVAL: number = 10000;

export enum ScheduleStatusString {
    PAUSED = 'JOB_SERVICE_DASHBOARD.PAUSED',
    RUNNING = 'JOB_SERVICE_DASHBOARD.RUNNING_STATUS',
}

export enum ScheduleExecuteBtnString {
    RESUME_ALL = 'JOB_SERVICE_DASHBOARD.RESUME_ALL_BTN_TEXT',
    PAUSE_ALL = 'JOB_SERVICE_DASHBOARD.PAUSE_ALL_BTN_TEXT',
}

export enum JobType {
    SCHEDULER = 'scheduler',
    ALL = 'all',
}

export const CronTypeI18nMap = {
    None: 'SCHEDULE.NONE',
    Daily: 'SCHEDULE.DAILY',
    Weekly: 'SCHEDULE.WEEKLY',
    Hourly: 'SCHEDULE.HOURLY',
    Custom: 'SCHEDULE.CUSTOM',
};

export interface ScheduleListResponse {
    pageSize: number;
    currentPage: number;
    total: number;
    scheduleList: ScheduleTask[];
}
