import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';
import { finalize } from 'rxjs/operators';
import { Subscription, timer } from 'rxjs';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import {
    ClrDatagridComparatorInterface,
    ReplicationJob,
    ReplicationTasks,
} from '../../../../../shared/services';
import {
    CURRENT_BASE_HREF,
    CustomComparator,
    doFiltering,
    doSorting,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../../shared/units/utils';
import { REFRESH_TIME_DIFFERENCE } from '../../../../../shared/entities/shared.const';
import { ClrDatagridStateInterface } from '@clr/angular';
import { ReplicationExecution } from '../../../../../../../ng-swagger-gen/models/replication-execution';
import { ReplicationService } from '../../../../../../../ng-swagger-gen/services';
import ListReplicationTasksParams = ReplicationService.ListReplicationTasksParams;
import { ReplicationTask } from '../../../../../../../ng-swagger-gen/models/replication-task';

const executionStatus = 'InProgress';
const STATUS_MAP = {
    Succeed: 'Succeeded',
};
const SUCCEED: string = 'Succeed';
@Component({
    selector: 'replication-tasks',
    templateUrl: './replication-tasks.component.html',
    styleUrls: ['./replication-tasks.component.scss'],
})
export class ReplicationTasksComponent implements OnInit, OnDestroy {
    isOpenFilterTag: boolean;
    inProgress: boolean = false;
    currentPage: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.REPLICATION_TASKS_COMPONENT
    );
    totalCount: number;
    loading = true;
    searchTask: string;
    defaultFilter = 'resourceType';
    tasks: ReplicationTask[];
    taskItem: ReplicationTasks[] = [];
    tasksCopy: ReplicationTasks[] = [];
    stopOnGoing: boolean;
    execution: ReplicationExecution;
    timerDelay: Subscription;
    executionId: string;
    startTimeComparator: ClrDatagridComparatorInterface<ReplicationTask> =
        new CustomComparator<ReplicationJob>('start_time', 'date');
    endTimeComparator: ClrDatagridComparatorInterface<ReplicationTask> =
        new CustomComparator<ReplicationJob>('end_time', 'date');
    tasksTimeout: any;
    constructor(
        private translate: TranslateService,
        private router: Router,
        private replicationService: ReplicationService,
        private errorHandler: ErrorHandler,
        private route: ActivatedRoute
    ) {}

    ngOnInit(): void {
        this.searchTask = '';
        this.executionId = this.route.snapshot.params['id'];
        const resolverData = this.route.snapshot.data;
        if (resolverData) {
            this.execution = <ReplicationExecution>(
                resolverData['replicationTasksRoutingResolver']
            );
            this.clrLoadPage();
        }
    }
    getExecutionDetail(): void {
        this.inProgress = true;
        if (this.executionId) {
            this.replicationService
                .getReplicationExecution({
                    id: +this.executionId,
                })
                .pipe(finalize(() => (this.inProgress = false)))
                .subscribe(
                    res => {
                        this.execution = res;
                        this.clrLoadPage();
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }

    clrLoadPage(): void {
        if (!this.timerDelay) {
            this.timerDelay = timer(
                REFRESH_TIME_DIFFERENCE,
                REFRESH_TIME_DIFFERENCE
            ).subscribe(() => {
                if (this.execution['status'] === executionStatus) {
                    this.getExecutionDetail();
                } else {
                    this.timerDelay.unsubscribe();
                    this.timerDelay = null;
                }
            });
        }
    }

    public get trigger(): string {
        return this.execution && this.execution['trigger']
            ? this.execution['trigger']
            : '';
    }

    public get startTime(): string {
        return this.execution && this.execution['start_time']
            ? this.execution['start_time']
            : null;
    }

    public get successNum(): number {
        return this.execution && this.execution['succeed'];
    }

    public get failedNum(): number {
        return this.execution && this.execution['failed'];
    }

    public get progressNum(): number {
        return this.execution && this.execution['in_progress'];
    }

    public get stoppedNum(): number {
        return this.execution && this.execution['stopped'];
    }

    stopJob() {
        this.stopOnGoing = true;
        this.replicationService
            .stopReplication({
                id: +this.executionId,
            })
            .subscribe(
                response => {
                    this.stopOnGoing = false;
                    this.getExecutionDetail();
                    this.translate
                        .get('REPLICATION.STOP_SUCCESS', {
                            param: this.executionId,
                        })
                        .subscribe((res: string) => {
                            this.errorHandler.info(res);
                        });
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }

    viewLog(taskId: number | string): string {
        return (
            CURRENT_BASE_HREF +
            '/replication' +
            '/executions/' +
            this.executionId +
            '/tasks/' +
            taskId +
            '/log'
        );
    }

    ngOnDestroy() {
        if (this.timerDelay) {
            this.timerDelay.unsubscribe();
        }
        if (this.tasksTimeout) {
            clearTimeout(this.tasksTimeout);
            this.tasksTimeout = null;
        }
    }

    clrLoadTasks(withLoading: boolean, state: ClrDatagridStateInterface): void {
        if (!state || !state.page || !this.executionId) {
            return;
        }
        if (state && state.page && state.page.size) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.REPLICATION_TASKS_COMPONENT,
                this.pageSize
            );
        }
        const param: ListReplicationTasksParams = {
            id: +this.executionId,
            page: this.currentPage,
            pageSize: this.pageSize,
            sort: getSortingString(state),
        };
        if (this.searchTask && this.searchTask !== '') {
            if (
                this.searchTask === STATUS_MAP.Succeed &&
                this.defaultFilter === 'status'
            ) {
                // convert 'Succeeded' to 'Succeed'
                param[this.defaultFilter] = SUCCEED;
            } else {
                param[this.defaultFilter] = this.searchTask;
            }
        }
        if (withLoading) {
            this.loading = true;
        }
        this.replicationService
            .listReplicationTasksResponse(param)
            .pipe(
                finalize(() => {
                    this.loading = false;
                })
            )
            .subscribe(
                res => {
                    if (res.headers) {
                        let xHeader: string = res.headers.get('X-Total-Count');
                        if (xHeader) {
                            this.totalCount = parseInt(xHeader, 0);
                        }
                    }
                    this.tasks = res.body; // Keep the data
                    // Do customising filtering and sorting
                    this.tasks = doFiltering<ReplicationTask>(
                        this.tasks,
                        state
                    );
                    this.tasks = doSorting<ReplicationTask>(this.tasks, state);
                    let count: number = 0;
                    if (this.tasks?.length) {
                        this.tasks.forEach(item => {
                            if (item.status === executionStatus) {
                                count++;
                            }
                        });
                    }
                    if (
                        count > 0 ||
                        this.execution?.status === executionStatus
                    ) {
                        if (!this.tasksTimeout) {
                            this.tasksTimeout = setTimeout(() => {
                                this.clrLoadTasks(false, {
                                    page: {},
                                });
                            }, REFRESH_TIME_DIFFERENCE);
                        }
                    }
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }
    onBack(): void {
        this.router.navigate(['harbor', 'replications']);
    }

    selectFilter($event: any): void {
        this.defaultFilter = $event['target'].value;
        this.doSearch(this.searchTask);
    }

    // refresh icon
    refreshTasks(): void {
        this.currentPage = 1;
        let state: ClrDatagridStateInterface = {
            page: {},
        };
        this.clrLoadTasks(true, state);
    }

    public doSearch(value: string): void {
        this.currentPage = 1;
        this.searchTask = value.trim();
        let state: ClrDatagridStateInterface = {
            page: {},
        };
        this.clrLoadTasks(true, state);
    }

    openFilter(isOpen: boolean): void {
        this.isOpenFilterTag = isOpen;
    }

    getStatusStr(status: string): string {
        if (STATUS_MAP && STATUS_MAP[status]) {
            return STATUS_MAP[status];
        }
        return status;
    }

    canStop() {
        return this.execution && this.execution.status === 'InProgress';
    }
}
