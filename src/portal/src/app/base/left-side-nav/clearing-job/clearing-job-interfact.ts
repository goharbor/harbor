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
