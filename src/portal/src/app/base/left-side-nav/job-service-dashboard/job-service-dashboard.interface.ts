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
