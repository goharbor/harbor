import {
    Component,
    OnInit,
    OnDestroy,
    Output,
    EventEmitter, ViewChild, Input,
} from "@angular/core";
import { ImmutableRetentionRule, RuleMetadate } from "../../tag-retention/retention";
import { InlineAlertComponent } from "../../../../../shared/components/inline-alert/inline-alert.component";
import { ImmutableTagService } from "../../immutable-tag/immutable-tag.service";

const EXISTING_RULE = "TAG_RETENTION.EXISTING_RULE";
const INVALID_RULE = "TAG_RETENTION.INVALID_RULE";
@Component({
    selector: 'app-add-acceleration-rule',
    templateUrl: './add-acceleration-rule.component.html',
    styleUrls: ['./add-acceleration-rule.component.scss']
})
export class AddAccelerationRuleComponent implements OnInit, OnDestroy {
    addRuleOpened: boolean = false;
    @Output() clickAdd = new EventEmitter<{
        rule: ImmutableRetentionRule;
        isAdd: boolean
    }>();
    @Input() rules: ImmutableRetentionRule[];
    @Input() projectId: number;
    metadata: RuleMetadate = new RuleMetadate();
    isAdd: boolean = true;
    editRuleOrigin: ImmutableRetentionRule;
    onGoing: boolean = false;
    @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;

    repoSelect: string = 'repoMatches';
    repositories: string = '**';
    tagsSelect: string = 'matches';
    tagsInput: string = "**";
    repoSelectEditOrigin: string;
    repositoriesEditOrigin: string;
    tagsSelectEditOrigin: string;
    tagsInputEditOrigin: string;
    constructor(private immutableTagService: ImmutableTagService) {
    }

    ngOnInit(): void {
    }

    ngOnDestroy(): void {
    }
    canNotAdd(): boolean {
        if (this.onGoing) {
            return true;
        }
        if (!this.repositories) {
            return true;
        }
        if (!this.tagsInput) {
            return true;
        }
        // tslint:disable-next-line:triple-equals
        if (!this.isAdd && this.repoSelect != this.repoSelectEditOrigin) {
            return false;
        }
        // tslint:disable-next-line:triple-equals
        if (!this.isAdd && this.repositories != this.repositoriesEditOrigin) {
            return false;
        }
        // tslint:disable-next-line:triple-equals
        if (!this.isAdd && this.tagsSelect != this.tagsSelectEditOrigin) {
            return false;
        }
        // tslint:disable-next-line:triple-equals
        if (!this.isAdd && this.tagsInput != this.tagsInputEditOrigin) {
            return false;
        }
        return !this.isAdd;
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
        if (this.isExistingRule()) {
            this.inlineAlert.showInlineError(EXISTING_RULE);
            return;
        }
        // remove whitespaces
        const rule: ImmutableRetentionRule = new ImmutableRetentionRule(this.projectId);
        rule.scope_selectors.repository[0].pattern = this.repositories;
        rule.scope_selectors.repository[0].decoration = this.repoSelect;
        rule.tag_selectors[0].pattern = this.tagsInput;
        rule.tag_selectors[0].decoration = this.tagsSelect;
        rule.scope_selectors.repository[0].pattern = rule.scope_selectors.repository[0].pattern.replace(/\s+/g, "");
        rule.tag_selectors[0].pattern = rule.tag_selectors[0].pattern.replace(/\s+/g, "");
        if (rule.scope_selectors.repository[0].decoration !== "repoMatches"
            && rule.scope_selectors.repository[0].pattern) {
            let str = rule.scope_selectors.repository[0].pattern;
            str = str.replace(/[{}]/g, "");
            const arr = str.split(',');
            for (let i = 0; i < arr.length; i++) {
                if (arr[i] && arr[i].trim() && arr[i] === "**") {
                    this.inlineAlert.showInlineError(INVALID_RULE);
                    return;
                }
            }
        }
        this.clickAdd.emit({
            rule: rule,
            isAdd: this.isAdd
        });
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
        if (this.repoSelect !== rule.scope_selectors.repository[0].decoration) {
            return false;
        }
        if (this.repositories !== rule.scope_selectors.repository[0].pattern) {
            return false;
        }

        if (this.tagsSelect !== rule.tag_selectors[0].decoration) {
            return false;
        }
        return this.tagsInput === rule.tag_selectors[0].pattern;
    }

    getI18nKey(str: string) {
        return this.immutableTagService.getI18nKey(str);
    }
}

