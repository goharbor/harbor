import { Component, OnDestroy, OnInit } from '@angular/core';
import { ClrDatagridStateInterface } from '@clr/angular/data/datagrid/interfaces/state.interface';
import {
    durationStr,
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';
import { JobserviceService } from '../../../../../../ng-swagger-gen/services/jobservice.service';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { finalize } from 'rxjs/operators';
import { JobQueue } from '../../../../../../ng-swagger-gen/models/job-queue';
import { NO, YES } from '../../clearing-job/clearing-job-interfact';
import { PendingJobsActions } from '../job-service-dashboard.interface';
import { forkJoin, Subscription } from 'rxjs';
import { ConfirmationDialogService } from '../../../global-confirmation-dialog/confirmation-dialog.service';
import {
    EventService,
    HarborEvent,
} from '../../../../services/event-service/event.service';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from 'src/app/shared/entities/shared.const';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import { errorHandler } from '../../../../shared/units/shared.utils';

@Component({
    selector: 'app-pending-job-list',
    templateUrl: './pending-job-list.component.html',
    styleUrls: ['./pending-job-list.component.scss'],
})
export class PendingListComponent implements OnInit, OnDestroy {
    loading: boolean = false;
    selectedRows: JobQueue[] = [];
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.PENDING_LIST_COMPONENT
    );
    jobQueue: JobQueue[] = [];
    loadingStop: boolean = false;
    loadingPause: boolean = false;
    loadingResume: boolean = false;
    confirmSub: Subscription;
    eventSub: Subscription;
    constructor(
        private jobServiceService: JobserviceService,
        private messageHandlerService: MessageHandlerService,
        private operateDialogService: ConfirmationDialogService,
        private eventService: EventService,
        private operationService: OperationService
    ) {}

    ngOnInit() {
        this.getJobs();
        this.initEventSub();
        this.initSub();
    }
    ngOnDestroy() {
        if (this.confirmSub) {
            this.confirmSub.unsubscribe();
            this.confirmSub = null;
        }
        if (this.eventSub) {
            this.eventSub.unsubscribe();
            this.eventSub = null;
        }
    }

    initSub() {
        if (!this.confirmSub) {
            this.confirmSub =
                this.operateDialogService.confirmationConfirm$.subscribe(
                    message => {
                        if (
                            message &&
                            message.state === ConfirmationState.CONFIRMED
                        ) {
                            if (
                                message.source ===
                                ConfirmationTargets.STOPS_JOBS
                            ) {
                                this.executeStop();
                            }
                            if (
                                message.source ===
                                ConfirmationTargets.PAUSE_JOBS
                            ) {
                                this.executePause();
                            }
                            if (
                                message.source ===
                                ConfirmationTargets.RESUME_JOBS
                            ) {
                                this.executeResume();
                            }
                        }
                    }
                );
        }
    }

    initEventSub() {
        if (!this.eventSub) {
            this.eventSub = this.eventService.subscribe(
                HarborEvent.REFRESH_JOB_SERVICE_DASHBOARD,
                () => {
                    this.getJobs();
                }
            );
        }
    }

    getJobs() {
        this.selectedRows = [];
        this.loading = true;
        this.jobServiceService
            .listJobQueues()
            .pipe(finalize(() => (this.loading = false)))
            .subscribe({
                next: res => {
                    this.jobQueue = res;
                },
                error: err => {
                    this.messageHandlerService.error(err);
                },
            });
    }

    clrLoad(state: ClrDatagridStateInterface): void {
        if (state?.page?.size) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.PENDING_LIST_COMPONENT,
                this.pageSize
            );
        }
    }

    isPaused(paused: boolean): string {
        if (paused) {
            return YES;
        }
        return NO;
    }
    getDuration(v: number): string {
        if (v) {
            return durationStr(v * 1000);
        }
        return null;
    }

    canPause(): boolean {
        if (this.selectedRows?.length) {
            return !this.selectedRows.some(item => item.paused);
        }
        return false;
    }

    canResume(): boolean {
        if (this.selectedRows?.length) {
            return !this.selectedRows.some(item => !item.paused);
        }
        return false;
    }

    stop() {
        const jobs: string = this.selectedRows
            .map(item => item.job_type)
            .join(',');
        this.operateDialogService.openComfirmDialog({
            data: undefined,
            param: jobs,
            title: 'JOB_SERVICE_DASHBOARD.CONFIRM_STOPPING_JOBS',
            message: 'JOB_SERVICE_DASHBOARD.CONFIRM_STOPPING_JOBS_CONTENT',
            targetId: ConfirmationTargets.STOPS_JOBS,
            buttons: ConfirmationButtons.CONFIRM_CANCEL,
        });
    }

    pause() {
        const jobs: string = this.selectedRows
            .map(item => item.job_type)
            .join(',');
        this.operateDialogService.openComfirmDialog({
            data: undefined,
            param: jobs,
            title: 'JOB_SERVICE_DASHBOARD.CONFIRM_PAUSING_JOBS',
            message: 'JOB_SERVICE_DASHBOARD.CONFIRM_PAUSING_JOBS_CONTENT',
            targetId: ConfirmationTargets.PAUSE_JOBS,
            buttons: ConfirmationButtons.CONFIRM_CANCEL,
        });
    }

    resume() {
        const jobs: string = this.selectedRows
            .map(item => item.job_type)
            .join(',');
        this.operateDialogService.openComfirmDialog({
            data: undefined,
            param: jobs,
            title: 'JOB_SERVICE_DASHBOARD.CONFIRM_RESUMING_JOBS',
            message: 'JOB_SERVICE_DASHBOARD.CONFIRM_RESUMING_JOBS_CONTENT',
            targetId: ConfirmationTargets.RESUME_JOBS,
            buttons: ConfirmationButtons.CONFIRM_CANCEL,
        });
    }

    executeStop() {
        this.loadingStop = true;
        const operationMessage = new OperateInfo();
        operationMessage.name =
            'JOB_SERVICE_DASHBOARD.OPERATION_STOP_SPECIFIED_QUEUES';
        operationMessage.state = OperationState.progressing;
        operationMessage.data.name = this.selectedRows
            .map(item => item.job_type)
            .join(',');
        this.operationService.publishInfo(operationMessage);
        forkJoin(
            this.selectedRows.map(item => {
                return this.jobServiceService.actionPendingJobs({
                    jobType: item.job_type,
                    actionRequest: {
                        action: PendingJobsActions.STOP,
                    },
                });
            })
        )
            .pipe(finalize(() => (this.loadingStop = false)))
            .subscribe({
                next: res => {
                    this.messageHandlerService.info(
                        'JOB_SERVICE_DASHBOARD.STOP_SUCCESS'
                    );
                    this.getJobs();
                    operateChanges(operationMessage, OperationState.success);
                },
                error: err => {
                    this.messageHandlerService.error(err);
                    operateChanges(
                        operationMessage,
                        OperationState.failure,
                        errorHandler(err)
                    );
                },
            });
    }

    executePause() {
        this.loadingPause = true;
        const operationMessage = new OperateInfo();
        operationMessage.name =
            'JOB_SERVICE_DASHBOARD.OPERATION_PAUSE_SPECIFIED_QUEUES';
        operationMessage.state = OperationState.progressing;
        operationMessage.data.name = this.selectedRows
            .map(item => item.job_type)
            .join(',');
        this.operationService.publishInfo(operationMessage);
        forkJoin(
            this.selectedRows.map(item => {
                return this.jobServiceService.actionPendingJobs({
                    jobType: item.job_type,
                    actionRequest: {
                        action: PendingJobsActions.PAUSE,
                    },
                });
            })
        )
            .pipe(finalize(() => (this.loadingPause = false)))
            .subscribe({
                next: res => {
                    operateChanges(operationMessage, OperationState.success);
                    this.messageHandlerService.info(
                        'JOB_SERVICE_DASHBOARD.PAUSE_SUCCESS'
                    );
                    this.getJobs();
                },
                error: err => {
                    this.messageHandlerService.error(err);
                    operateChanges(
                        operationMessage,
                        OperationState.failure,
                        errorHandler(err)
                    );
                },
            });
    }
    executeResume() {
        this.loadingResume = true;
        const operationMessage = new OperateInfo();
        operationMessage.name =
            'JOB_SERVICE_DASHBOARD.OPERATION_RESUME_SPECIFIED_QUEUES';
        operationMessage.state = OperationState.progressing;
        operationMessage.data.name = this.selectedRows
            .map(item => item.job_type)
            .join(',');
        this.operationService.publishInfo(operationMessage);
        forkJoin(
            this.selectedRows.map(item => {
                return this.jobServiceService.actionPendingJobs({
                    jobType: item.job_type,
                    actionRequest: {
                        action: PendingJobsActions.RESUME,
                    },
                });
            })
        )
            .pipe(finalize(() => (this.loadingResume = false)))
            .subscribe({
                next: res => {
                    operateChanges(operationMessage, OperationState.success);
                    this.messageHandlerService.info(
                        'JOB_SERVICE_DASHBOARD.RESUME_SUCCESS'
                    );
                    this.getJobs();
                },
                error: err => {
                    this.messageHandlerService.error(err);
                    operateChanges(
                        operationMessage,
                        OperationState.failure,
                        errorHandler(err)
                    );
                },
            });
    }
}
