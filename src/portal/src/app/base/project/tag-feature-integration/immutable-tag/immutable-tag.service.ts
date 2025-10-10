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

@Injectable()
export class ImmutableTagService {
    private I18nMap: object = {
        immutable: 'ACTION_IMMUTABLE',
        repoMatches: 'MAT',
        repoExcludes: 'EXC',
        matches: 'MAT',
        excludes: 'EXC',
        withLabels: 'WITH',
        withoutLabels: 'WITHOUT',
        none: 'NONE',
        nothing: 'NONE',
        always: 'RULE_NAME_5',
        nDaysSinceLastPull: 'RULE_NAME_6',
        nDaysSinceLastPush: 'RULE_NAME_7',
        immutable_template: 'RULE_NAME_5',
        'pulled within the last # days': 'RULE_TEMPLATE_6',
        'pushed within the last # days': 'RULE_TEMPLATE_7',
    };

    getI18nKey(str: string): string {
        if (this.I18nMap[str.trim()]) {
            return 'IMMUTABLE_TAG.' + this.I18nMap[str.trim()];
        }
        return str;
    }
}
