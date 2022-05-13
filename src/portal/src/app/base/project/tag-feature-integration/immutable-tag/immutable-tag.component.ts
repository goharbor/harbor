import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { ImmutableTagService } from './immutable-tag.service';
import { ImmutableRetentionRule } from '../tag-retention/retention';
import { finalize } from 'rxjs/operators';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { clone } from '../../../../shared/units/utils';
import { AddImmutableRuleComponent } from './add-rule/add-immutable-rule.component';

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
        public errorHandler: ErrorHandler
    ) {}

    ngOnInit() {
        this.projectId = +this.route.snapshot.parent.parent.parent.params['id'];
        this.getRules();
        this.getMetadata();
    }

    getMetadata() {
        this.immutableTagService.getRetentionMetadata().subscribe(
            response => {
                this.addRuleComponent.metadata = response;
            },
            error => {
                this.errorHandler.error(error);
            }
        );
    }

    getRules() {
        this.immutableTagService.getRules(this.projectId).subscribe(
            response => {
                this.rules = response as ImmutableRetentionRule[];
                this.loadingRule = false;
            },
            error => {
                this.errorHandler.error(error);
                this.loadingRule = false;
            }
        );
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
        this.immutableTagService
            .updateRule(this.projectId, cloneRule)
            .subscribe(
                response => {
                    this.getRules();
                },
                error => {
                    this.loadingRule = false;
                    this.errorHandler.error(error);
                }
            );
    }
    deleteRule(ruleId) {
        // // if rules is empty, clear schedule.
        this.ruleIndex = -1;
        this.loadingRule = true;
        this.immutableTagService.deleteRule(this.projectId, ruleId).subscribe(
            response => {
                this.getRules();
            },
            error => {
                this.loadingRule = false;
                this.errorHandler.error(error);
            }
        );
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
        this.immutableTagService.getProjectInfo(this.projectId).subscribe(
            response => {
                this.getRules();
            },
            error => {
                this.loadingRule = false;
                this.errorHandler.error(error);
            }
        );
    }

    clickAdd(rule) {
        this.loadingRule = true;
        this.addRuleComponent.onGoing = true;
        if (this.addRuleComponent.isAdd) {
            if (!rule.id) {
                this.immutableTagService
                    .createRule(this.projectId, rule)
                    .pipe(
                        finalize(() => (this.addRuleComponent.onGoing = false))
                    )
                    .subscribe(
                        response => {
                            this.refreshAfterCreatRetention();
                            this.addRuleComponent.close();
                        },
                        error => {
                            if (error && error.error && error.error.message) {
                                error = this.immutableTagService.getI18nKey(
                                    error.error.message
                                );
                            }
                            this.addRuleComponent.inlineAlert.showInlineError(
                                error
                            );
                            this.loadingRule = false;
                        }
                    );
            } else {
                this.immutableTagService
                    .updateRule(this.projectId, rule)
                    .pipe(
                        finalize(() => (this.addRuleComponent.onGoing = false))
                    )
                    .subscribe(
                        response => {
                            this.getRules();
                            this.addRuleComponent.close();
                        },
                        error => {
                            this.loadingRule = false;
                            if (error && error.error && error.error.message) {
                                error = this.immutableTagService.getI18nKey(
                                    error.error.message
                                );
                            }
                            this.addRuleComponent.inlineAlert.showInlineError(
                                error
                            );
                        }
                    );
            }
        } else {
            this.immutableTagService
                .updateRule(this.projectId, rule)
                .pipe(finalize(() => (this.addRuleComponent.onGoing = false)))
                .subscribe(
                    response => {
                        this.getRules();
                        this.addRuleComponent.close();
                    },
                    error => {
                        if (error && error.error && error.error.message) {
                            error = this.immutableTagService.getI18nKey(
                                error.error.message
                            );
                        }
                        this.addRuleComponent.inlineAlert.showInlineError(
                            error
                        );
                        this.loadingRule = false;
                    }
                );
        }
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
