import { Component, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';
import { debounceTime, finalize, switchMap } from 'rxjs/operators';
import {
    clone,
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';
import { Task } from '../../../../../../ng-swagger-gen/models/task';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { Project } from '../../project';
import {
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../../shared/services';
import { Execution } from '../../../../../../ng-swagger-gen/models/execution';
import { PreheatService } from '../../../../../../ng-swagger-gen/services/preheat.service';
import {
    EXECUTION_STATUS,
    P2pProviderService,
    TIME_OUT,
} from '../p2p-provider.service';
import { forkJoin, Observable, Subject, Subscription } from 'rxjs';
import { ClrDatagridStateInterface, ClrLoadingState } from '@clr/angular';

@Component({
    selector: 'task-list',
    templateUrl: './task-list.component.html',
    styleUrls: ['./task-list.component.scss'],
})
export class TaskListComponent implements OnInit, OnDestroy {
    projectId: number;
    projectName: string;
    isOpenFilterTag: boolean;
    inProgress: boolean = false;
    currentPage: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.P2P_TASKS_COMPONENT
    );
    totalCount: number;
    loading = true;
    tasks: Task[];
    stopOnGoing: boolean;
    executionId: string;
    preheatPolicyName: string;
    execution: Execution;
    hasUpdatePermission: boolean = false;
    btnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    timeout: any;
    timeoutForTaskList: any;
    searchString: string;
    private _searchSubject: Subject<string> = new Subject<string>();
    private _searchSubscription: Subscription;
    filterKey: string = 'id';
    constructor(
        private translate: TranslateService,
        private router: Router,
        private route: ActivatedRoute,
        private messageHandlerService: MessageHandlerService,
        private preheatService: PreheatService,
        private p2pProviderService: P2pProviderService,
        private userPermissionService: UserPermissionService
    ) {}

    ngOnInit(): void {
        this.projectId = +this.route.snapshot.parent.parent.params['id'];
        const resolverData = this.route.snapshot.parent.parent.data;
        if (resolverData) {
            let project = <Project>resolverData['projectResolver'];
            this.projectName = project.name;
        }
        this.executionId = this.route.snapshot.params['executionId'];
        this.preheatPolicyName =
            this.route.snapshot.params['preheatPolicyName'];
        if (this.executionId && this.preheatPolicyName && this.projectName) {
            this.getExecutionDetail(true);
        }
        this.getPermissions();
        this.subscribeSearch();
    }
    subscribeSearch() {
        if (!this._searchSubscription) {
            this._searchSubscription = this._searchSubject
                .pipe(
                    debounceTime(500),
                    switchMap(searchString => {
                        this.loading = true;
                        let params: string;
                        if (this.searchString) {
                            params = encodeURIComponent(
                                `${this.filterKey}=~${searchString}`
                            );
                        }
                        return this.preheatService
                            .ListTasksResponse({
                                projectName: this.projectName,
                                preheatPolicyName: this.preheatPolicyName,
                                executionId: +this.executionId,
                                page: this.currentPage,
                                pageSize: this.pageSize,
                                q: params,
                            })
                            .pipe(finalize(() => (this.loading = false)));
                    })
                )
                .subscribe(res => {
                    if (res.headers) {
                        let xHeader: string = res.headers.get('x-total-count');
                        if (xHeader) {
                            this.totalCount = parseInt(xHeader, 0);
                        }
                    }
                    this.tasks = res.body;
                    this.setLoop();
                });
        }
    }
    ngOnDestroy(): void {
        if (this.timeout) {
            clearTimeout(this.timeout);
            this.timeout = null;
        }
        if (this.timeoutForTaskList) {
            clearTimeout(this.timeoutForTaskList);
            this.timeoutForTaskList = null;
        }
        if (this._searchSubscription) {
            this._searchSubscription.unsubscribe();
            this._searchSubscription = null;
        }
    }
    getPermissions() {
        const permissionsList: Observable<boolean>[] = [];
        permissionsList.push(
            this.userPermissionService.getPermission(
                this.projectId,
                USERSTATICPERMISSION.P2P_PROVIDER.KEY,
                USERSTATICPERMISSION.P2P_PROVIDER.VALUE.UPDATE
            )
        );
        this.btnState = ClrLoadingState.LOADING;
        forkJoin(...permissionsList).subscribe(
            Rules => {
                [this.hasUpdatePermission] = Rules;
                this.btnState = ClrLoadingState.SUCCESS;
            },
            error => {
                this.messageHandlerService.error(error);
                this.btnState = ClrLoadingState.ERROR;
            }
        );
    }
    getExecutionDetail(withLoading: boolean): void {
        if (withLoading) {
            this.inProgress = true;
        }
        if (this.executionId) {
            this.preheatService
                .GetExecution({
                    projectName: this.projectName,
                    preheatPolicyName: this.preheatPolicyName,
                    executionId: +this.executionId,
                })
                .pipe(finalize(() => (this.inProgress = false)))
                .subscribe(
                    res => {
                        this.execution = res;
                        if (
                            !this.execution ||
                            this.p2pProviderService.willChangStatus(
                                this.execution.status
                            )
                        ) {
                            if (this.timeout) {
                                clearTimeout(this.timeout);
                                this.timeout = null;
                            }
                            if (!this.timeout) {
                                this.timeout = setTimeout(() => {
                                    this.getExecutionDetail(false);
                                }, TIME_OUT);
                            }
                        }
                    },
                    error => {
                        this.messageHandlerService.error(error);
                    }
                );
        }
    }
    trigger(): string {
        return this.execution && this.execution.trigger
            ? this.execution.trigger
            : '';
    }

    startTime(): string {
        return this.execution && this.execution.start_time
            ? this.execution.start_time
            : null;
    }

    successNum(): number {
        if (this.execution && this.execution.metrics) {
            return this.execution.metrics.success_task_count
                ? this.execution.metrics.success_task_count
                : 0;
        }
        return 0;
    }

    failedNum(): number {
        if (this.execution && this.execution.metrics) {
            return this.execution.metrics.error_task_count
                ? this.execution.metrics.error_task_count
                : 0;
        }
        return 0;
    }

    progressNum(): number {
        if (this.execution && this.execution.metrics) {
            const num: number =
                (this.execution.metrics.pending_task_count
                    ? this.execution.metrics.pending_task_count
                    : 0) +
                (this.execution.metrics.running_task_count
                    ? this.execution.metrics.running_task_count
                    : 0) +
                (this.execution.metrics.scheduled_task_count
                    ? this.execution.metrics.scheduled_task_count
                    : 0);
            return num ? num : 0;
        }
        return 0;
    }

    stoppedNum(): number {
        if (this.execution && this.execution.metrics) {
            return this.execution.metrics.stopped_task_count
                ? this.execution.metrics.stopped_task_count
                : 0;
        }
        return 0;
    }

    stopJob() {
        this.stopOnGoing = true;
        const execution: Execution = clone(this.execution);
        execution.status = EXECUTION_STATUS.STOPPED;
        this.preheatService
            .StopExecution({
                projectName: this.projectName,
                preheatPolicyName: this.preheatPolicyName,
                executionId: +this.executionId,
                execution: execution,
            })
            .subscribe(
                response => {
                    this.stopOnGoing = false;
                    this.getExecutionDetail(true);
                    this.translate
                        .get('REPLICATION.STOP_SUCCESS', {
                            param: this.executionId,
                        })
                        .subscribe((res: string) => {
                            this.messageHandlerService.showSuccess(res);
                        });
                },
                error => {
                    this.messageHandlerService.error(error);
                }
            );
    }

    viewLog(taskId: number | string): string {
        return (
            this.preheatService.rootUrl +
            `/projects/${this.projectName}/preheat/policies/${this.preheatPolicyName}/executions/${this.executionId}/tasks/${taskId}/logs`
        );
    }
    clrLoadTasks(withLoading, state?: ClrDatagridStateInterface): void {
        if (withLoading) {
            this.loading = true;
        }
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.P2P_TASKS_COMPONENT,
                this.pageSize
            );
        }
        let params: string;
        if (this.searchString) {
            params = encodeURIComponent(
                `${this.filterKey}=~${this.searchString}`
            );
        }
        this.preheatService
            .ListTasksResponse({
                projectName: this.projectName,
                preheatPolicyName: this.preheatPolicyName,
                executionId: +this.executionId,
                page: this.currentPage,
                pageSize: this.pageSize,
                q: params,
            })
            .pipe(
                finalize(() => {
                    this.loading = false;
                })
            )
            .subscribe(
                res => {
                    if (res.headers) {
                        let xHeader: string = res.headers.get('x-total-count');
                        if (xHeader) {
                            this.totalCount = parseInt(xHeader, 0);
                        }
                    }
                    this.tasks = res.body;
                    this.setLoop();
                },
                error => {
                    this.messageHandlerService.error(error);
                }
            );
    }
    onBack(): void {
        this.router.navigate([
            'harbor',
            'projects',
            `${this.projectId}`,
            'p2p-provider',
            'policies',
        ]);
    }
    // refresh icon
    refreshTasks(): void {
        this.currentPage = 1;
        this.totalCount = 0;
        this.clrLoadTasks(true);
    }
    getDuration(t: Task): string {
        return this.p2pProviderService.getDuration(t.start_time, t.end_time);
    }
    isInProgress(): boolean {
        return (
            this.execution &&
            this.p2pProviderService.willChangStatus(this.execution.status)
        );
    }
    isSuccess(): boolean {
        return (
            this.execution && this.execution.status === EXECUTION_STATUS.SUCCESS
        );
    }
    isFailed(): boolean {
        return (
            this.execution &&
            (this.execution.status === EXECUTION_STATUS.ERROR ||
                this.execution.status === EXECUTION_STATUS.STOPPED)
        );
    }
    canStop(): boolean {
        return (
            this.execution &&
            this.p2pProviderService.willChangStatus(this.execution.status)
        );
    }
    setLoop() {
        if (this.timeoutForTaskList) {
            clearTimeout(this.timeoutForTaskList);
            this.timeoutForTaskList = null;
        }
        if (this.tasks && this.tasks.length) {
            for (let i = 0; i < this.tasks.length; i++) {
                if (
                    this.p2pProviderService.willChangStatus(
                        this.tasks[i].status
                    )
                ) {
                    if (!this.timeoutForTaskList) {
                        this.timeoutForTaskList = setTimeout(() => {
                            this.clrLoadTasks(false);
                        }, TIME_OUT);
                    }
                }
            }
        }
    }
    selectFilterKey($event: any): void {
        this.filterKey = $event['target'].value;
    }
    doFilter(terms: string): void {
        this.searchString = terms;
        if (terms.trim()) {
            this._searchSubject.next(terms.trim());
        } else {
            this.clrLoadTasks(true);
        }
    }
    openFilter(isOpen: boolean): void {
        this.isOpenFilterTag = isOpen;
    }
}
