import {
    Component,
    Output,
    EventEmitter,
    ViewChild,
    OnInit,
} from '@angular/core';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { CronScheduleComponent } from '../../../../shared/components/cron-schedule';
import { OriginCron } from '../../../../shared/services';
import { finalize } from 'rxjs/operators';
import { GcService } from '../../../../../../ng-swagger-gen/services/gc.service';
import { GCHistory } from '../../../../../../ng-swagger-gen/models/gchistory';
import { ScheduleType } from '../../../../shared/entities/shared.const';

const ONE_MINUTE = 60000;

@Component({
    selector: 'gc-config',
    templateUrl: './gc.component.html',
    styleUrls: ['./gc.component.scss'],
})
export class GcComponent implements OnInit {
    originCron: OriginCron;
    disableGC: boolean = false;
    getLabelCurrent = 'GC.CURRENT_SCHEDULE';
    @Output() loadingGcStatus = new EventEmitter<boolean>();
    @ViewChild(CronScheduleComponent)
    CronScheduleComponent: CronScheduleComponent;
    shouldDeleteUntagged: boolean;
    dryRunOnGoing: boolean = false;

    constructor(
        private gcService: GcService,
        private errorHandler: ErrorHandler
    ) {}

    ngOnInit() {
        this.getCurrentSchedule();
    }

    getCurrentSchedule() {
        this.loadingGcStatus.emit(true);
        this.gcService
            .getGCSchedule()
            .pipe(
                finalize(() => {
                    this.loadingGcStatus.emit(false);
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
        } else {
            this.shouldDeleteUntagged = false;
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
                        dry_run: false,
                    },
                    schedule: {
                        type: ScheduleType.MANUAL,
                    },
                },
            })
            .subscribe(
                response => {
                    this.errorHandler.info('GC.MSG_SUCCESS');
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }

    dryRun() {
        this.dryRunOnGoing = true;
        this.gcService
            .createGCSchedule({
                schedule: {
                    parameters: {
                        delete_untagged: this.shouldDeleteUntagged,
                        dry_run: true,
                    },
                    schedule: {
                        type: ScheduleType.MANUAL,
                    },
                },
            })
            .pipe(finalize(() => (this.dryRunOnGoing = false)))
            .subscribe(
                response => {
                    this.errorHandler.info('GC.DRY_RUN_SUCCESS');
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }

    private enableGc() {
        this.disableGC = false;
    }

    saveGcSchedule(cron: string) {
        if (this.originCron && this.originCron.type !== ScheduleType.NONE) {
            // no schedule, then create
            this.gcService
                .createGCSchedule({
                    schedule: {
                        parameters: {
                            delete_untagged: this.shouldDeleteUntagged,
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
                        this.CronScheduleComponent.resetSchedule();
                        this.getCurrentSchedule(); // refresh schedule
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
                        this.CronScheduleComponent.resetSchedule();
                        this.getCurrentSchedule(); // refresh schedule
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
}
