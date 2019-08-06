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
import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { AddRuleComponent } from "./add-rule/add-rule.component";
import { ClrDatagridStringFilterInterface } from "@clr/angular";
import { TagRetentionService } from "./tag-retention.service";
import { Retention, Rule } from "./retention";
import { Project } from "../project";
import { clone, ErrorHandler } from "@harbor/ui";
import { OriginCron } from "@harbor/ui";
import { CronScheduleComponent } from "@harbor/ui";

const MIN = 60000;
const SEC = 1000;
const MIN_STR = "min";
const SEC_STR = "sec";
const SCHEDULE_TYPE = {
    NONE: "None",
    DAILY: "Daily",
    WEEKLY: "Weekly",
    HOURLY: "Hourly",
    CUSTOM: "Custom"
};
@Component({
    selector: 'tag-retention',
    templateUrl: './tag-retention.component.html',
    styleUrls: ['./tag-retention.component.scss']
})
export class TagRetentionComponent implements OnInit {
    serialFilter: ClrDatagridStringFilterInterface<any> = {
        accepts(item: any, search: string): boolean {
            return item.id.toString().indexOf(search) !== -1;
        }
    };
    statusFilter: ClrDatagridStringFilterInterface<any> = {
        accepts(item: any, search: string): boolean {
            return item.status.toLowerCase().indexOf(search.toLowerCase()) !== -1;
        }
    };
    dryRunFilter: ClrDatagridStringFilterInterface<any> = {
        accepts(item: any, search: string): boolean {
            let str = item.dry_run ? 'YES' : 'NO';
            return str.indexOf(search) !== -1;
        }
    };
    projectId: number;
    isRetentionRunOpened: boolean = false;
    isAbortedOpened: boolean = false;
    selectedItem: any = null;
    ruleIndex: number = -1;
    index: number = -1;
    retentionId: number;
    retention: Retention = new Retention();
    editIndex: number;
    executionList = [];
    historyList = [];
    loadingExecutions: boolean = true;
    loadingHistories: boolean = true;
    label: string = 'TAG_RETENTION.TRIGGER';
    loadingRule: boolean = false;
    currentPage: number = 1;
    pageSize: number = 10;
    totalCount: number = 0;
    @ViewChild('cronScheduleComponent', { static: false })
    cronScheduleComponent: CronScheduleComponent;
    @ViewChild('addRule', { static: false }) addRuleComponent: AddRuleComponent;
    constructor(
        private route: ActivatedRoute,
        private tagRetentionService: TagRetentionService,
        private errorHandler: ErrorHandler,
    ) {
    }
    originCron(): OriginCron {
        let originCron: OriginCron = {
            type: SCHEDULE_TYPE.DAILY,
            cron: "0 0 0 * * *"
        };
        originCron.cron = this.retention.trigger.settings.cron;
        if (originCron.cron === "") {
            originCron.type = SCHEDULE_TYPE.NONE;
        } else if (originCron.cron === "0 0 * * * *") {
            originCron.type = SCHEDULE_TYPE.HOURLY;
        } else if (originCron.cron === "0 0 0 * * *") {
            originCron.type = SCHEDULE_TYPE.DAILY;
        } else if (originCron.cron === "0 0 0 * * 0") {
            originCron.type = SCHEDULE_TYPE.WEEKLY;
        } else {
            originCron.type = SCHEDULE_TYPE.CUSTOM;
        }
        return originCron;
    }

    ngOnInit() {
        this.projectId = +this.route.snapshot.parent.params['id'];
        this.retention.scope = {
            level: "project",
            ref: this.projectId
        };
        let resolverData = this.route.snapshot.parent.data;
        if (resolverData) {
            let project = <Project>resolverData["projectResolver"];
            if (project.metadata && project.metadata.retention_id) {
                this.retentionId = project.metadata.retention_id;
            }
        }
        this.getRetention();
        this.getMetadata();
    }
    updateCron(cron: string) {
        let retention: Retention = clone(this.retention);
        retention.trigger.settings.cron = cron;
        if (!this.retentionId) {
            this.tagRetentionService.createRetention(retention).subscribe(
                response => {
                    this.cronScheduleComponent.isEditMode = false;
                    this.refreshAfterCreatRetention();
                }, error => {
                    this.errorHandler.error(error);
                });
        } else {
            this.tagRetentionService.updateRetention(this.retentionId, retention).subscribe(
                response => {
                    this.cronScheduleComponent.isEditMode = false;
                    this.retention = retention;
                }, error => {
                    this.errorHandler.error(error);
                });
        }
    }
    getMetadata() {
        this.tagRetentionService.getRetentionMetadata().subscribe(
            response => {
                this.addRuleComponent.metadata = response;
            }, error => {
                this.errorHandler.error(error);
            });
    }

    getRetention() {
        if (this.retentionId) {
            this.tagRetentionService.getRetention(this.retentionId).subscribe(
                response => {
                    this.retention = response;
                    this.loadingRule = false;
                }, error => {
                    this.errorHandler.error(error);
                    this.loadingRule = false;
                });
        }
    }

    editRuleByIndex(index) {
        this.editIndex = index;
        this.addRuleComponent.rule = clone(this.retention.rules[index]);
        this.addRuleComponent.editRuleOrigin = clone(this.retention.rules[index]);
        this.addRuleComponent.open();
        this.addRuleComponent.isAdd = false;
        this.ruleIndex = -1;
    }
  toggleDisable(index, isActionDisable) {
        let retention: Retention = clone(this.retention);
        retention.rules[index].disabled = isActionDisable;
        this.ruleIndex = -1;
        this.loadingRule = true;
        this.tagRetentionService.updateRetention(this.retentionId, retention).subscribe(
            response => {
                this.loadingRule = false;
                this.retention = retention;
            }, error => {
                this.loadingRule = false;
                this.errorHandler.error(error);
            });
    }
    deleteRule(index) {
        let retention: Retention = clone(this.retention);
        retention.rules.splice(index, 1);
        this.ruleIndex = -1;
        this.loadingRule = true;
        this.tagRetentionService.updateRetention(this.retentionId, retention).subscribe(
            response => {
                this.loadingRule = false;
                this.retention = retention;
            }, error => {
                this.loadingRule = false;
                this.errorHandler.error(error);
            });
    }

    openAddRule() {
        this.addRuleComponent.open();
        this.addRuleComponent.isAdd = true;
        this.addRuleComponent.rule = new Rule();
    }

    runRetention() {
        this.isRetentionRunOpened = false;
        this.tagRetentionService.runNowTrigger(this.retentionId).subscribe(
            response => {
                this.refreshList();
            }, error => {
                this.errorHandler.error(error);
            });
    }

    whatIfRun() {
        this.tagRetentionService.whatIfRunTrigger(this.retentionId).subscribe(
            response => {
                this.refreshList();
            }, error => {
                this.errorHandler.error(error);
            });
    }

    refreshList() {
        this.index = -1 ;
        this.selectedItem = null;
        this.loadingExecutions = true;
        if (this.retentionId) {
            this.tagRetentionService.getRunNowList(this.retentionId, this.currentPage, this.pageSize).subscribe(
              response => {
                    // Get total count
                    if (response.headers) {
                        let xHeader: string = response.headers.get("x-total-count");
                        if (xHeader) {
                            this.totalCount = parseInt(xHeader, 0);
                        }
                    }
                    this.executionList = response.body as Array<any>;
                    this.loadingExecutions = false;
                    TagRetentionComponent.calculateDuration(this.executionList);
                }, error => {
                    this.loadingExecutions = false;
                    this.errorHandler.error(error);
                });
        } else {
            this.loadingExecutions = false;
        }
    }

    static calculateDuration(arr: Array<any>) {
        if (arr && arr.length > 0) {
            for (let i = 0; i < arr.length; i++) {
                let duration = new Date(arr[i].end_time).getTime() - new Date(arr[i].start_time).getTime();
                let min = Math.floor(duration / MIN);
                let sec = Math.floor((duration % MIN) / SEC);
                arr[i]['duration'] = "";
                if ((min || sec) && duration > 0) {
                    if (min) {
                        arr[i]['duration'] += '' + min + MIN_STR;
                    }
                    if (sec) {
                        arr[i]['duration'] += '' + sec + SEC_STR;
                    }
                } else if ( min === 0 && sec === 0 && duration > 0) {
                    arr[i]['duration'] = "0";
                } else {
                    arr[i]['duration'] = "N/A";
                }
            }
        }
    }

    abortRun() {
        this.isAbortedOpened = true;
        this.tagRetentionService.AbortRun(this.retentionId, this.selectedItem.id).subscribe(
            res => {
                this.refreshList();
            }, error => {
                this.errorHandler.error(error);
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

    openDetail(index, executionId) {
        if (this.index !== index) {
            this.index = index;
            this.historyList = [];
            this.loadingHistories = true;
            this.tagRetentionService.getExecutionHistory(this.retentionId, executionId).subscribe(
                res => {
                    this.loadingHistories = false;
                    this.historyList = res;
                    TagRetentionComponent.calculateDuration(this.historyList);
                }, error => {
                    this.loadingHistories = false;
                    this.errorHandler.error(error);
                });
        } else {
            this.index = -1;
        }
    }

    refreshAfterCreatRetention() {
        this.tagRetentionService.getProjectInfo(this.projectId).subscribe(
            response => {
                this.retentionId = response.metadata.retention_id;
                this.getRetention();
            }, error => {
                this.loadingRule = false;
                this.errorHandler.error(error);
            });
    }

    clickAdd(rule) {
        this.loadingRule = true;
        if (this.addRuleComponent.isAdd) {
            let retention: Retention = clone(this.retention);
            retention.rules.push(rule);
            if (!this.retentionId) {
                this.tagRetentionService.createRetention(retention).subscribe(
                    response => {
                        this.refreshAfterCreatRetention();
                    }, error => {
                        this.errorHandler.error(error);
                        this.loadingRule = false;
                    });
            } else {
                this.tagRetentionService.updateRetention(this.retentionId, retention).subscribe(
                    response => {
                        this.loadingRule = false;
                        this.retention = retention;
                    }, error => {
                        this.loadingRule = false;
                        this.errorHandler.error(error);
                    });
            }
        } else {
            let retention: Retention = clone(this.retention);
            retention.rules[this.editIndex] = rule;
            this.tagRetentionService.updateRetention(this.retentionId, retention).subscribe(
                response => {
                    this.retention = retention;
                    this.loadingRule = false;
                }, error => {
                    this.errorHandler.error(error);
                    this.loadingRule = false;
                });
        }
    }

    seeLog(executionId, taskId) {
        this.tagRetentionService.seeLog(this.retentionId, executionId, taskId);
    }

    formatPattern(pattern: string): string {
        return pattern.replace(/[{}]/g, "");
    }

    getI18nKey(str: string) {
        return this.tagRetentionService.getI18nKey(str);
    }
    clrLoad() {
        this.refreshList();
    }
}
