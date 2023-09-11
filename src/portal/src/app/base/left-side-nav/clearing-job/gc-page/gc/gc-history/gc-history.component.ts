import { Component, OnDestroy, OnInit } from '@angular/core';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { forkJoin, Subscription, timer } from 'rxjs';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../../../../shared/entities/shared.const';
import { GcService } from '../../../../../../../../ng-swagger-gen/services/gc.service';
import {
    CURRENT_BASE_HREF,
    formatSize,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../../../shared/units/utils';
import { ClrDatagridStateInterface } from '@clr/angular';
import { finalize } from 'rxjs/operators';
import { GCHistory } from '../../../../../../../../ng-swagger-gen/models/gchistory';
import {
    JOB_STATUS,
    NO,
    REFRESH_STATUS_TIME_DIFFERENCE,
    YES,
} from '../../../clearing-job-interfact';
import { ConfirmationMessage } from '../../../../../global-confirmation-dialog/confirmation-message';
import { ConfirmationDialogService } from '../../../../../global-confirmation-dialog/confirmation-dialog.service';

@Component({
    selector: 'gc-history',
    templateUrl: './gc-history.component.html',
    styleUrls: ['./gc-history.component.scss'],
})
export class GcHistoryComponent implements OnInit, OnDestroy {
    jobs: Array<GCHistory> = [];
    loading: boolean = true;
    timerDelay: Subscription;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.GC_HISTORY_COMPONENT,
        5
    );
    page: number = 1;
    total: number = 0;
    state: ClrDatagridStateInterface;
    selectedRow: GCHistory[] = [];
    isStopOnGoing: boolean = false;
    subscription: Subscription;
    protected readonly NO = NO;
    protected readonly YES = YES;
    constructor(
        private gcService: GcService,
        private errorHandler: ErrorHandler,
        private confirmationDialogService: ConfirmationDialogService
    ) {}
    ngOnInit() {
        if (!this.subscription) {
            this.subscription =
                this.confirmationDialogService.confirmationConfirm$.subscribe(
                    message => {
                        if (
                            message &&
                            message.state === ConfirmationState.CONFIRMED &&
                            message.source === ConfirmationTargets.STOP_GC
                        ) {
                            this.stopGc(message.data);
                        }
                    }
                );
        }
    }

    ngOnDestroy() {
        if (this.timerDelay) {
            this.timerDelay.unsubscribe();
            this.timerDelay = null;
        }
        if (this.subscription) {
            this.subscription.unsubscribe();
            this.subscription = null;
        }
    }

    stopGc(gCHistory: GCHistory[]) {
        this.isStopOnGoing = true;
        forkJoin(
            gCHistory.map(item => {
                return this.gcService.stopGC({
                    gcId: item.id,
                });
            })
        )
            .pipe(finalize(() => (this.isStopOnGoing = false)))
            .subscribe({
                next: res => {
                    this.errorHandler.info('CLEARANCES.STOP_GC_SUCCESS');
                },
                error: err => {
                    this.errorHandler.error(err);
                },
            });
    }

    refresh() {
        this.page = 1;
        this.total = 0;
        this.getJobs(true);
    }

    getJobs(withLoading: boolean, state?: ClrDatagridStateInterface) {
        if (state) {
            this.state = state;
        }
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.GC_HISTORY_COMPONENT,
                this.pageSize
            );
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
        }
        if (withLoading) {
            this.loading = true;
        }
        this.gcService
            .getGCHistoryResponse({
                page: this.page,
                pageSize: this.pageSize,
                q: q,
                sort: sort,
            })
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                res => {
                    // Get total count
                    if (res.headers) {
                        const xHeader: string =
                            res.headers.get('X-Total-Count');
                        if (xHeader) {
                            this.total = parseInt(xHeader, 0);
                        }
                        if (!withLoading) {
                            if (res?.body?.length) {
                                res.body.forEach(item => {
                                    this.jobs.forEach(item2 => {
                                        if (item2.id === item.id) {
                                            item2.job_status = item.job_status;
                                            item2.update_time =
                                                item.update_time;
                                            item2.job_parameters =
                                                item.job_parameters;
                                        }
                                    });
                                });
                            }
                        } else {
                            this.selectedRow = [];
                            this.jobs = res.body;
                        }
                    }
                    // to avoid some jobs not finished.
                    if (!this.timerDelay) {
                        this.timerDelay = timer(
                            REFRESH_STATUS_TIME_DIFFERENCE,
                            REFRESH_STATUS_TIME_DIFFERENCE
                        ).subscribe(() => {
                            let count: number = 0;
                            this.jobs.forEach(job => {
                                if (
                                    job.job_status === JOB_STATUS.PENDING ||
                                    job.job_status === JOB_STATUS.RUNNING
                                ) {
                                    count++;
                                }
                            });
                            if (count > 0) {
                                this.getJobs(false, this.state);
                            } else {
                                this.timerDelay.unsubscribe();
                                this.timerDelay = null;
                            }
                        });
                    }
                },
                error => {
                    this.errorHandler.error(error);
                    this.loading = false;
                }
            );
    }

    isDryRun(param: string): string {
        if (param) {
            const paramObj: any = JSON.parse(param);
            if (paramObj && paramObj.dry_run) {
                return YES;
            }
        }
        return NO;
    }
    getBlobs(param: string): number {
        if (param) {
            const paramObj: any = JSON.parse(param);
            if (paramObj && paramObj.purged_blobs) {
                return paramObj.purged_blobs;
            }
        }
        return 0;
    }

    getManifest(param: string): number {
        if (param) {
            const paramObj: any = JSON.parse(param);
            if (paramObj && paramObj.purged_manifests) {
                return paramObj.purged_manifests;
            }
        }
        return 0;
    }

    getSize(param: string): string {
        if (param) {
            const paramObj: any = JSON.parse(param);
            if (paramObj && paramObj.freed_space) {
                return formatSize(paramObj.freed_space);
            }
        }
        return '0';
    }

    getLogLink(id): string {
        return `${CURRENT_BASE_HREF}/system/gc/${id}/log`;
    }
    canStop(): boolean {
        if (this.isStopOnGoing) {
            return false;
        }
        if (this.selectedRow?.length) {
            return (
                this.selectedRow.filter(item => {
                    return (
                        item.job_status === JOB_STATUS.PENDING ||
                        item.job_status === JOB_STATUS.RUNNING
                    );
                })?.length > 0
            );
        }
        return false;
    }

    openStopExecutionsDialog() {
        const executionIds = this.selectedRow.map(robot => robot.id).join(',');
        let StopExecutionsMessage = new ConfirmationMessage(
            'REPLICATION.STOP_TITLE',
            'REPLICATION.STOP_SUMMARY',
            executionIds,
            this.selectedRow,
            ConfirmationTargets.STOP_GC,
            ConfirmationButtons.CONFIRM_CANCEL
        );
        this.confirmationDialogService.openComfirmDialog(StopExecutionsMessage);
    }

    shouldShowDetails(status: string): boolean {
        return status !== JOB_STATUS.PENDING && status !== JOB_STATUS.RUNNING;
    }
}
