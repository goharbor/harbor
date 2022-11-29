import { Component, OnDestroy, OnInit } from '@angular/core';
import { ConfirmationDialogService } from '../../../global-confirmation-dialog/confirmation-dialog.service';
import { JobserviceService } from '../../../../../../ng-swagger-gen/services/jobservice.service';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { finalize } from 'rxjs/operators';
import {
    INTERVAL,
    JobType,
    PendingJobsActions,
} from '../job-service-dashboard.interface';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../../shared/entities/shared.const';
import { Subscription } from 'rxjs';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { errorHandler } from '../../../../shared/units/shared.utils';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';

@Component({
    selector: 'app-pending-job-card',
    templateUrl: './pending-job-card.component.html',
    styleUrls: ['./pending-job-card.component.scss'],
})
export class PendingCardComponent implements OnInit, OnDestroy {
    loading: boolean = false;
    timeout: any;
    loadingStopAll: boolean = false;
    confirmSub: Subscription;
    constructor(
        private operateDialogService: ConfirmationDialogService,
        private jobServiceService: JobserviceService,
        private messageHandlerService: MessageHandlerService,
        private operationService: OperationService,
        private jobServiceDashboardSharedDataService: JobServiceDashboardSharedDataService
    ) {}

    ngOnInit() {
        this.loopGetPendingJobs(true);
        this.initSub();
    }
    ngOnDestroy() {
        if (this.timeout) {
            clearTimeout(this.timeout);
            this.timeout = null;
        }
        if (this.confirmSub) {
            this.confirmSub.unsubscribe();
            this.confirmSub = null;
        }
    }
    get jobQueue() {
        return this.jobServiceDashboardSharedDataService
            .getJobQueues()
            .sort((a, b) => {
                const ACount: number = a?.count | 0;
                const BCount: number = b?.count | 0;
                return BCount - ACount;
            });
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
                                ConfirmationTargets.STOP_ALL_PENDING_JOBS
                            ) {
                                this.executeStopAll();
                            }
                        }
                    }
                );
        }
    }

    loopGetPendingJobs(withLoading?: boolean) {
        if (withLoading) {
            this.loading = true;
        }
        this.jobServiceDashboardSharedDataService
            .retrieveJobQueues(true)
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(res => {
                this.timeout = setTimeout(() => {
                    this.loopGetPendingJobs();
                }, INTERVAL);
            });
    }

    total(): number {
        if (this.jobQueue?.length) {
            let count: number = 0;
            this.jobQueue.forEach((item, index) => {
                count += item?.count | 0;
            });
            return count;
        }
        return 0;
    }

    otherCount(): number {
        if (this.jobQueue?.length) {
            let count: number = 0;
            this.jobQueue.forEach((item, index) => {
                if (index > 1) {
                    count += item?.count | 0;
                }
            });
            return count;
        }
        return 0;
    }

    stopAll() {
        this.operateDialogService.openComfirmDialog({
            data: undefined,
            param: null,
            title: 'JOB_SERVICE_DASHBOARD.CONFIRM_STOP_ALL',
            message: 'JOB_SERVICE_DASHBOARD.CONFIRM_STOP_ALL_CONTENT',
            targetId: ConfirmationTargets.STOP_ALL_PENDING_JOBS,
            buttons: ConfirmationButtons.CONFIRM_CANCEL,
        });
    }

    executeStopAll() {
        this.loadingStopAll = true;
        const operationMessage = new OperateInfo();
        operationMessage.name =
            'JOB_SERVICE_DASHBOARD.OPERATION_STOP_ALL_QUEUES';
        operationMessage.state = OperationState.progressing;
        this.operationService.publishInfo(operationMessage);
        this.jobServiceService
            .actionPendingJobs({
                jobType: JobType.ALL,
                actionRequest: {
                    action: PendingJobsActions.STOP,
                },
            })
            .pipe(finalize(() => (this.loadingStopAll = false)))
            .subscribe({
                next: res => {
                    this.messageHandlerService.info(
                        'JOB_SERVICE_DASHBOARD.STOP_ALL_SUCCESS'
                    );
                    this.refreshNow();
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

    refreshNow() {
        if (this.timeout) {
            clearTimeout(this.timeout);
            this.timeout = null;
        }
        this.loopGetPendingJobs(true);
    }
}
