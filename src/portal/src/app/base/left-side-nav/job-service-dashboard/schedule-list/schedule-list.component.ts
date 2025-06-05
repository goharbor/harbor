// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
import { PAGE_SIZE_OPTIONS } from 'src/app/shared/entities/shared.const';

@Component({
    selector: 'app-schedule-list',
    templateUrl: './schedule-list.component.html',
    styleUrls: ['./schedule-list.component.scss'],
})
export class ScheduleListComponent {
    clrPageSizeOptions: number[] = PAGE_SIZE_OPTIONS;
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
