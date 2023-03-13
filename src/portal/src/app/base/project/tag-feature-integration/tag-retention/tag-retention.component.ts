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
import { Component, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { AddRuleComponent } from './add-rule/add-rule.component';
import { ClrDatagridStateInterface } from '@clr/angular';
import { TagRetentionService } from './tag-retention.service';
import {
    PENDING,
    Retention,
    RetentionAction,
    Rule,
    RuleMetadate,
    RUNNING,
    TIMEOUT,
} from './retention';
import { finalize } from 'rxjs/operators';
import { CronScheduleComponent } from '../../../../shared/components/cron-schedule';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { OriginCron } from '../../../../shared/services';
import {
    clone,
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';
import { RetentionService } from '../../../../../../ng-swagger-gen/services/retention.service';
import { RetentionPolicy } from '../../../../../../ng-swagger-gen/models/retention-policy';
import { ProjectService } from '../../../../../../ng-swagger-gen/services/project.service';

const MIN = 60000;
const SEC = 1000;
const MIN_STR = 'min';
const SEC_STR = 'sec';
const SCHEDULE_TYPE = {
    NONE: 'None',
    DAILY: 'Daily',
    WEEKLY: 'Weekly',
    HOURLY: 'Hourly',
    CUSTOM: 'Custom',
};
const DECORATION = {
    MATCHES: 'matches',
    EXCLUDES: 'excludes',
};

@Component({
    selector: 'tag-retention',
    templateUrl: './tag-retention.component.html',
    styleUrls: ['./tag-retention.component.scss'],
})
export class TagRetentionComponent implements OnInit, OnDestroy {
    projectId: number;
    isRetentionRunOpened: boolean = false;
    isAbortedOpened: boolean = false;
    isConfirmOpened: boolean = false;
    cron: string;
    selectedItem: any = null;
    ruleIndex: number = -1;
    retentionId: number;
    retention: Retention = new Retention();
    editIndex: number;
    executionList = [];
    executionId: number;
    loadingExecutions: boolean = true;
    label: string = 'TAG_RETENTION.TRIGGER';
    loadingRule: boolean = false;
    currentPage: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.TAG_RETENTION_COMPONENT
    );
    totalCount: number = 0;
    @ViewChild('cronScheduleComponent')
    cronScheduleComponent: CronScheduleComponent;
    @ViewChild('addRule') addRuleComponent: AddRuleComponent;
    executionTimeout;
    constructor(
        private route: ActivatedRoute,
        private tagRetentionService: TagRetentionService,
        private retentionService: RetentionService,
        private errorHandler: ErrorHandler,
        private projectService: ProjectService
    ) {}
    originCron(): OriginCron {
        let originCron: OriginCron = {
            type: SCHEDULE_TYPE.NONE,
            cron: '',
        };
        originCron.cron = this.retention.trigger.settings.cron;
        if (originCron.cron === '') {
            originCron.type = SCHEDULE_TYPE.NONE;
        } else if (originCron.cron === '0 0 * * * *') {
            originCron.type = SCHEDULE_TYPE.HOURLY;
        } else if (originCron.cron === '0 0 0 * * *') {
            originCron.type = SCHEDULE_TYPE.DAILY;
        } else if (originCron.cron === '0 0 0 * * 0') {
            originCron.type = SCHEDULE_TYPE.WEEKLY;
        } else {
            originCron.type = SCHEDULE_TYPE.CUSTOM;
        }
        return originCron;
    }

    ngOnInit() {
        this.projectId = +this.route.snapshot.parent.parent.parent.params['id'];
        this.retention.scope = {
            level: 'project',
            ref: this.projectId,
        };
        this.refreshAfterCreatRetention();
        this.getMetadata();
    }
    ngOnDestroy() {
        if (this.executionTimeout) {
            clearTimeout(this.executionTimeout);
            this.executionTimeout = null;
        }
    }

    openConfirm(cron: string) {
        if (cron) {
            this.isConfirmOpened = true;
            this.cron = cron;
        } else {
            this.updateCron(cron);
        }
    }
    closeConfirm() {
        this.isConfirmOpened = false;
        this.updateCron(this.cron);
    }
    updateCron(cron: string) {
        let retention: RetentionPolicy = clone(this.retention);
        retention.trigger.settings['cron'] = cron;
        if (retention?.trigger?.settings['next_scheduled_time']) {
            // should not have next_scheduled_time for updating
            delete retention?.trigger?.settings['next_scheduled_time'];
        }
        if (!this.retentionId) {
            this.retentionService
                .createRetention({
                    policy: retention,
                })
                .subscribe({
                    next: res => {
                        this.cronScheduleComponent.isEditMode = false;
                        this.refreshAfterCreatRetention();
                    },
                    error: err => {
                        this.errorHandler.error(err);
                    },
                });
        } else {
            this.retentionService
                .updateRetention({
                    id: this.retentionId,
                    policy: retention,
                })
                .subscribe({
                    next: res => {
                        this.cronScheduleComponent.isEditMode = false;
                        this.getRetention();
                    },
                    error: err => {
                        this.errorHandler.error(err);
                    },
                });
        }
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

    getRetention() {
        if (this.retentionId) {
            this.retentionService
                .getRetention({
                    id: this.retentionId,
                })
                .subscribe({
                    next: res => {
                        if (res?.rules?.length) {
                            res.rules.forEach(item => {
                                if (!item.params) {
                                    item.params = {};
                                }
                                this.setRuleUntagged(item);
                            });
                        }
                        this.retention = res as Retention;
                        this.loadingRule = false;
                    },
                    error: err => {
                        this.errorHandler.error(err);
                        this.loadingRule = false;
                    },
                });
        }
    }

    editRuleByIndex(index) {
        this.editIndex = index;
        this.addRuleComponent.rule = clone(this.retention.rules[index]);
        this.addRuleComponent.editRuleOrigin = clone(
            this.retention.rules[index]
        );
        this.addRuleComponent.open();
        this.addRuleComponent.isAdd = false;
        this.ruleIndex = -1;
    }
    toggleDisable(index, isActionDisable) {
        let retention: RetentionPolicy = clone(this.retention);
        retention.rules[index].disabled = isActionDisable;
        this.ruleIndex = -1;
        this.loadingRule = true;
        this.retentionService
            .updateRetention({
                id: this.retentionId,
                policy: retention,
            })
            .subscribe({
                next: res => {
                    this.getRetention();
                },
                error: err => {
                    this.loadingRule = false;
                    this.errorHandler.error(err);
                },
            });
    }
    deleteRule(index) {
        let retention: RetentionPolicy = clone(this.retention);
        retention.rules.splice(index, 1);
        // if rules is empty, clear schedule.
        if (retention.rules && retention.rules.length === 0) {
            retention.trigger.settings['cron'] = '';
        }
        this.ruleIndex = -1;
        this.loadingRule = true;
        this.retentionService
            .updateRetention({
                id: this.retentionId,
                policy: retention,
            })
            .subscribe({
                next: res => {
                    this.getRetention();
                },
                error: err => {
                    this.loadingRule = false;
                    this.errorHandler.error(err);
                },
            });
    }
    setRuleUntagged(rule) {
        if (!rule.tag_selectors[0].extras) {
            if (rule.tag_selectors[0].decoration === DECORATION.MATCHES) {
                rule.tag_selectors[0].extras = JSON.stringify({
                    untagged: true,
                });
            }
            if (rule.tag_selectors[0].decoration === DECORATION.EXCLUDES) {
                rule.tag_selectors[0].extras = JSON.stringify({
                    untagged: false,
                });
            }
        } else {
            let extras = JSON.parse(rule.tag_selectors[0].extras);
            if (extras.untagged === undefined) {
                if (rule.tag_selectors[0].decoration === DECORATION.MATCHES) {
                    extras.untagged = true;
                }
                if (rule.tag_selectors[0].decoration === DECORATION.EXCLUDES) {
                    extras.untagged = false;
                }
                rule.tag_selectors[0].extras = JSON.stringify(extras);
            }
        }
    }
    openAddRule() {
        this.addRuleComponent.open();
        this.addRuleComponent.isAdd = true;
        this.addRuleComponent.rule = new Rule();
    }

    runRetention() {
        this.isRetentionRunOpened = false;
        this.retentionService
            .triggerRetentionExecution({
                id: this.retentionId,
                body: {
                    dry_run: false,
                },
            })
            .subscribe({
                next: res => {
                    this.refreshList();
                },
                error: err => {
                    this.errorHandler.error(err);
                },
            });
    }

    whatIfRun() {
        this.retentionService
            .triggerRetentionExecution({
                id: this.retentionId,
                body: {
                    dry_run: true,
                },
            })
            .subscribe({
                next: res => {
                    this.refreshList();
                },
                error: err => {
                    this.errorHandler.error(err);
                },
            });
    }
    loopGettingExecutions() {
        if (
            this.executionList &&
            this.executionList.length &&
            this.executionList.some(item => {
                return item.status === RUNNING || item.status === PENDING;
            })
        ) {
            this.executionTimeout = setTimeout(() => {
                this.retentionService
                    .listRetentionExecutionsResponse({
                        id: this.retentionId,
                        page: this.currentPage,
                        pageSize: this.pageSize,
                    })
                    .pipe(finalize(() => (this.loadingExecutions = false)))
                    .subscribe(res => {
                        // Get total count
                        if (res.headers) {
                            let xHeader: string =
                                res.headers.get('x-total-count');
                            if (xHeader) {
                                this.totalCount = parseInt(xHeader, 0);
                            }
                        }
                        // data grid will be re-rendered if reassign "this.executionList"
                        // so only refresh the status and the end_time property
                        if (res?.body?.length) {
                            res.body.forEach(item => {
                                this.executionList.forEach(item2 => {
                                    if (item2.id === item.id) {
                                        item2.status = item.status;
                                        item2.end_time = item.end_time;
                                    }
                                });
                            });
                        }
                        TagRetentionComponent.calculateDuration(
                            this.executionList
                        );
                        this.loopGettingExecutions();
                    });
            }, TIMEOUT);
        }
    }
    refreshList(state?: ClrDatagridStateInterface) {
        this.selectedItem = null;
        if (this.retentionId) {
            if (state && state.page) {
                this.pageSize = state.page.size;
                setPageSizeToLocalStorage(
                    PageSizeMapKeys.TAG_RETENTION_COMPONENT,
                    this.pageSize
                );
            }
            this.loadingExecutions = true;
            this.retentionService
                .listRetentionExecutionsResponse({
                    id: this.retentionId,
                    page: this.currentPage,
                    pageSize: this.pageSize,
                })
                .pipe(finalize(() => (this.loadingExecutions = false)))
                .subscribe({
                    next: res => {
                        // Get total count
                        if (res.headers) {
                            let xHeader: string =
                                res.headers.get('x-total-count');
                            if (xHeader) {
                                this.totalCount = parseInt(xHeader, 0);
                            }
                        }
                        this.executionList = res.body as Array<any>;
                        TagRetentionComponent.calculateDuration(
                            this.executionList
                        );
                        this.loopGettingExecutions();
                    },
                    error: err => {
                        this.errorHandler.error(err);
                    },
                });
        } else {
            setTimeout(() => {
                this.loadingExecutions = false;
            }, 0);
        }
    }

    static calculateDuration(arr: Array<any>) {
        if (arr && arr.length > 0) {
            for (let i = 0; i < arr.length; i++) {
                if (arr[i].end_time && arr[i].start_time) {
                    let duration =
                        new Date(arr[i].end_time).getTime() -
                        new Date(arr[i].start_time).getTime();
                    let min = Math.floor(duration / MIN);
                    let sec = Math.floor((duration % MIN) / SEC);
                    arr[i]['duration'] = '';
                    if ((min || sec) && duration > 0) {
                        if (min) {
                            arr[i]['duration'] += '' + min + MIN_STR;
                        }
                        if (sec) {
                            arr[i]['duration'] += '' + sec + SEC_STR;
                        }
                    } else {
                        arr[i]['duration'] = '0';
                    }
                } else {
                    arr[i]['duration'] = 'N/A';
                }
            }
        }
    }

    abortRun() {
        this.isAbortedOpened = true;
        this.retentionService
            .operateRetentionExecution({
                id: this.retentionId,
                eid: this.selectedItem.id,
                body: {
                    action: RetentionAction.STOP,
                },
            })
            .subscribe({
                next: res => {
                    this.refreshList();
                },
                error: err => {
                    this.errorHandler.error(err);
                },
            });
    }

    abortRetention() {
        this.isAbortedOpened = false;
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
                    this.retentionId = +res.metadata.retention_id;
                    this.refreshList();
                    this.getRetention();
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
            let retention: RetentionPolicy = clone(this.retention);
            retention.rules.push(rule);
            if (!this.retentionId) {
                this.retentionService
                    .createRetention({
                        policy: retention,
                    })
                    .subscribe({
                        next: res => {
                            this.refreshAfterCreatRetention();
                            this.addRuleComponent.close();
                            this.addRuleComponent.onGoing = false;
                        },
                        error: err => {
                            if (err && err.error && err.error.message) {
                                err = this.tagRetentionService.getI18nKey(
                                    err.error.message
                                );
                            }
                            this.addRuleComponent.inlineAlert.showInlineError(
                                err
                            );
                            this.loadingRule = false;
                            this.addRuleComponent.onGoing = false;
                        },
                    });
            } else {
                this.updateRetention(retention);
            }
        } else {
            let retention: RetentionPolicy = clone(this.retention);
            retention.rules[this.editIndex] = rule;
            this.updateRetention(retention);
        }
    }

    updateRetention(retention: RetentionPolicy) {
        this.retentionService
            .updateRetention({
                id: this.retentionId,
                policy: retention,
            })
            .subscribe({
                next: res => {
                    this.getRetention();
                    this.addRuleComponent.close();
                    this.addRuleComponent.onGoing = false;
                },
                error: err => {
                    this.loadingRule = false;
                    this.addRuleComponent.onGoing = false;
                    if (err && err.error && err.error.message) {
                        err = this.tagRetentionService.getI18nKey(
                            err.error.message
                        );
                    }
                    this.addRuleComponent.inlineAlert.showInlineError(err);
                },
            });
    }

    seeLog(executionId, taskId) {
        this.tagRetentionService.seeLog(this.retentionId, executionId, taskId);
    }

    formatPattern(pattern: string): string {
        let str: string = pattern;
        if (/^{\S+}$/.test(str)) {
            return str.slice(1, str.length - 1);
        }
        return str;
    }

    getI18nKey(str: string) {
        return this.tagRetentionService.getI18nKey(str);
    }
    clrLoad(state: ClrDatagridStateInterface) {
        this.refreshList(state);
    }
    /**
     *
     * @param extras Json string
     */
    showUntagged(extras) {
        return JSON.parse(extras).untagged;
    }
}
