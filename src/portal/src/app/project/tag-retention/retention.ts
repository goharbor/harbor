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
export class Retention {
    algorithm: string;
    rules: Array<Rule>;
    trigger: {
        kind: string;
    };
    scope: {
        level: string,
        ref: number;
    };
    cap: number;
    constructor () {
      this.rules = [];
      this.algorithm = "OR";
      this.trigger = {
          kind: ""
      };
    }
}
export class Rule {
    isDisabled: boolean;
    id: number;
    priority: number;
    action: string;
    template: string;
    params: {
        num: number
    };
    tag_selectors: Array<{kind: string; decoration: string; pattern: string}>;
    scope_selectors: {
        repository: Array<{
            kind: string;
            decoration: string;
            pattern: string
        }>;
    };
    constructor () {
        this.action = "retain";
        this.params = {
            num: null
        };
        this.scope_selectors = {
            repository: [
                {
                    kind: 'doublestar',
                    decoration: 'repoMatches',
                    pattern: '**'
                }
            ]
        };
        this.tag_selectors = [
            {
                kind: 'doublestar',
                decoration: 'matches',
                pattern: '**'
            },
            {
                kind: 'label',
                decoration: null,
                pattern: null
            }
        ];
    }
}


export class RuleMetadate {
    templates: Array<{
        rule_template: string;
        display_text: string;
        action: "retain",
        params: Array<{
            type: string;
            unit: string;
            required: boolean;
        }>
    }>;
    scope_selectors: Array<{
        display_text: string;
        kind: string;
        decorations: Array<string>
    }>;
    tag_selectors: Array<{
        display_text: string;
        kind: string;
        decorations: Array<string>
    }>;
    constructor () {
        this.templates = [];
        this.scope_selectors = [
            {
                display_text: null,
                kind: null,
                decorations: []
            }
            ];
        this.tag_selectors = [
            {
                display_text: null,
                kind: null,
                decorations: []
            },
            {
                display_text: null,
                kind: null,
                decorations: []
            }
        ];
    }
}

