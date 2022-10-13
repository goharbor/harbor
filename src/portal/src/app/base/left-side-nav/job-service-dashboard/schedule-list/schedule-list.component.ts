import { Component, OnDestroy, OnInit } from '@angular/core';
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
import {
    EventService,
    HarborEvent,
} from '../../../../services/event-service/event.service';
import { Subscription } from 'rxjs';
import { ScheduleService } from '../../../../../../ng-swagger-gen/services/schedule.service';

@Component({
    selector: 'app-schedule-list',
    templateUrl: './schedule-list.component.html',
    styleUrls: ['./schedule-list.component.scss'],
})
export class ScheduleListComponent implements OnInit, OnDestroy {
    loadingSchedules: boolean = false;
    schedules: ScheduleTask[] = [];
    total: number = 0;
    page: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.SCHEDULE_LIST_COMPONENT
    );
    eventSub: Subscription;
    constructor(
        private messageHandlerService: MessageHandlerService,
        private eventService: EventService,
        private scheduleService: ScheduleService
    ) {}

    ngOnInit() {
        this.initEventSub();
    }

    ngOnDestroy() {
        if (this.eventSub) {
            this.eventSub.unsubscribe();
            this.eventSub = null;
        }
    }

    initEventSub() {
        if (!this.eventSub) {
            this.eventSub = this.eventService.subscribe(
                HarborEvent.REFRESH_JOB_SERVICE_DASHBOARD,
                () => {
                    this.clrLoad();
                }
            );
        }
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
        this.scheduleService
            .listSchedulesResponse({
                page: this.page,
                pageSize: this.pageSize,
            })
            .pipe(finalize(() => (this.loadingSchedules = false)))
            .subscribe({
                next: res => {
                    // Get total count
                    if (res.headers) {
                        let xHeader: string = res.headers.get('x-total-count');
                        if (xHeader) {
                            this.total = Number.parseInt(xHeader, 10);
                        }
                    }
                    this.schedules = doSorting(res.body, state);
                },
                error: err => {
                    this.messageHandlerService.error(err);
                },
            });
    }

    json(v: string): object {
        if (v) {
            return JSON.parse(v);
        }
        return null;
    }
}
