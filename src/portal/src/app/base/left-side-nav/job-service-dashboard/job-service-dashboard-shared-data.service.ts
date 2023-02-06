import { Injectable } from '@angular/core';
import { JobQueue } from '../../../../../ng-swagger-gen/models/job-queue';
import { JobserviceService } from '../../../../../ng-swagger-gen/services/jobservice.service';
import { map, Observable } from 'rxjs';
import { All, ScheduleListResponse } from './job-service-dashboard.interface';
import { Worker } from 'ng-swagger-gen/models';
import { ScheduleService } from '../../../../../ng-swagger-gen/services/schedule.service';
import { ClrDatagridStateInterface } from '@clr/angular/data/datagrid/interfaces/state.interface';
import { doSorting } from '../../../shared/units/utils';

@Injectable()
export class JobServiceDashboardSharedDataService {
    private _jobQueues: JobQueue[] = [];
    private _allWorkers: Worker[] = [];
    private _scheduleListResponse: ScheduleListResponse;

    private _scheduleListParam: ScheduleService.ListSchedulesParams = {
        page: 1,
        pageSize: 1,
    };

    private _state: ClrDatagridStateInterface;
    constructor(
        private jobServiceService: JobserviceService,
        private scheduleService: ScheduleService
    ) {}
    getJobQueues(): JobQueue[] {
        return this._jobQueues;
    }
    getAllWorkers(): Worker[] {
        return this._allWorkers;
    }
    getScheduleListResponse(): ScheduleListResponse {
        return this._scheduleListResponse;
    }

    retrieveJobQueues(isAutoRefresh?: boolean): Observable<JobQueue[]> {
        return this.jobServiceService.listJobQueues().pipe(
            map(res => {
                // For auto-refresh
                // if execute this._jobQueues = res, the selected rows of the datagrid will be reset
                // so only refresh properties here
                if (isAutoRefresh && this._jobQueues?.length && res?.length) {
                    this._jobQueues.forEach(item => {
                        res.forEach(item2 => {
                            if (item2.job_type === item.job_type) {
                                item.count = item2.count;
                                item.latency = item2.latency;
                                item.paused = item2.paused;
                            }
                        });
                    });
                } else {
                    this._jobQueues = res;
                }
                // map null and undefined to 0
                this._jobQueues.forEach(item => {
                    if (!item.count) {
                        item.count = 0;
                    }
                    if (!item.latency) {
                        item.latency = 0;
                    }
                });
                return this._jobQueues;
            })
        );
    }
    retrieveScheduleListResponse(
        params?: ScheduleService.ListSchedulesParams,
        state?: ClrDatagridStateInterface
    ): Observable<ScheduleListResponse> {
        if (params) {
            this._scheduleListParam = params;
        }
        if (state) {
            this._state = state;
        }
        return this.scheduleService
            .listSchedulesResponse(this._scheduleListParam)
            .pipe(
                map(res => {
                    const result: ScheduleListResponse = {
                        pageSize: this._scheduleListParam?.pageSize | 1,
                        currentPage: this._scheduleListParam?.page | 1,
                        scheduleList: res.body,
                        total: Number.parseInt(
                            res.headers.get('x-total-count'),
                            10
                        ),
                    };
                    this._scheduleListResponse = result;
                    if (
                        this._state &&
                        this._scheduleListResponse?.scheduleList?.length
                    ) {
                        this._scheduleListResponse.scheduleList = doSorting(
                            this._scheduleListResponse?.scheduleList,
                            this._state
                        );
                    }
                    return this._scheduleListResponse;
                })
            );
    }
    retrieveAllWorkers(isAutoRefresh?: boolean): Observable<Worker[]> {
        return this.jobServiceService
            .getWorkers({
                poolId: All.ALL_WORKERS.toString(),
            })
            .pipe(
                map(res => {
                    // For auto-refresh
                    // if execute this._allWorkers = res, the selected rows of the datagrid will be reset
                    // so only refresh properties here
                    if (
                        isAutoRefresh &&
                        this._allWorkers?.length &&
                        res?.length
                    ) {
                        this._allWorkers.forEach(item => {
                            res.forEach(item2 => {
                                if (item2.id === item.id) {
                                    item.job_id = item2.job_id;
                                    item.job_name = item2.job_name;
                                    item.check_in = item2.check_in;
                                    item.checkin_at = item2.checkin_at;
                                    item.start_at = item2.start_at;
                                }
                            });
                        });
                    } else {
                        this._allWorkers = res;
                    }
                    return this._allWorkers;
                })
            );
    }
}
