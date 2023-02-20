import { Component, OnDestroy, OnInit } from '@angular/core';
import { ConfirmationDialogService } from '../../../global-confirmation-dialog/confirmation-dialog.service';
import { JobserviceService } from '../../../../../../ng-swagger-gen/services/jobservice.service';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { finalize } from 'rxjs/operators';
import {
    INTERVAL,
    JobType,
    PendingJobsActions,
    ScheduleExecuteBtnString,
    ScheduleStatusString,
} from '../job-service-dashboard.interface';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../../shared/entities/shared.const';
import { Subscription } from 'rxjs';
import { EventService } from '../../../../services/event-service/event.service';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { errorHandler } from '../../../../shared/units/shared.utils';
import { ScheduleService } from '../../../../../../ng-swagger-gen/services/schedule.service';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';

@Component({
    selector: 'app-schedule-card',
    templateUrl: './schedule-card.component.html',
    styleUrls: ['./schedule-card.component.scss'],
})
export class ScheduleCardComponent implements OnInit, OnDestroy {
    isPaused: boolean = false;
    loadingStatus: boolean = false;
    confirmSub: Subscription;
    statusTimeout: any;
    clickOnGoing: boolean = false;
    constructor(
        private operateDialogService: ConfirmationDialogService,
        private jobServiceService: JobserviceService,
        private messageHandlerService: MessageHandlerService,
        private eventService: EventService,
        private operationService: OperationService,
        private scheduleService: ScheduleService,
        private jobServiceDashboardSharedDataService: JobServiceDashboardSharedDataService
    ) {}

    ngOnInit() {
        this.initSub();
        this.loopGetStatus(true);
        this.getScheduleCount();
    }

    ngOnDestroy() {
        if (this.statusTimeout) {
            clearTimeout(this.statusTimeout);
            this.statusTimeout = null;
        }
        if (this.confirmSub) {
            this.confirmSub.unsubscribe();
            this.confirmSub = null;
        }
    }

    get scheduleCount(): number {
        return (
            this.jobServiceDashboardSharedDataService.getScheduleListResponse()
                ?.total | 0
        );
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
                                ConfirmationTargets.RESUME_ALL_SCHEDULES
                            ) {
                                this.execute(PendingJobsActions.RESUME);
                            }
                            if (
                                message.source ===
                                ConfirmationTargets.PAUSE_ALL_SCHEDULES
                            ) {
                                this.execute(PendingJobsActions.PAUSE);
                            }
                        }
                    }
                );
        }
    }

    loopGetStatus(withLoading?: boolean) {
        if (withLoading) {
            this.loadingStatus = true;
        }
        this.scheduleService
            .getSchedulePaused({
                jobType: JobType.ALL,
            })
            .pipe(finalize(() => (this.loadingStatus = false)))
            .subscribe(res => {
                this.isPaused = res?.paused;
                this.statusTimeout = setTimeout(() => {
                    this.loopGetStatus();
                    this.getScheduleCount();
                }, INTERVAL);
            });
    }

    getScheduleCount() {
        this.jobServiceDashboardSharedDataService
            .retrieveScheduleListResponse()
            .subscribe({
                next: res => {},
                error: err => {
                    this.messageHandlerService.error(err);
                },
            });
    }

    btnText(): string {
        if (this.isPaused) {
            return ScheduleExecuteBtnString.RESUME_ALL;
        }
        return ScheduleExecuteBtnString.PAUSE_ALL;
    }

    statusStr(): string {
        if (this.isPaused) {
            return ScheduleStatusString.PAUSED;
        }
        return ScheduleStatusString.RUNNING;
    }

    pauseOrResume() {
        if (this.isPaused) {
            this.operateDialogService.openComfirmDialog({
                data: undefined,
                param: null,
                title: 'JOB_SERVICE_DASHBOARD.CONFIRM_RESUMING_ALL',
                message: 'JOB_SERVICE_DASHBOARD.CONFIRM_RESUMING_ALL_CONTENT',
                targetId: ConfirmationTargets.RESUME_ALL_SCHEDULES,
                buttons: ConfirmationButtons.CONFIRM_CANCEL,
            });
        } else {
            this.operateDialogService.openComfirmDialog({
                data: undefined,
                param: null,
                title: 'JOB_SERVICE_DASHBOARD.CONFIRM_PAUSING_ALL',
                message: 'JOB_SERVICE_DASHBOARD.CONFIRM_PAUSING_ALL_CONTENT',
                targetId: ConfirmationTargets.PAUSE_ALL_SCHEDULES,
                buttons: ConfirmationButtons.CONFIRM_CANCEL,
            });
        }
    }

    execute(action: PendingJobsActions) {
        this.clickOnGoing = true;
        const operationMessage = new OperateInfo();
        operationMessage.name = this.isPaused
            ? 'JOB_SERVICE_DASHBOARD.OPERATION_RESUME_SCHEDULE'
            : 'JOB_SERVICE_DASHBOARD.OPERATION_PAUSE_SCHEDULE';
        operationMessage.state = OperationState.progressing;
        this.operationService.publishInfo(operationMessage);
        this.jobServiceService
            .actionPendingJobs({
                jobType: JobType.SCHEDULER,
                actionRequest: {
                    action: action,
                },
            })
            .pipe(finalize(() => (this.clickOnGoing = false)))
            .subscribe({
                next: res => {
                    if (this.isPaused) {
                        this.messageHandlerService.info(
                            'JOB_SERVICE_DASHBOARD.RESUME_ALL_SUCCESS'
                        );
                    } else {
                        this.messageHandlerService.info(
                            'JOB_SERVICE_DASHBOARD.PAUSE_ALL_SUCCESS'
                        );
                    }
                    this.refreshNow();
                    this.jobServiceDashboardSharedDataService
                        .retrieveJobQueues()
                        .subscribe();
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
        this.loopGetStatus(true);
    }
}
