// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
    EventEmitter,
    Input,
    OnDestroy,
    OnInit,
    Output,
    ViewChild,
} from '@angular/core';
import {
    catchError,
    debounceTime,
    distinctUntilChanged,
    finalize,
    map,
    switchMap,
} from 'rxjs/operators';
import {
    forkJoin,
    Observable,
    Subscription,
    throwError as observableThrowError,
    timer,
} from 'rxjs';
import { TranslateService } from '@ngx-translate/core';
import { ListReplicationRuleComponent } from './list-replication-rule/list-replication-rule.component';
import { CreateEditRuleComponent } from './create-edit-rule/create-edit-rule.component';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import {
    Comparator,
    ReplicationJob,
    ReplicationJobItem,
} from '../../../../shared/services';

import {
    calculatePage,
    CustomComparator,
    doFiltering,
    doSorting,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';

import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
    REFRESH_TIME_DIFFERENCE,
} from '../../../../shared/entities/shared.const';
import { ConfirmationDialogComponent } from '../../../../shared/components/confirmation-dialog';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import { Router } from '@angular/router';
import { FilterComponent } from '../../../../shared/components/filter/filter.component';
import { ClrDatagridStateInterface } from '@clr/angular';
import { errorHandler } from '../../../../shared/units/shared.utils';
import { ConfirmationMessage } from '../../../global-confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../../../global-confirmation-dialog/confirmation-state-message';
import { ReplicationService } from 'ng-swagger-gen/services/replication.service';
import { ReplicationPolicy } from '../../../../../../ng-swagger-gen/models/replication-policy';
import { ReplicationExecutionFilter } from '../replication';

const ONE_HOUR_SECONDS: number = 3600;
const ONE_MINUTE_SECONDS: number = 60;
const ONE_DAY_SECONDS: number = 24 * ONE_HOUR_SECONDS;
const IN_PROCESS: string = 'InProgress';

const ruleStatus: { [key: string]: any } = [
    { key: 'all', description: 'REPLICATION.ALL_STATUS' },
    { key: '1', description: 'REPLICATION.ENABLED' },
    { key: '0', description: 'REPLICATION.DISABLED' },
];

export class SearchOption {
    ruleId: number | string;
    ruleName: string = '';
    trigger: string = '';
    status: string = '';
    page: number = 1;
}

const STATUS_MAP = {
    Succeed: 'Succeeded',
};

@Component({
    selector: 'hbr-replication',
    templateUrl: './replication.component.html',
    styleUrls: ['./replication.component.scss'],
})
export class ReplicationComponent implements OnInit, OnDestroy {
    @Input() projectId: number | string;
    @Input() projectName: string;
    @Input() isSystemAdmin: boolean;
    @Input() withAdmiral: boolean;
    @Input() withReplicationJob: boolean;
    @Input() hasCreateReplicationPermission: boolean;
    @Input() hasUpdateReplicationPermission: boolean;
    @Input() hasDeleteReplicationPermission: boolean;
    @Input() hasExecuteReplicationPermission: boolean;
    @Output() openCreateRule = new EventEmitter<any>();
    @Output() openEdit = new EventEmitter<string | number>();
    @Output() goToRegistry = new EventEmitter<any>();

    search: SearchOption = new SearchOption();
    isOpenFilterTag: boolean;
    ruleStatus = ruleStatus;
    currentRuleStatus: { key: string; description: string };
    currentTerm: string;
    defaultFilter = 'trigger';
    selectedRow: ReplicationJobItem[] = [];
    isStopOnGoing: boolean;
    hiddenJobList = true;

    jobs: ReplicationJobItem[];

    @ViewChild(ListReplicationRuleComponent)
    listReplicationRule: ListReplicationRuleComponent;

    @ViewChild(CreateEditRuleComponent)
    createEditPolicyComponent: CreateEditRuleComponent;

    @ViewChild('replicationConfirmDialog')
    replicationConfirmDialog: ConfirmationDialogComponent;

    @ViewChild('StopConfirmDialog')
    StopConfirmDialog: ConfirmationDialogComponent;

    creationTimeComparator: Comparator<ReplicationJob> =
        new CustomComparator<ReplicationJob>('start_time', 'date');
    updateTimeComparator: Comparator<ReplicationJob> =
        new CustomComparator<ReplicationJob>('end_time', 'date');

    // Server driven pagination
    currentPage: number = 1;
    totalCount: number = 0;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.LIST_REPLICATION_RULE_COMPONENT_EXECUTIONS
    );
    currentState: ClrDatagridStateInterface;
    jobsLoading: boolean = false;
    timerDelay: Subscription;
    @ViewChild(FilterComponent, { static: true })
    filterComponent: FilterComponent;
    searchSub: Subscription;

    constructor(
        private router: Router,
        private errorHandlerEntity: ErrorHandler,
        private replicationService: ReplicationService,
        private operationService: OperationService,
        private translateService: TranslateService
    ) {}

    public get showPaginationIndex(): boolean {
        return this.totalCount > 0;
    }

    ngOnInit() {
        if (!this.searchSub) {
            this.searchSub = this.filterComponent.filterTerms
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged(),
                    switchMap(ruleName => {
                        this.listReplicationRule.loading = true;
                        this.listReplicationRule.page = 1;
                        return this.replicationService.listReplicationPoliciesResponse(
                            {
                                page: this.listReplicationRule.page,
                                pageSize: this.listReplicationRule.pageSize,
                                q: ruleName
                                    ? encodeURIComponent(`name=~${ruleName}`)
                                    : null,
                            }
                        );
                    })
                )
                .subscribe({
                    next: response => {
                        this.hideJobs();
                        // Get total count
                        if (response.headers) {
                            let xHeader: string =
                                response.headers.get('x-total-count');
                            if (xHeader) {
                                this.listReplicationRule.totalCount = parseInt(
                                    xHeader,
                                    0
                                );
                            }
                        }
                        this.listReplicationRule.selectedRow = null; // Clear selection
                        this.listReplicationRule.rules =
                            response.body as ReplicationPolicy[];
                        this.listReplicationRule.loading = false;
                    },
                    error: error => {
                        this.errorHandlerEntity.error(error);
                        this.listReplicationRule.loading = false;
                    },
                });
        }
        this.currentRuleStatus = this.ruleStatus[0];
    }

    ngOnDestroy() {
        if (this.timerDelay) {
            this.timerDelay.unsubscribe();
        }
        if (this.searchSub) {
            this.searchSub.unsubscribe();
            this.searchSub = null;
        }
    }

    // open replication rule
    openModal(): void {
        this.createEditPolicyComponent.openCreateEditRule();
    }

    // edit replication rule
    openEditRule(rule: ReplicationPolicy) {
        if (rule) {
            this.createEditPolicyComponent.openCreateEditRule(rule);
        }
    }

    goRegistry(): void {
        this.goToRegistry.emit();
    }

    goToLink(exeId: number): void {
        let linkUrl = ['harbor', 'replications', exeId, 'tasks'];
        this.router.navigate(linkUrl);
    }

    // Server driven data loading
    clrLoadJobs(withLoading: boolean, state: ClrDatagridStateInterface): void {
        if (!state || !state.page || !this.search.ruleId) {
            return;
        }
        this.pageSize = state.page.size;
        setPageSizeToLocalStorage(
            PageSizeMapKeys.LIST_REPLICATION_RULE_COMPONENT_EXECUTIONS,
            this.pageSize
        );
        this.currentState = state;

        let pageNumber: number = calculatePage(state);
        if (pageNumber <= 0) {
            pageNumber = 1;
        }
        const params: ReplicationService.ListReplicationExecutionsParams = {
            policyId: +this.search.ruleId,
            page: pageNumber,
            pageSize: this.pageSize,
            sort: getSortingString(state),
        };
        if (
            this.defaultFilter === ReplicationExecutionFilter.TRIGGER &&
            this.currentTerm
        ) {
            params.trigger = this.currentTerm;
        }
        if (
            this.defaultFilter === ReplicationExecutionFilter.STATUS &&
            this.currentTerm
        ) {
            params.status = this.currentTerm;
        }
        if (withLoading) {
            this.jobsLoading = true;
        }
        this.selectedRow = [];
        this.replicationService
            .listReplicationExecutionsResponse(params)
            .subscribe(
                response => {
                    this.totalCount = Number.parseInt(
                        response.headers.get('x-total-count'),
                        10
                    );
                    this.jobs = response.body as ReplicationJobItem[];
                    if (!this.timerDelay) {
                        this.timerDelay = timer(
                            REFRESH_TIME_DIFFERENCE,
                            REFRESH_TIME_DIFFERENCE
                        ).subscribe(() => {
                            let count: number = 0;
                            this.jobs.forEach(job => {
                                if (job.status === IN_PROCESS) {
                                    count++;
                                }
                            });
                            if (count > 0) {
                                this.clrLoadJobs(false, this.currentState);
                            } else {
                                this.timerDelay.unsubscribe();
                                this.timerDelay = null;
                            }
                        });
                    }
                    // Do filtering and sorting
                    this.jobs = doFiltering<ReplicationJobItem>(
                        this.jobs,
                        state
                    );
                    this.jobs = doSorting<ReplicationJobItem>(this.jobs, state);

                    this.jobsLoading = false;
                },
                error => {
                    this.jobsLoading = false;
                    this.errorHandlerEntity.error(error);
                }
            );
    }

    public doSearchExecutions(terms: string): void {
        this.currentTerm = terms.trim();
        // Trigger data loading and start from first page
        this.jobsLoading = true;
        this.currentPage = 1;
        this.jobsLoading = true;
        // Force reloading
        this.loadFirstPage();
    }

    loadFirstPage(): void {
        let st: ClrDatagridStateInterface = this.currentState;
        if (!st) {
            st = {
                page: {},
            };
        }
        st.page.size = this.pageSize;
        st.page.from = 0;
        st.page.to = this.pageSize - 1;

        this.clrLoadJobs(true, st);
    }

    selectOneRule(rule: ReplicationPolicy) {
        if (rule && rule.id) {
            this.hiddenJobList = false;
            this.search.ruleId = rule.id || '';
            this.loadFirstPage();
        }
    }

    replicateManualRule(rule: ReplicationPolicy) {
        if (rule) {
            let replicationMessage = new ConfirmationMessage(
                'REPLICATION.REPLICATION_TITLE',
                'REPLICATION.REPLICATION_SUMMARY',
                rule.name,
                rule,
                ConfirmationTargets.TARGET,
                ConfirmationButtons.REPLICATE_CANCEL
            );
            this.replicationConfirmDialog.open(replicationMessage);
        }
    }

    confirmReplication(message: ConfirmationAcknowledgement) {
        if (
            message &&
            message.source === ConfirmationTargets.TARGET &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            let rule: ReplicationPolicy = message.data;

            if (rule) {
                forkJoin(this.replicationOperate(rule)).subscribe(
                    item => {
                        this.selectOneRule(rule);
                    },
                    error => {
                        this.errorHandlerEntity.error(error);
                    }
                );
            }
        }
    }

    replicationOperate(rule: ReplicationPolicy): Observable<any> {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.REPLICATION';
        operMessage.data.id = rule.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = rule.name;
        this.operationService.publishInfo(operMessage);

        return this.replicationService
            .startReplication({
                execution: {
                    policy_id: rule.id,
                },
            })
            .pipe(
                map(response => {
                    this.translateService
                        .get('BATCH.REPLICATE_SUCCESS')
                        .subscribe(res =>
                            operateChanges(operMessage, OperationState.success)
                        );
                }),
                catchError(error => {
                    const message = errorHandler(error);
                    this.translateService
                        .get(message)
                        .subscribe(res =>
                            operateChanges(
                                operMessage,
                                OperationState.failure,
                                res
                            )
                        );
                    return observableThrowError(error);
                })
            );
    }

    doFilterJob($event: any): void {
        this.defaultFilter = $event['target'].value;
        this.doSearchJobs(this.currentTerm);
    }

    doSearchJobs(terms: string) {
        if (!terms) {
            return;
        }
        this.currentTerm = terms.trim();
        this.currentPage = 1;
        this.loadFirstPage();
    }

    hideJobs() {
        this.search.ruleId = 0;
        this.jobs = [];
        this.hiddenJobList = true;
    }

    openStopExecutionsDialog(targets: ReplicationJobItem[]) {
        let ExecutionId = targets.map(robot => robot.id).join(',');
        let StopExecutionsMessage = new ConfirmationMessage(
            'REPLICATION.STOP_TITLE',
            'REPLICATION.STOP_SUMMARY',
            ExecutionId,
            targets,
            ConfirmationTargets.STOP_EXECUTIONS,
            ConfirmationButtons.STOP_CANCEL
        );
        this.StopConfirmDialog.open(StopExecutionsMessage);
    }

    canStop() {
        if (this.selectedRow?.length) {
            let flag: boolean = true;
            this.selectedRow.forEach(item => {
                if (item.status !== IN_PROCESS) {
                    flag = false;
                }
            });
            return flag;
        }
        return false;
    }

    confirmStop(message: ConfirmationAcknowledgement) {
        if (
            message &&
            message.state === ConfirmationState.CONFIRMED &&
            message.source === ConfirmationTargets.STOP_EXECUTIONS
        ) {
            this.StopExecutions(message.data);
        }
    }

    StopExecutions(targets: ReplicationJobItem[]): void {
        if (targets && targets.length < 1) {
            return;
        }

        this.isStopOnGoing = true;
        if (this.jobs && this.jobs.length) {
            let ExecutionsStop$ = targets.map(target =>
                this.StopOperate(target)
            );
            forkJoin(ExecutionsStop$)
                .pipe(
                    finalize(() => {
                        this.refreshJobs();
                        this.isStopOnGoing = false;
                    })
                )
                .subscribe(
                    () => {},
                    error => {
                        this.errorHandlerEntity.error(error);
                    }
                );
        }
    }

    StopOperate(targets: ReplicationJobItem): any {
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.STOP_EXECUTIONS';
        operMessage.data.id = targets.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = targets.id;
        this.operationService.publishInfo(operMessage);

        return this.replicationService
            .stopReplication({
                id: targets.id,
            })
            .pipe(
                map(response => {
                    this.translateService
                        .get('BATCH.STOP_SUCCESS')
                        .subscribe(res =>
                            operateChanges(operMessage, OperationState.success)
                        );
                }),
                catchError(error => {
                    const message = errorHandler(error);
                    this.translateService
                        .get(message)
                        .subscribe(res =>
                            operateChanges(
                                operMessage,
                                OperationState.failure,
                                res
                            )
                        );
                    return observableThrowError(error);
                })
            );
    }

    reloadRules(isReady: boolean) {
        if (isReady) {
            this.search.ruleName = '';
            this.filterComponent.currentValue = '';
            this.listReplicationRule.refreshRule();
        }
    }

    refreshRules() {
        this.search.ruleName = '';
        if (this.filterComponent.currentValue) {
            this.filterComponent.currentValue = '';
            this.filterComponent.filterTerms.next(''); // will trigger refreshing
        } else {
            this.listReplicationRule.refreshRule(); // manually refresh
        }
    }

    refreshJobs() {
        this.currentTerm = '';
        this.currentPage = 1;
        let st: ClrDatagridStateInterface = {
            page: {
                from: 0,
                to: this.pageSize - 1,
                size: this.pageSize,
            },
        };
        this.clrLoadJobs(true, st);
    }

    openFilter(isOpen: boolean): void {
        this.isOpenFilterTag = isOpen;
    }

    getDuration(j: ReplicationJobItem) {
        if (!j) {
            return;
        }

        let start_time = new Date(j.start_time).getTime();
        let end_time = new Date(j.end_time).getTime();
        let timesDiff = end_time - start_time;
        let timesDiffSeconds = timesDiff / 1000;
        let minutes = Math.floor(timesDiffSeconds / ONE_MINUTE_SECONDS);
        let seconds = Math.floor(timesDiffSeconds % ONE_MINUTE_SECONDS);
        if (minutes > 0) {
            if (seconds === 0) {
                return minutes + 'm';
            }
            return minutes + 'm' + seconds + 's';
        }

        if (seconds > 0) {
            return seconds + 's';
        }

        if (seconds <= 0 && timesDiff > 0) {
            return timesDiff + 'ms';
        } else {
            return '-';
        }
    }

    getStatusStr(status: string): string {
        if (STATUS_MAP && STATUS_MAP[status]) {
            return STATUS_MAP[status];
        }
        return status;
    }
}
