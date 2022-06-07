import { Component, OnDestroy } from '@angular/core';
import { ClrDatagridStateInterface } from '@clr/angular';
import { GCHistory } from 'ng-swagger-gen/models/gchistory';
import { finalize, Subscription, timer } from 'rxjs';
import { REFRESH_TIME_DIFFERENCE } from 'src/app/shared/entities/shared.const';
import { ErrorHandler } from 'src/app/shared/units/error-handler/error-handler';
import {
    CURRENT_BASE_HREF,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from 'src/app/shared/units/utils';
import { PurgeService } from '../../../../../../../ng-swagger-gen/services/purge.service';
import { JOB_STATUS, NO, YES } from '../../clearing-job-interfact';

@Component({
    selector: 'app-purge-history',
    templateUrl: './purge-history.component.html',
    styleUrls: ['./purge-history.component.scss'],
})
export class PurgeHistoryComponent implements OnDestroy {
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
    constructor(
        private purgeService: PurgeService,
        private errorHandler: ErrorHandler
    ) {}
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
        this.purgeService
            .getPurgeHistoryResponse({
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
                        this.jobs = res.body;
                    }
                    // to avoid some jobs not finished.
                    if (!this.timerDelay) {
                        this.timerDelay = timer(
                            REFRESH_TIME_DIFFERENCE,
                            REFRESH_TIME_DIFFERENCE
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

    ngOnDestroy() {
        if (this.timerDelay) {
            this.timerDelay.unsubscribe();
            this.timerDelay = null;
        }
    }

    getLogLink(id): string {
        return `${CURRENT_BASE_HREF}/system/purgeaudit/${id}/log`;
    }
}
