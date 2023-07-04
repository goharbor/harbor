import { Component, ViewChild, OnInit, OnDestroy } from '@angular/core';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { CronScheduleComponent } from '../../../../../shared/components/cron-schedule';
import { OriginCron } from '../../../../../shared/services';
import { finalize } from 'rxjs/operators';
import { GcService } from '../../../../../../../ng-swagger-gen/services/gc.service';
import { GCHistory } from '../../../../../../../ng-swagger-gen/models/gchistory';
import { ScheduleType } from '../../../../../shared/entities/shared.const';
import { GcHistoryComponent } from './gc-history/gc-history.component';
import {
    JOB_STATUS,
    REFRESH_STATUS_TIME_DIFFERENCE,
    WORKER_OPTIONS,
} from '../../clearing-job-interfact';
import { clone } from '../../../../../shared/units/utils';

const ONE_MINUTE = 60000;

@Component({
    selector: 'gc-config',
    templateUrl: './gc.component.html',
    styleUrls: ['./gc.component.scss'],
})
export class GcComponent implements OnInit, OnDestroy {
    originCron: OriginCron;
    disableGC: boolean = false;
    getLabelCurrent = 'GC.CURRENT_SCHEDULE';
    loadingGcStatus = false;
    @ViewChild(CronScheduleComponent)
    cronScheduleComponent: CronScheduleComponent;
    shouldDeleteUntagged: boolean;
    workerNum: number = 1;
    workerOptions: number[] = clone(WORKER_OPTIONS);
    dryRunOnGoing: boolean = false;

    lastCompletedTime: string;
    loadingLastCompletedTime: boolean = false;
    isDryRun: boolean = false;
    nextScheduledTime: string;
    statusTimeout: any;
    @ViewChild(GcHistoryComponent) gcHistoryComponent: GcHistoryComponent;
    constructor(
        private gcService: GcService,
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
        this.gcService
            .getGCHistory({
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
        this.gcService
            .getGCSchedule()
            .pipe(
                finalize(() => {
                    this.loadingGcStatus = false;
                })
            )
            .subscribe(
                schedule => {
                    this.initSchedule(schedule);
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }

    private initSchedule(gcHistory: GCHistory) {
        this.nextScheduledTime = gcHistory?.schedule?.next_scheduled_time
            ? gcHistory?.schedule?.next_scheduled_time
            : null;
        if (gcHistory && gcHistory.schedule) {
            this.originCron = {
                type: gcHistory.schedule.type,
                cron: gcHistory.schedule.cron,
            };
        } else {
            this.originCron = {
                type: ScheduleType.NONE,
                cron: '',
            };
        }
        if (gcHistory && gcHistory.job_parameters) {
            this.shouldDeleteUntagged = JSON.parse(
                gcHistory.job_parameters
            ).delete_untagged;
            this.workerNum = +JSON.parse(gcHistory.job_parameters).workers;
        } else {
            this.shouldDeleteUntagged = false;
            this.workerNum = 1;
        }
    }

    gcNow(): void {
        this.disableGC = true;
        setTimeout(() => {
            this.enableGc();
        }, ONE_MINUTE);

        this.gcService
            .createGCSchedule({
                schedule: {
                    parameters: {
                        delete_untagged: this.shouldDeleteUntagged,
                        workers: +this.workerNum,
                        dry_run: false,
                    },
                    schedule: {
                        type: ScheduleType.MANUAL,
                    },
                },
            })
            .subscribe({
                next: response => {
                    this.errorHandler.info('GC.MSG_SUCCESS');
                    this.refresh();
                },
                error: error => {
                    this.errorHandler.error(error);
                },
            });
    }

    dryRun() {
        this.dryRunOnGoing = true;
        this.gcService
            .createGCSchedule({
                schedule: {
                    parameters: {
                        delete_untagged: this.shouldDeleteUntagged,
                        workers: +this.workerNum,
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
        if (this.originCron && this.originCron.type === ScheduleType.NONE) {
            // no schedule, then create
            this.gcService
                .createGCSchedule({
                    schedule: {
                        parameters: {
                            delete_untagged: this.shouldDeleteUntagged,
                            workers: +this.workerNum,
                            dry_run: false,
                        },
                        schedule: {
                            type: GcComponent.getScheduleType(cron),
                            cron: cron,
                        },
                    },
                })
                .subscribe(
                    response => {
                        this.errorHandler.info('GC.MSG_SCHEDULE_RESET');
                        this.cronScheduleComponent.resetSchedule();
                        this.getCurrentSchedule(false); // refresh schedule
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        } else {
            this.gcService
                .updateGCSchedule({
                    schedule: {
                        parameters: {
                            delete_untagged: this.shouldDeleteUntagged,
                            workers: +this.workerNum,
                            dry_run: false,
                        },
                        schedule: {
                            type: GcComponent.getScheduleType(cron),
                            cron: cron,
                        },
                    },
                })
                .subscribe(
                    response => {
                        this.errorHandler.info('GC.MSG_SCHEDULE_RESET');
                        this.cronScheduleComponent.resetSchedule();
                        this.getCurrentSchedule(false); // refresh schedule
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }

    static getScheduleType(
        cron: string
    ): 'Hourly' | 'Daily' | 'Weekly' | 'Custom' | 'Manual' | 'None' {
        if (cron) {
            if (cron === '0 0 * * * *') {
                return ScheduleType.HOURLY;
            }
            if (cron === '0 0 0 * * *') {
                return ScheduleType.DAILY;
            }
            if (cron === '0 0 0 * * 0') {
                return ScheduleType.WEEKLY;
            }
            return ScheduleType.CUSTOM;
        }
        return ScheduleType.NONE;
    }
    refresh() {
        this.getStatus();
        this.gcHistoryComponent?.refresh();
    }
}
