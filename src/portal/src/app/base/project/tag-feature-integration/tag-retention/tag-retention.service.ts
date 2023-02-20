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
import { CURRENT_BASE_HREF } from '../../../../shared/units/utils';

@Injectable()
export class TagRetentionService {
    private I18nMap: object = {
        retain: 'ACTION_RETAIN',
        lastXDays: 'RULE_NAME_1',
        latestActiveK: 'RULE_NAME_2',
        latestPushedK: 'RULE_NAME_3',
        latestPulledN: 'RULE_NAME_4',
        always: 'RULE_NAME_5',
        nDaysSinceLastPull: 'RULE_NAME_6',
        nDaysSinceLastPush: 'RULE_NAME_7',
        'the artifacts from the last # days': 'RULE_TEMPLATE_1',
        'the most recent active # artifacts': 'RULE_TEMPLATE_2',
        'the most recently pushed # artifacts': 'RULE_TEMPLATE_3',
        'the most recently pulled # artifacts': 'RULE_TEMPLATE_4',
        'pulled within the last # days': 'RULE_TEMPLATE_6',
        'pushed within the last # days': 'RULE_TEMPLATE_7',
        repoMatches: 'MAT',
        repoExcludes: 'EXC',
        matches: 'MAT',
        excludes: 'EXC',
        withLabels: 'WITH',
        withoutLabels: 'WITHOUT',
        COUNT: 'UNIT_COUNT',
        DAYS: 'UNIT_DAY',
        none: 'NONE',
        nothing: 'NONE',
        'Parameters nDaysSinceLastPull is too large': 'DAYS_LARGE',
        'Parameters nDaysSinceLastPush is too large': 'DAYS_LARGE',
        'Parameters latestPushedK is too large': 'COUNT_LARGE',
        'Parameters latestPulledN is too large': 'COUNT_LARGE',
    };

    getI18nKey(str: string): string {
        if (this.I18nMap[str.trim()]) {
            return 'TAG_RETENTION.' + this.I18nMap[str.trim()];
        }
        return str;
    }

    seeLog(retentionId, executionId, taskId) {
        window.open(
            `${CURRENT_BASE_HREF}/retentions/${retentionId}/executions/${executionId}/tasks/${taskId}`,
            '_blank'
        );
    }
}
