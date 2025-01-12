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
import {
    Component,
    Output,
    EventEmitter,
    ViewChild,
    Input,
} from '@angular/core';
import {
    ImmutableRetentionRule,
    RuleMetadate,
    Template,
} from '../../tag-retention/retention';
import { ImmutableTagService } from '../immutable-tag.service';
import { compareValue } from '../../../../../shared/units/utils';
import { InlineAlertComponent } from '../../../../../shared/components/inline-alert/inline-alert.component';

const EXISTING_RULE = 'TAG_RETENTION.EXISTING_RULE';
const INVALID_RULE = 'TAG_RETENTION.INVALID_RULE';
@Component({
    selector: 'app-add-immutable-rule',
    templateUrl: './add-immutable-rule.component.html',
    styleUrls: ['./add-immutable-rule.component.scss'],
})
export class AddImmutableRuleComponent {
    addRuleOpened: boolean = false;
    @Output() clickAdd = new EventEmitter<ImmutableRetentionRule>();
    @Input() rules: ImmutableRetentionRule[];
    @Input() projectId: number;
    metadata: RuleMetadate = new RuleMetadate();
    rule: ImmutableRetentionRule;
    isAdd: boolean = true;
    editRuleOrigin: ImmutableRetentionRule;
    onGoing: boolean = false;
    @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;
    constructor(private immutableTagService: ImmutableTagService) {
        this.rule = new ImmutableRetentionRule(this.projectId);
    }

    get repoSelect() {
        if (
            this.rule &&
            this.rule.scope_selectors &&
            this.rule.scope_selectors.repository[0]
        ) {
            return this.rule.scope_selectors.repository[0].decoration;
        }
        return '';
    }

    set repoSelect(repoSelect) {
        if (
            this.rule &&
            this.rule.scope_selectors &&
            this.rule.scope_selectors.repository[0]
        ) {
            this.rule.scope_selectors.repository[0].decoration = repoSelect;
        }
    }

    set repositories(repositories) {
        if (
            this.rule &&
            this.rule.scope_selectors &&
            this.rule.scope_selectors.repository &&
            this.rule.scope_selectors.repository[0]
        ) {
            if (
                repositories.indexOf(',') !== -1 &&
                repositories.indexOf('{') === -1 &&
                repositories.indexOf('}') === -1
            ) {
                this.rule.scope_selectors.repository[0].pattern =
                    '{' + repositories + '}';
            } else {
                this.rule.scope_selectors.repository[0].pattern = repositories;
            }
        }
    }

    get repositories() {
        if (
            this.rule &&
            this.rule.scope_selectors &&
            this.rule.scope_selectors.repository &&
            this.rule.scope_selectors.repository[0] &&
            this.rule.scope_selectors.repository[0].pattern
        ) {
            let str: string = this.rule.scope_selectors.repository[0].pattern;
            if (/^{\S+}$/.test(str)) {
                return str.slice(1, str.length - 1);
            }
            return str;
        }
        return '';
    }

    get tagsSelect() {
        if (
            this.rule &&
            this.rule.tag_selectors &&
            this.rule.tag_selectors[0]
        ) {
            return this.rule.tag_selectors[0].decoration;
        }
        return '';
    }

    set tagsSelect(tagsSelect) {
        if (
            this.rule &&
            this.rule.tag_selectors &&
            this.rule.tag_selectors[0]
        ) {
            this.rule.tag_selectors[0].decoration = tagsSelect;
        }
    }

    set tagsInput(tagsInput) {
        if (
            this.rule &&
            this.rule.tag_selectors &&
            this.rule.tag_selectors[0]
        ) {
            if (
                tagsInput.indexOf(',') !== -1 &&
                tagsInput.indexOf('{') === -1 &&
                tagsInput.indexOf('}') === -1
            ) {
                this.rule.tag_selectors[0].pattern = '{' + tagsInput + '}';
            } else {
                this.rule.tag_selectors[0].pattern = tagsInput;
            }
        }
    }

    get tagsInput() {
        if (
            this.rule &&
            this.rule.tag_selectors &&
            this.rule.tag_selectors[0] &&
            this.rule.tag_selectors[0].pattern
        ) {
            let str: string = this.rule.tag_selectors[0].pattern;
            if (/^{\S+}$/.test(str)) {
                return str.slice(1, str.length - 1);
            }
            return str;
        }
        return '';
    }

    filterTemplate(t: Template) {
        return t.action === 'immutable';
    }

    canNotAdd(): boolean {
        if (this.onGoing) {
            return true;
        }
        if (!this.isAdd && compareValue(this.editRuleOrigin, this.rule)) {
            return true;
        }
        return !(
            this.rule &&
            this.rule.scope_selectors &&
            this.rule.scope_selectors.repository &&
            this.rule.scope_selectors.repository[0] &&
            this.rule.scope_selectors.repository[0].pattern &&
            this.rule.scope_selectors.repository[0].pattern.replace(
                /[{}]/g,
                ''
            ) &&
            this.rule.tag_selectors &&
            this.rule.tag_selectors[0] &&
            this.rule.tag_selectors[0].pattern &&
            this.rule.tag_selectors[0].pattern.replace(/[{}]/g, '')
        );
    }

    open() {
        this.addRuleOpened = true;
        this.inlineAlert.alertClose = true;
        this.onGoing = false;
    }

    close() {
        this.addRuleOpened = false;
    }

    cancel() {
        this.close();
    }

    add() {
        // remove whitespaces
        this.rule.scope_selectors.repository[0].pattern =
            this.rule.scope_selectors.repository[0].pattern.replace(/\s+/g, '');
        this.rule.tag_selectors[0].pattern =
            this.rule.tag_selectors[0].pattern.replace(/\s+/g, '');
        if (
            this.rule.scope_selectors.repository[0].decoration !==
                'repoMatches' &&
            this.rule.scope_selectors.repository[0].pattern
        ) {
            let str = this.rule.scope_selectors.repository[0].pattern;
            str = str.replace(/[{}]/g, '');
            const arr = str.split(',');
            for (let i = 0; i < arr.length; i++) {
                if (arr[i] && arr[i].trim() && arr[i] === '**') {
                    this.inlineAlert.showInlineError(INVALID_RULE);
                    return;
                }
            }
        }
        if (this.isExistingRule()) {
            this.inlineAlert.showInlineError(EXISTING_RULE);
            return;
        }
        this.clickAdd.emit(this.rule);
    }
    isExistingRule(): boolean {
        if (this.rules && this.rules.length > 0) {
            for (let i = 0; i < this.rules.length; i++) {
                if (this.isSameRule(this.rules[i])) {
                    return true;
                }
            }
        }
        return false;
    }
    isSameRule(rule: ImmutableRetentionRule): boolean {
        if (
            this.rule.scope_selectors.repository[0].decoration !==
            rule.scope_selectors.repository[0].decoration
        ) {
            return false;
        }
        if (
            this.rule.scope_selectors.repository[0].pattern !==
            rule.scope_selectors.repository[0].pattern
        ) {
            return false;
        }

        if (
            this.rule.tag_selectors[0].decoration !==
            rule.tag_selectors[0].decoration
        ) {
            return false;
        }
        return (
            this.rule.tag_selectors[0].pattern === rule.tag_selectors[0].pattern
        );
    }

    getI18nKey(str: string) {
        return this.immutableTagService.getI18nKey(str);
    }

    set template(template) {
        this.rule.template = template;
    }

    get template() {
        return this.rule.template;
    }

    get unit(): string {
        let str = '';
        this.metadata.templates.forEach(t => {
            if (t.rule_template === this.rule.template) {
                str = t.params[0].unit;
            }
        });
        return str;
    }

    get num() {
        return this.rule.params[this.template];
    }

    set num(num) {
        if (num) {
            num = num.trim();
        }
        if (parseInt(num, 10) > 0) {
            num = parseInt(num, 10);
        }
        this.rule.params[this.template] = num;
    }

    hasParam(): boolean {
        if (this.metadata && this.metadata.templates) {
            let flag: boolean = false;
            this.metadata.templates.forEach(t => {
                if (t.rule_template === this.template) {
                    if (t.params && t.params.length > 0) {
                        flag = true;
                    }
                }
            });
            return flag;
        }
        return false;
    }
}
