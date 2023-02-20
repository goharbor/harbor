import { Component, ViewChild, OnInit, OnDestroy } from '@angular/core';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { CronScheduleComponent } from '../../../../../shared/components/cron-schedule';
import { OriginCron } from '../../../../../shared/services';
import { finalize } from 'rxjs/operators';
import { ScheduleType } from '../../../../../shared/entities/shared.const';
import { GcComponent } from '../../gc-page/gc/gc.component';
import { PurgeService } from '../../../../../../../ng-swagger-gen/services/purge.service';
import { ExecHistory } from '../../../../../../../ng-swagger-gen/models/exec-history';
import {
    JOB_STATUS,
    REFRESH_STATUS_TIME_DIFFERENCE,
    RETENTION_OPERATIONS,
    RETENTION_OPERATIONS_I18N_MAP,
    RetentionTimeUnit,
} from '../../clearing-job-interfact';
import { clone } from '../../../../../shared/units/utils';
import { PurgeHistoryComponent } from '../history/purge-history.component';
import { NgForm } from '@angular/forms';

const ONE_MINUTE: number = 60000;
const ONE_DAY: number = 24;
const MAX_RETENTION_DAYS: number = 10000;

@Component({
    selector: 'app-set-job',
    templateUrl: './set-job.component.html',
    styleUrls: ['./set-job.component.scss'],
})
export class SetJobComponent implements OnInit, OnDestroy {
    originCron: OriginCron;
    disableGC: boolean = false;
    getLabelCurrent = 'CLEARANCES.SCHEDULE_TO_PURGE';
    loadingGcStatus = false;
    @ViewChild(CronScheduleComponent)
    cronScheduleComponent: CronScheduleComponent;
    dryRunOnGoing: boolean = false;
    lastCompletedTime: string;
    loadingLastCompletedTime: boolean = false;
    isDryRun: boolean = false;
    nextScheduledTime: string;
    statusTimeout: any;

    retentionTime: number;
    retentionUnit: string = RetentionTimeUnit.DAYS;
    operations: string[] = clone(RETENTION_OPERATIONS);
    selectedOperations: string[] = clone(RETENTION_OPERATIONS);
    @ViewChild(PurgeHistoryComponent)
    purgeHistoryComponent: PurgeHistoryComponent;
    @ViewChild('purgeForm')
    purgeForm: NgForm;
    constructor(
        private purgeService: PurgeService,
        private errorHandler: ErrorHandler
    ) {}

    ngOnInit() {
        this.getCurrentSchedule(true);
        this.getStatus();
    }
    ngOnDestroy() {
        if (this.statusTimeout) {
            clearTimeout(this.statusTimeout);
            this.statusTimeout = null;
        }
    }
    // get the latest non-dry-run execution to get the status
    getStatus() {
        this.loadingLastCompletedTime = true;
        this.purgeService
            .getPurgeHistory({
                page: 1,
                pageSize: 1,
                sort: '-update_time',
            })
            .subscribe(res => {
                if (res?.length) {
                    this.isDryRun = JSON.parse(res[0]?.job_parameters).dry_run;
                    this.lastCompletedTime = res[0]?.update_time;
                    if (
                        res[0]?.job_status === JOB_STATUS.RUNNING ||
                        res[0]?.job_status === JOB_STATUS.PENDING
                    ) {
                        this.statusTimeout = setTimeout(() => {
                            this.getStatus();
                        }, REFRESH_STATUS_TIME_DIFFERENCE);
                        return;
                    }
                }
                this.loadingLastCompletedTime = false;
            });
    }
    getCurrentSchedule(withLoading: boolean) {
        if (withLoading) {
            this.loadingGcStatus = true;
        }
        this.purgeService
            .getPurgeSchedule()
            .pipe(
                finalize(() => {
                    this.loadingGcStatus = false;
                })
            )
            .subscribe({
                next: schedule => {
                    this.initSchedule(schedule);
                },
                error: error => {
                    this.errorHandler.error(error);
                },
            });
    }

    initSchedule(purgeHistory: ExecHistory) {
        this.nextScheduledTime = purgeHistory?.schedule?.next_scheduled_time
            ? purgeHistory?.schedule?.next_scheduled_time
            : null;
        if (purgeHistory && purgeHistory.schedule) {
            this.originCron = {
                type: purgeHistory.schedule.type,
                cron: purgeHistory.schedule.cron,
            };
            if (purgeHistory && purgeHistory.job_parameters) {
                const obj = JSON.parse(purgeHistory.job_parameters);
                if (obj?.include_operations) {
                    this.selectedOperations =
                        obj?.include_operations?.split(',');
                } else {
                    this.selectedOperations = [];
                }
                if (
                    obj?.audit_retention_hour > ONE_DAY &&
                    obj?.audit_retention_hour % ONE_DAY === 0
                ) {
                    this.retentionTime = obj?.audit_retention_hour / ONE_DAY;
                    this.retentionUnit = RetentionTimeUnit.DAYS;
                } else {
                    this.retentionTime = obj?.audit_retention_hour;
                    this.retentionUnit = RetentionTimeUnit.HOURS;
                }
            } else {
                this.retentionTime = null;
                this.selectedOperations = clone(RETENTION_OPERATIONS);
                this.retentionUnit = RetentionTimeUnit.DAYS;
            }
        } else {
            this.originCron = {
                type: ScheduleType.NONE,
                cron: '',
            };
        }
    }

    gcNow(): void {
        this.disableGC = true;
        setTimeout(() => {
            this.enableGc();
        }, ONE_MINUTE);
        const retentionTime: number =
            this.retentionUnit === RetentionTimeUnit.DAYS
                ? this.retentionTime * 24
                : this.retentionTime;
        this.purgeService
            .createPurgeSchedule({
                schedule: {
                    parameters: {
                        audit_retention_hour: +retentionTime,
                        include_operations: this.selectedOperations.join(','),
                        dry_run: false,
                    },
                    schedule: {
                        type: ScheduleType.MANUAL,
                    },
                },
            })
            .subscribe({
                next: response => {
                    this.errorHandler.info('CLEARANCES.PURGE_NOW_SUCCESS');
                    this.refresh();
                },
                error: error => {
                    this.errorHandler.error(error);
                },
            });
    }

    dryRun() {
        this.dryRunOnGoing = true;
        const retentionTime: number =
            this.retentionUnit === RetentionTimeUnit.DAYS
                ? this.retentionTime * 24
                : this.retentionTime;
        this.purgeService
            .createPurgeSchedule({
                schedule: {
                    parameters: {
                        audit_retention_hour: +retentionTime,
                        include_operations: this.selectedOperations.join(','),
                        dry_run: true,
                    },
                    schedule: {
                        type: ScheduleType.MANUAL,
                    },
                },
            })
            .pipe(finalize(() => (this.dryRunOnGoing = false)))
            .subscribe({
                next: response => {
                    this.errorHandler.info('GC.DRY_RUN_SUCCESS');
                    this.refresh();
                },
                error: error => {
                    this.errorHandler.error(error);
                },
            });
    }

    private enableGc() {
        this.disableGC = false;
    }

    saveGcSchedule(cron: string) {
        const retentionTime: number =
            this.retentionUnit === RetentionTimeUnit.DAYS
                ? this.retentionTime * 24
                : this.retentionTime;
        if (this.originCron && this.originCron.type === ScheduleType.NONE) {
            // no schedule, then create
            this.purgeService
                .createPurgeSchedule({
                    schedule: {
                        parameters: {
                            audit_retention_hour: +retentionTime,
                            include_operations:
                                this.selectedOperations.join(','),
                            dry_run: false,
                        },
                        schedule: {
                            type: GcComponent.getScheduleType(cron),
                            cron: cron,
                        },
                    },
                })
                .subscribe({
                    next: response => {
                        this.errorHandler.info(
                            'CLEARANCES.PURGE_SCHEDULE_RESET'
                        );
                        this.cronScheduleComponent.resetSchedule();
                        this.getCurrentSchedule(false); // refresh schedule
                    },
                    error: error => {
                        this.errorHandler.error(error);
                    },
                });
        } else {
            this.purgeService
                .updatePurgeSchedule({
                    schedule: {
                        parameters: {
                            audit_retention_hour: +retentionTime,
                            include_operations:
                                this.selectedOperations.join(','),
                            dry_run: false,
                        },
                        schedule: {
                            type: GcComponent.getScheduleType(cron),
                            cron: cron,
                        },
                    },
                })
                .subscribe({
                    next: response => {
                        this.errorHandler.info(
                            'CLEARANCES.PURGE_SCHEDULE_RESET'
                        );
                        this.cronScheduleComponent.resetSchedule();
                        this.getCurrentSchedule(false); // refresh schedule
                    },
                    error: error => {
                        this.errorHandler.error(error);
                    },
                });
        }
    }
    hasOperation(operation: string): boolean {
        return this.selectedOperations?.indexOf(operation) !== -1;
    }
    operationsToText(operation: string): string {
        if (RETENTION_OPERATIONS_I18N_MAP[operation]) {
            return RETENTION_OPERATIONS_I18N_MAP[operation];
        }
        return operation;
    }
    setOperation(operation: string) {
        if (this.selectedOperations.indexOf(operation) === -1) {
            this.selectedOperations.push(operation);
        } else {
            this.selectedOperations.splice(
                this.selectedOperations.findIndex(item => item === operation),
                1
            );
        }
    }
    refresh() {
        this.getStatus();
        this.purgeHistoryComponent?.refresh();
    }
    isValid(): boolean {
        if (this.cronScheduleComponent?.scheduleType === ScheduleType.NONE) {
            return true;
        }
        return !(
            this.purgeForm?.invalid || !(this.selectedOperations?.length > 0)
        );
    }
    isRetentionTimeValid() {
        if (this.retentionUnit === RetentionTimeUnit.DAYS) {
            return (
                this.retentionTime > 0 &&
                this.retentionTime <= MAX_RETENTION_DAYS
            );
        }
        return (
            this.retentionTime > 0 &&
            this.retentionTime <= MAX_RETENTION_DAYS * ONE_DAY
        );
    }
}
