import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { ImmutableTagService } from './immutable-tag.service';
import {
    ImmutableRetentionRule,
    RuleMetadate,
} from '../tag-retention/retention';
import { finalize } from 'rxjs/operators';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { clone } from '../../../../shared/units/utils';
import { AddImmutableRuleComponent } from './add-rule/add-immutable-rule.component';
import { ImmutableService } from '../../../../../../ng-swagger-gen/services/immutable.service';
import { RetentionService } from '../../../../../../ng-swagger-gen/services/retention.service';
import { ProjectService } from '../../../../../../ng-swagger-gen/services/project.service';

@Component({
    selector: 'app-immutable-tag',
    templateUrl: './immutable-tag.component.html',
    styleUrls: ['./immutable-tag.component.scss'],
})
export class ImmutableTagComponent implements OnInit {
    projectId: number;
    selectedItem: any = null;
    ruleIndex: number = -1;
    index: number = -1;
    rules: ImmutableRetentionRule[] = [];
    editIndex: number;
    loadingRule: boolean = false;

    @ViewChild('addRule') addRuleComponent: AddImmutableRuleComponent;
    constructor(
        private route: ActivatedRoute,
        private immutableTagService: ImmutableTagService,
        private immutableService: ImmutableService,
        private retentionService: RetentionService,
        public errorHandler: ErrorHandler,
        private projectService: ProjectService
    ) {}

    ngOnInit() {
        this.projectId = +this.route.snapshot.parent.parent.parent.params['id'];
        this.getRules();
        this.getMetadata();
    }

    getMetadata() {
        this.retentionService.getRentenitionMetadata().subscribe({
            next: res => {
                this.addRuleComponent.metadata = res as RuleMetadate;
            },
            error: err => {
                this.errorHandler.error(err);
            },
        });
    }

    getRules() {
        this.immutableService
            .ListImmuRules({
                projectNameOrId: this.projectId.toString(),
                pageSize: 15,
            })
            .subscribe({
                next: res => {
                    this.rules = res as ImmutableRetentionRule[];
                    this.loadingRule = false;
                },
                error: err => {
                    this.errorHandler.error(err);
                    this.loadingRule = false;
                },
            });
    }

    editRuleByIndex(index) {
        this.editIndex = index;
        this.addRuleComponent.rule = clone(this.rules[index]);
        this.addRuleComponent.editRuleOrigin = clone(this.rules[index]);
        this.addRuleComponent.open();
        this.addRuleComponent.isAdd = false;
        this.ruleIndex = -1;
    }
    toggleDisable(rule, isActionDisable) {
        let cloneRule = clone(rule);
        cloneRule.disabled = isActionDisable;
        this.ruleIndex = -1;
        this.loadingRule = true;
        this.immutableService
            .UpdateImmuRule({
                immutableRuleId: cloneRule.id,
                projectNameOrId: this.projectId.toString(),
                ImmutableRule: cloneRule,
            })
            .subscribe({
                next: res => {
                    this.getRules();
                },
                error: err => {
                    this.loadingRule = false;
                    this.errorHandler.error(err);
                },
            });
    }
    deleteRule(ruleId) {
        // // if rules is empty, clear schedule.
        this.ruleIndex = -1;
        this.loadingRule = true;
        this.immutableService
            .DeleteImmuRule({
                projectNameOrId: this.projectId.toString(),
                immutableRuleId: ruleId,
            })
            .subscribe({
                next: res => {
                    this.getRules();
                },
                error: err => {
                    this.loadingRule = false;
                    this.errorHandler.error(err);
                },
            });
    }

    openAddRule() {
        this.addRuleComponent.open();
        this.addRuleComponent.isAdd = true;
        this.addRuleComponent.rule = new ImmutableRetentionRule(this.projectId);
    }

    openEditor(index) {
        if (this.ruleIndex !== index) {
            this.ruleIndex = index;
        } else {
            this.ruleIndex = -1;
        }
    }

    refreshAfterCreatRetention() {
        this.projectService
            .getProject({
                projectNameOrId: this.projectId.toString(),
            })
            .subscribe({
                next: res => {
                    this.getRules();
                },
                error: err => {
                    this.loadingRule = false;
                    this.errorHandler.error(err);
                },
            });
    }

    clickAdd(rule) {
        this.loadingRule = true;
        this.addRuleComponent.onGoing = true;
        if (this.addRuleComponent.isAdd) {
            if (!rule.id) {
                this.immutableService
                    .CreateImmuRule({
                        projectNameOrId: this.projectId.toString(),
                        ImmutableRule: rule,
                    })
                    .pipe(
                        finalize(() => (this.addRuleComponent.onGoing = false))
                    )
                    .subscribe({
                        next: res => {
                            this.refreshAfterCreatRetention();
                            this.addRuleComponent.close();
                        },
                        error: err => {
                            if (err && err.error && err.error.message) {
                                err = this.immutableTagService.getI18nKey(
                                    err.error.message
                                );
                            }
                            this.addRuleComponent.inlineAlert.showInlineError(
                                err
                            );
                            this.loadingRule = false;
                        },
                    });
            } else {
                this.updateRule(rule);
            }
        } else {
            this.updateRule(rule);
        }
    }

    updateRule(rule: any) {
        this.immutableService
            .UpdateImmuRule({
                projectNameOrId: this.projectId.toString(),
                immutableRuleId: rule.id,
                ImmutableRule: rule,
            })
            .pipe(finalize(() => (this.addRuleComponent.onGoing = false)))
            .subscribe({
                next: res => {
                    this.getRules();
                    this.addRuleComponent.close();
                },
                error: err => {
                    this.loadingRule = false;
                    if (err && err.error && err.error.message) {
                        err = this.immutableTagService.getI18nKey(
                            err.error.message
                        );
                    }
                    this.addRuleComponent.inlineAlert.showInlineError(err);
                },
            });
    }

    formatPattern(pattern: string): string {
        let str: string = pattern;
        if (/^{\S+}$/.test(str)) {
            return str.slice(1, str.length - 1);
        }
        return str;
    }

    getI18nKey(str: string) {
        return this.immutableTagService.getI18nKey(str);
    }
}
