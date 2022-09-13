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

import { Injectable } from '@angular/core';

const ONE_MINUTE_SECONDS: number = 60;

@Injectable()
export class P2pProviderService {
    constructor() {}

    getDuration(start: string, end: string): string {
        if (!start || !end) {
            return '-';
        }
        let startTime = new Date(start).getTime();
        let endTime = new Date(end).getTime();
        let timesDiff = endTime - startTime;
        let timesDiffSeconds = timesDiff / 1000;
        let minutes = Math.floor(timesDiffSeconds / ONE_MINUTE_SECONDS);
        let seconds = Math.floor(timesDiffSeconds % ONE_MINUTE_SECONDS);
        if (minutes > 0) {
            if (seconds === 0) {
                return minutes + 'm';
            }
            return minutes + 'm' + seconds + 's';
        }
        if (seconds > 0) {
            return seconds + 's';
        }
        if (seconds <= 0 && timesDiff > 0) {
            return timesDiff + 'ms';
        } else {
            return '-';
        }
    }
    willChangStatus(status: string): boolean {
        return (
            status === EXECUTION_STATUS.PENDING ||
            status === EXECUTION_STATUS.RUNNING ||
            status === EXECUTION_STATUS.SCHEDULED
        );
    }
}

export enum EXECUTION_STATUS {
    PENDING = 'Pending',
    RUNNING = 'Running',
    STOPPED = 'Stopped',
    ERROR = 'Error',
    SUCCESS = 'Success',
    SCHEDULED = 'Scheduled',
}

export enum TRIGGER {
    MANUAL = 'manual',
    SCHEDULED = 'scheduled',
    EVENT_BASED = 'event_based',
}

export const TRIGGER_I18N_MAP = {
    manual: 'P2P_PROVIDER.MANUAL',
    scheduled: 'P2P_PROVIDER.SCHEDULED',
    event_based: 'P2P_PROVIDER.EVENT_BASED',
};

export const TIME_OUT: number = 7000;

export const PROJECT_SEVERITY_LEVEL_MAP = {
    critical: 5,
    high: 4,
    medium: 3,
    low: 2,
    none: 1,
};

export const PROJECT_SEVERITY_LEVEL_TO_TEXT_MAP = {
    5: 'critical',
    4: 'high',
    3: 'medium',
    2: 'low',
    1: 'none',
};

export enum FILTER_TYPE {
    REPOS = 'repository',
    TAG = 'tag',
    SIGNATURE = 'signature',
    LABEL = 'label',
    VULNERABILITY = 'vulnerability',
}
