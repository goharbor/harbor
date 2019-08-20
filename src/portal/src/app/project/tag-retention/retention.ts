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
        references: object;
        settings: {
            cron: string;
        }
    };
    scope: {
        level: string,
        ref: number;
    };
    cap: number;

    constructor() {
        this.rules = [];
        this.algorithm = "or";
        this.trigger = {
            kind: "Schedule",
            references: {},
            settings: {
                cron: "",
            }
        };
    }
}

export class Rule {
    disabled: boolean;
    id: number;
    priority: number;
    action: string;
    template: string;
    params: object;
    tag_selectors: Array<Selector>;
    scope_selectors: {
        repository: Array<Selector>;
    };

    constructor() {
        this.disabled = false;
        this.action = "retain";
        this.params = {};
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
                decoration: "withLabels",
                pattern: null
            }
        ];
    }
}

export class Selector {
    kind: string;
    decoration: string;
    pattern: string;
}

export class Param {
    type: string;
    unit: string;
    required: boolean;
}

export class Template {
    rule_template: string;
    display_text: string;
    action: "retain";
    params: Array<Param>;
}

export class SelectorRuleMetadate {
    display_text: string;
    kind: string;
    decorations: Array<string>;
}

export class RuleMetadate {
    templates: Array<Template>;
    scope_selectors: Array<SelectorRuleMetadate>;
    tag_selectors: Array<SelectorRuleMetadate>;

    constructor() {
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

