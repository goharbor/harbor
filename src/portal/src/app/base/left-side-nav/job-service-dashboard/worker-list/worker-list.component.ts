import { Component, OnDestroy, OnInit } from '@angular/core';
import { ClrDatagridStateInterface } from '@clr/angular/data/datagrid/interfaces/state.interface';
import { Worker } from 'ng-swagger-gen/models';
import { WorkerPool } from 'ng-swagger-gen/models/worker-pool';
import { JobserviceService } from 'ng-swagger-gen/services';
import { finalize, forkJoin, Subscription } from 'rxjs';
import { MessageHandlerService } from 'src/app/shared/services/message-handler.service';
import {
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from 'src/app/shared/units/utils';
import { ConfirmationMessage } from '../../../global-confirmation-dialog/confirmation-message';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../../shared/entities/shared.const';
import { ConfirmationDialogService } from '../../../global-confirmation-dialog/confirmation-dialog.service';
import {
    EventService,
    HarborEvent,
} from '../../../../services/event-service/event.service';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { errorHandler } from '../../../../shared/units/shared.utils';

@Component({
    selector: 'app-worker-list',
    templateUrl: './worker-list.component.html',
    styleUrls: ['./worker-list.component.scss'],
})
export class WorkerListComponent implements OnInit, OnDestroy {
    loadingPools: boolean = false;
    selectedPool: WorkerPool;
    pools: WorkerPool[] = [];
    loadingWorkers: boolean = false;
    workers: Worker[] = [];
    selected: Worker[] = [];

    poolPageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.WORKER_LIST_COMPONENT_POOL,
        5
    );

    workerPageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.WORKER_LIST_COMPONENT_WORKER
    );
    loadingFree: boolean = false;
    confirmSub: Subscription;
    eventSub: Subscription;
    constructor(
        private jobServiceService: JobserviceService,
        private messageHandlerService: MessageHandlerService,
        private operateDialogService: ConfirmationDialogService,
        private eventService: EventService,
        private operationService: OperationService
    ) {}

    ngOnInit(): void {
        this.getPools();
        this.initSub();
        this.initEventSub();
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
                                ConfirmationTargets.FREE_SPECIFIED_WORKERS
                            ) {
                                this.executeFreeWorkers();
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
                    this.selectionChanged();
                }
            );
        }
    }

    getPools() {
        this.loadingPools = true;
        this.jobServiceService
            .getWorkerPools()
            .pipe(finalize(() => (this.loadingPools = false)))
            .subscribe({
                next: res => {
                    this.pools = res;
                    if (res?.length) {
                        this.selectedPool = res[0];
                        this.selectionChanged();
                    }
                },
                error: err => {
                    this.messageHandlerService.error(err);
                },
            });
    }

    selectionChanged() {
        this.loadingWorkers = true;
        this.jobServiceService
            .getWorkers({ poolId: this.selectedPool?.worker_pool_id })
            .pipe(finalize(() => (this.loadingWorkers = false)))
            .subscribe({
                next: res => {
                    this.workers = res;
                },
                error: err => {
                    this.messageHandlerService.error(err);
                },
            });
    }

    string(v: any) {
        if (v) {
            return JSON.stringify(v);
        }
        return null;
    }

    clrLoadPool(state: ClrDatagridStateInterface): void {
        if (state?.page?.size) {
            this.poolPageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.WORKER_LIST_COMPONENT_POOL,
                this.poolPageSize
            );
        }
    }

    clrLoadWorker(state: ClrDatagridStateInterface): void {
        if (state?.page?.size) {
            this.workerPageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.WORKER_LIST_COMPONENT_POOL,
                this.workerPageSize
            );
        }
    }

    canFree(): boolean {
        if (this.selected?.length) {
            let flag: boolean = true;
            this.selected.forEach(item => {
                if (!item.job_id) {
                    flag = false;
                }
            });
            return flag;
        }
        return false;
    }

    freeWorker() {
        const workers: string = this.selected.map(item => item.id).join(',');
        const deletionMessage = new ConfirmationMessage(
            'JOB_SERVICE_DASHBOARD.CONFIRM_FREE_WORKERS',
            'JOB_SERVICE_DASHBOARD.CONFIRM_FREE_WORKERS_CONTENT',
            workers,
            this.selected,
            ConfirmationTargets.FREE_SPECIFIED_WORKERS,
            ConfirmationButtons.CONFIRM_CANCEL
        );
        this.operateDialogService.openComfirmDialog(deletionMessage);
    }

    executeFreeWorkers() {
        this.loadingFree = true;
        const operationMessage = new OperateInfo();
        operationMessage.name =
            'JOB_SERVICE_DASHBOARD.OPERATION_FREE_SPECIFIED_WORKERS';
        operationMessage.state = OperationState.progressing;
        operationMessage.data.name = this.selected
            .map(item => item.id)
            .join(',');
        this.operationService.publishInfo(operationMessage);
        forkJoin(
            this.selected.map(item => {
                return this.jobServiceService.stopRunningJob({
                    jobId: item.job_id,
                });
            })
        )
            .pipe(finalize(() => (this.loadingFree = false)))
            .subscribe({
                next: res => {
                    this.messageHandlerService.info(
                        'JOB_SERVICE_DASHBOARD.FREE_WORKER_SUCCESS'
                    );
                    this.selectionChanged();
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
}
