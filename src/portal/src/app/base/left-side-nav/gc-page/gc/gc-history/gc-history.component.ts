import { Component, OnDestroy } from '@angular/core';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { Subscription, timer } from 'rxjs';
import { REFRESH_TIME_DIFFERENCE } from '../../../../../shared/entities/shared.const';
import { GcService } from '../../../../../../../ng-swagger-gen/services/gc.service';
import {
    CURRENT_BASE_HREF,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../../shared/units/utils';
import { ClrDatagridStateInterface } from '@clr/angular';
import { finalize } from 'rxjs/operators';
import { GCHistory } from '../../../../../../../ng-swagger-gen/models/gchistory';

const JOB_STATUS = {
    PENDING: 'Pending',
    RUNNING: 'Running',
};
const YES: string = 'TAG_RETENTION.YES';
const NO: string = 'TAG_RETENTION.NO';

@Component({
    selector: 'gc-history',
    templateUrl: './gc-history.component.html',
    styleUrls: ['./gc-history.component.scss'],
})
export class GcHistoryComponent implements OnDestroy {
    jobs: Array<GCHistory> = [];
    loading: boolean = true;
    timerDelay: Subscription;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.GC_HISTORY_COMPONENT
    );
    page: number = 1;
    total: number = 0;
    state: ClrDatagridStateInterface;
    constructor(
        private gcService: GcService,
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
        return `${CURRENT_BASE_HREF}/system/gc/${id}/log`;
    }
}
