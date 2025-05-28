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
import {Retention, Rule, RuleMetadate, Template} from '../retention';
import { TagRetentionService } from '../tag-retention.service';
import { compareValue } from '../../../../../shared/units/utils';
import { InlineAlertComponent } from '../../../../../shared/components/inline-alert/inline-alert.component';

const EXISTING_RULE = 'TAG_RETENTION.EXISTING_RULE';
const INVALID_RULE = 'TAG_RETENTION.INVALID_RULE';
const MAX = 2100000000;
@Component({
    selector: 'add-rule',
    templateUrl: './add-rule.component.html',
    styleUrls: ['./add-rule.component.scss'],
})
export class AddRuleComponent {
    addRuleOpened: boolean = false;
    @Output() clickAdd = new EventEmitter<Rule>();
    @Input() retention: Retention;
    metadata: RuleMetadate = new RuleMetadate();
    rule: Rule = new Rule();
    isAdd: boolean = true;
    editRuleOrigin: Rule;
    onGoing: boolean = false;
    @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;
    constructor(private tagRetentionService: TagRetentionService) {}

    set template(template) {
        this.rule.template = template;
    }

    get template() {
        return this.rule.template;
    }

    get unit(): string {
        let str = '';
        this.metadata.templates.forEach(t => {
          if (
            t.rule_template === this.rule.template &&
            t.action === 'retain'
          ) {
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

  filterTemplate(t: Template) {
    return t.action === 'retain';
  }

    get repoSelect() {
        return this.rule.scope_selectors.repository[0].decoration;
    }

    set repoSelect(repoSelect) {
        this.rule.scope_selectors.repository[0].decoration = repoSelect;
    }

    set repositories(repositories) {
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

    get repositories() {
        let str: string = this.rule.scope_selectors.repository[0].pattern;
        if (/^{\S+}$/.test(str)) {
            return str.slice(1, str.length - 1);
        }
        return str;
    }

    get tagsSelect() {
        return this.rule.tag_selectors[0].decoration;
    }

    set tagsSelect(tagsSelect) {
        this.rule.tag_selectors[0].decoration = tagsSelect;
    }

    set tagsInput(tagsInput) {
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

    get tagsInput() {
        let str: string = this.rule.tag_selectors[0].pattern;
        if (/^{\S+}$/.test(str)) {
            return str.slice(1, str.length - 1);
        }
        return str;
    }
    set untagged(untagged) {
        let extras = JSON.parse(this.rule.tag_selectors[0].extras);
        extras.untagged = untagged;
        this.rule.tag_selectors[0].extras = JSON.stringify(extras);
    }

    get untagged() {
        if (this.rule.tag_selectors[0] && this.rule.tag_selectors[0].extras) {
            let extras = JSON.parse(this.rule.tag_selectors[0].extras);
            if (extras.untagged !== undefined) {
                return extras.untagged;
            }
            return false;
        } else {
            return false;
        }
    }

    get labelsSelect() {
        return this.rule.tag_selectors[1].decoration;
    }

    set labelsSelect(labelsSelect) {
        this.rule.tag_selectors[1].decoration = labelsSelect;
    }

    set labelsInput(labelsInput) {
        this.rule.tag_selectors[1].pattern = labelsInput;
    }

    get labelsInput() {
        return this.rule.tag_selectors[1].pattern;
    }

    canNotAdd(): boolean {
        if (this.onGoing) {
            return true;
        }
        if (!this.isAdd && compareValue(this.editRuleOrigin, this.rule)) {
            return true;
        }
        if (!this.hasParam()) {
            return !(
                this.rule.template &&
                this.rule.scope_selectors.repository[0].pattern &&
                this.rule.scope_selectors.repository[0].pattern.replace(
                    /[{}]/g,
                    ''
                ) &&
                this.rule.tag_selectors[0].pattern &&
                this.rule.tag_selectors[0].pattern.replace(/[{}]/g, '')
            );
        } else {
            return !(
                this.rule.template &&
                this.rule.params[this.template] &&
                parseInt(this.rule.params[this.template], 10) >= 0 &&
                parseInt(this.rule.params[this.template], 10) < MAX &&
                this.rule.scope_selectors.repository[0].pattern &&
                this.rule.scope_selectors.repository[0].pattern.replace(
                    /[{}]/g,
                    ''
                ) &&
                this.rule.tag_selectors[0].pattern &&
                this.rule.tag_selectors[0].pattern.replace(/[{}]/g, '')
            );
        }
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
        // convert string "0" to number 0
        if (this.rule.params[this.template] === '0') {
            this.rule.params[this.template] = 0;
        }
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
        if (
            this.retention &&
            this.retention.rules &&
            this.retention.rules.length > 0
        ) {
            for (let i = 0; i < this.retention.rules.length; i++) {
                if (this.isSameRule(this.retention.rules[i])) {
                    return true;
                }
            }
        }
        return false;
    }
    isSameRule(rule: Rule): boolean {
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
        if (this.rule.template !== rule.template) {
            return false;
        }
        if (
            this.hasParam() &&
            JSON.stringify(this.rule.params) !== JSON.stringify(rule.params)
        ) {
            return false;
        }
        if (
            this.rule.tag_selectors[0].decoration !==
            rule.tag_selectors[0].decoration
        ) {
            return false;
        }
        if (
            this.rule.tag_selectors[0].extras !== rule.tag_selectors[0].extras
        ) {
            return false;
        }
        return (
            this.rule.tag_selectors[0].pattern === rule.tag_selectors[0].pattern
        );
    }

    getI18nKey(str: string) {
        return this.tagRetentionService.getI18nKey(str);
    }
    hasParam(): boolean {
        if (this.metadata && this.metadata.templates) {
            let flag: boolean = false;
            this.metadata.templates.forEach(t => {
              if (
                t.rule_template === this.template &&
                t.action === 'retain'
              ) {
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
