import { Component, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { finalize } from 'rxjs/operators';
import {
    EXECUTION_STATUS,
    TIME_OUT,
} from '../../p2p-provider/p2p-provider.service';
import { WebhookService } from '../../../../../../ng-swagger-gen/services/webhook.service';
import { Execution } from '../../../../../../ng-swagger-gen/models/execution';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import {
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';
import { ClrDatagridStateInterface } from '@clr/angular';
import { Task } from 'ng-swagger-gen/models/task';

@Component({
    selector: 'app-tasks',
    templateUrl: './tasks.component.html',
    styleUrls: ['./tasks.component.scss'],
})
export class TasksComponent implements OnInit, OnDestroy {
    projectId: number;
    policyId: number;
    tasks: Task[] = [];
    executionId: number;
    loading: boolean = true;
    inProgress: boolean = false;
    execution: Execution;
    executionTimeout: any;
    currentPage: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.WEBHOOK_TASKS_COMPONENT
    );
    totalCount: number;
    state: ClrDatagridStateInterface;
    timeoutForTaskList: any;
    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private webhookService: WebhookService,
        private messageHandlerService: MessageHandlerService
    ) {}
    ngOnInit(): void {
        this.projectId = +this.route.snapshot.parent.parent.params['id'];
        this.policyId = +this.route.snapshot.params['policyId'];
        this.executionId = +this.route.snapshot.params['executionId'];
        if (this.executionId) {
            this.getExecutionDetail(true);
        }
    }

    ngOnDestroy(): void {
        if (this.executionTimeout) {
            clearTimeout(this.executionTimeout);
            this.executionTimeout = null;
        }
        if (this.timeoutForTaskList) {
            clearTimeout(this.timeoutForTaskList);
            this.timeoutForTaskList = null;
        }
    }

    getExecutionDetail(withLoading: boolean): void {
        if (withLoading) {
            this.inProgress = true;
        }
        if (this.executionId) {
            this.webhookService
                .ListExecutionsOfWebhookPolicy({
                    webhookPolicyId: this.policyId,
                    projectNameOrId: this.projectId.toString(),
                    q: encodeURIComponent(`id=${this.executionId}`),
                })
                .pipe(finalize(() => (this.inProgress = false)))
                .subscribe({
                    next: res => {
                        if (res?.length) {
                            this.execution = res[0];
                            if (
                                !this.execution ||
                                this.execution.status ===
                                    EXECUTION_STATUS.PENDING ||
                                this.execution.status ===
                                    EXECUTION_STATUS.RUNNING ||
                                this.execution.status ===
                                    EXECUTION_STATUS.SCHEDULED
                            ) {
                                if (this.executionTimeout) {
                                    clearTimeout(this.executionTimeout);
                                    this.executionTimeout = null;
                                }
                                if (!this.executionTimeout) {
                                    this.executionTimeout = setTimeout(() => {
                                        this.getExecutionDetail(false);
                                    }, TIME_OUT);
                                }
                            }
                        }
                    },
                    error: error => {
                        this.messageHandlerService.error(error);
                    },
                });
        }
    }

    clrLoadTasks(state: ClrDatagridStateInterface, withLoading: boolean) {
        if (state) {
            this.state = state;
        }
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.P2P_TASKS_COMPONENT,
                this.pageSize
            );
        }
        if (withLoading) {
            this.loading = true;
        }
        let q: string;
        if (state && state.filters && state.filters.length) {
            q = encodeURIComponent(
                `${state.filters[0].property}=~${state.filters[0].value}`
            );
        }
        let sort: string;
        if (state && state.sort && state.sort.by) {
            sort = getSortingString(state);
        } else {
            // sort by creation_time desc by default
            sort = `-creation_time`;
        }
        if (withLoading) {
            this.loading = true;
        }
        this.webhookService
            .ListTasksOfWebhookExecutionResponse({
                projectNameOrId: this.projectId.toString(),
                webhookPolicyId: this.policyId,
                executionId: +this.executionId,
                page: this.currentPage,
                pageSize: this.pageSize,
                sort: sort,
                q: q,
            })
            .pipe(
                finalize(() => {
                    this.loading = false;
                })
            )
            .subscribe({
                next: res => {
                    if (res.headers) {
                        let xHeader: string = res.headers.get('x-total-count');
                        if (xHeader) {
                            this.totalCount = parseInt(xHeader, 0);
                        }
                    }
                    this.tasks = res.body;
                    this.setLoop();
                },
                error: error => {
                    this.messageHandlerService.error(error);
                },
            });
    }

    setLoop() {
        if (this.timeoutForTaskList) {
            clearTimeout(this.timeoutForTaskList);
            this.timeoutForTaskList = null;
        }
        if (this.tasks && this.tasks.length) {
            for (let i = 0; i < this.tasks.length; i++) {
                if (this.willChangStatus(this.tasks[i].status)) {
                    if (!this.timeoutForTaskList) {
                        this.timeoutForTaskList = setTimeout(() => {
                            this.clrLoadTasks(this.state, false);
                        }, TIME_OUT);
                    }
                }
            }
        }
    }

    willChangStatus(status: string): boolean {
        return (
            status === EXECUTION_STATUS.PENDING ||
            status === EXECUTION_STATUS.RUNNING ||
            status === EXECUTION_STATUS.SCHEDULED
        );
    }

    refreshTasks() {
        this.clrLoadTasks(this.state, true);
    }
    viewLog(taskId: number | string): string {
        return `/api/v2.0/projects/${this.projectId}/webhook/policies/${this.policyId}/executions/${this.executionId}/tasks/${taskId}/log`;
    }

    isInProgress(): boolean {
        return this.execution && this.willChangStatus(this.execution.status);
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

    toString(v: any) {
        if (v) {
            return JSON.stringify(v);
        }
        return '';
    }
}
