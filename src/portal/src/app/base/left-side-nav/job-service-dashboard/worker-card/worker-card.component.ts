import { Component, OnDestroy, OnInit } from '@angular/core';
import { Worker } from 'ng-swagger-gen/models';
import { JobserviceService } from 'ng-swagger-gen/services';
import { finalize, Subscription } from 'rxjs';
import { ConfirmationDialogService } from 'src/app/base/global-confirmation-dialog/confirmation-dialog.service';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from 'src/app/shared/entities/shared.const';
import { MessageHandlerService } from 'src/app/shared/services/message-handler.service';
import { All, INTERVAL } from '../job-service-dashboard.interface';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { errorHandler } from '../../../../shared/units/shared.utils';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';

@Component({
    selector: 'app-worker-card',
    templateUrl: './worker-card.component.html',
    styleUrls: ['./worker-card.component.scss'],
})
export class WorkerCardComponent implements OnInit, OnDestroy {
    denominator: number = 0;
    statusTimeout: any;
    loadingFreeAll: boolean = false;
    confirmSub: Subscription;
    busyWorkers: Worker[] = [];

    constructor(
        private operateDialogService: ConfirmationDialogService,
        private jobServiceService: JobserviceService,
        private messageHandlerService: MessageHandlerService,
        private operationService: OperationService,
        private jobServiceDashboardSharedDataService: JobServiceDashboardSharedDataService
    ) {}

    ngOnInit(): void {
        this.loopWorkerStatus();
        this.initSub();
    }
    ngOnDestroy(): void {
        if (this.statusTimeout) {
            clearTimeout(this.statusTimeout);
            this.statusTimeout = null;
        }
        if (this.confirmSub) {
            this.confirmSub.unsubscribe();
            this.confirmSub = null;
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
                                ConfirmationTargets.FREE_ALL_WORKERS
                            ) {
                                this.executeFreeAll();
                            }
                        }
                    }
                );
        }
    }

    loopWorkerStatus(isAutoRefresh?: boolean) {
        this.jobServiceDashboardSharedDataService
            .retrieveAllWorkers(isAutoRefresh)
            .subscribe(res => {
                if (res) {
                    this.denominator = res.length;
                    this.busyWorkers = [];
                    res.forEach(item => {
                        if (item.job_id) {
                            this.busyWorkers.push(item);
                        }
                    });
                    this.statusTimeout = setTimeout(() => {
                        this.loopWorkerStatus(true);
                    }, INTERVAL);
                }
            });
    }

    freeAll() {
        this.operateDialogService.openComfirmDialog({
            data: undefined,
            param: null,
            title: 'JOB_SERVICE_DASHBOARD.CONFIRM_FREE_ALL',
            message: 'JOB_SERVICE_DASHBOARD.CONFIRM_FREE_ALL_CONTENT',
            targetId: ConfirmationTargets.FREE_ALL_WORKERS,
            buttons: ConfirmationButtons.CONFIRM_CANCEL,
        });
    }

    executeFreeAll() {
        this.loadingFreeAll = true;
        const operationMessage = new OperateInfo();
        operationMessage.name = 'JOB_SERVICE_DASHBOARD.OPERATION_FREE_ALL';
        operationMessage.state = OperationState.progressing;
        this.operationService.publishInfo(operationMessage);
        this.jobServiceService
            .stopRunningJob({ jobId: All.ALL_WORKERS })
            .pipe(finalize(() => (this.loadingFreeAll = false)))
            .subscribe({
                next: res => {
                    this.messageHandlerService.info(
                        'JOB_SERVICE_DASHBOARD.FREE_ALL_SUCCESS'
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
        if (this.statusTimeout) {
            clearTimeout(this.statusTimeout);
            this.statusTimeout = null;
        }
        this.loopWorkerStatus();
    }
}
