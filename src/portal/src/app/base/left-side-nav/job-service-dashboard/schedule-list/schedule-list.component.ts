import { Component } from '@angular/core';
import { ClrDatagridStateInterface } from '@clr/angular/data/datagrid/interfaces/state.interface';
import {
    doSorting,
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { finalize } from 'rxjs/operators';
import { ScheduleTask } from '../../../../../../ng-swagger-gen/models/schedule-task';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';

@Component({
    selector: 'app-schedule-list',
    templateUrl: './schedule-list.component.html',
    styleUrls: ['./schedule-list.component.scss'],
})
export class ScheduleListComponent {
    loadingSchedules: boolean = true;
    total: number = 0;
    page: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.SCHEDULE_LIST_COMPONENT
    );
    constructor(
        private messageHandlerService: MessageHandlerService,
        private jobServiceDashboardSharedDataService: JobServiceDashboardSharedDataService
    ) {}

    get schedules(): ScheduleTask[] {
        return (
            this.jobServiceDashboardSharedDataService.getScheduleListResponse()
                ?.scheduleList || []
        );
    }

    clrLoad(state?: ClrDatagridStateInterface): void {
        if (state?.page?.size) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.SCHEDULE_LIST_COMPONENT,
                this.pageSize
            );
        }
        this.loadingSchedules = true;
        this.jobServiceDashboardSharedDataService
            .retrieveScheduleListResponse(
                {
                    page: this.page,
                    pageSize: this.pageSize,
                },
                state
            )
            .pipe(finalize(() => (this.loadingSchedules = false)))
            .subscribe({
                next: res => {
                    this.total = res.total;
                    doSorting(res.scheduleList, state);
                },
                error: err => {
                    this.messageHandlerService.error(err);
                },
            });
    }
}
