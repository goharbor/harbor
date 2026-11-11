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
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ScheduleCardComponent } from './schedule-card.component';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { JobserviceService } from '../../../../../../ng-swagger-gen/services/jobservice.service';
import {
    ScheduleListResponse,
    ScheduleStatusString,
} from '../job-service-dashboard.interface';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { ScheduleTask } from '../../../../../../ng-swagger-gen/models/schedule-task';
import { ScheduleService } from '../../../../../../ng-swagger-gen/services/schedule.service';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';

describe('ScheduleCardComponent', () => {
    let component: ScheduleCardComponent;
    let fixture: ComponentFixture<ScheduleCardComponent>;

    const fakedJobserviceService = {};

    const fakedScheduleService = {
        getSchedulePaused() {
            return of({}).pipe(delay(0));
        },
        listSchedulesResponse() {
            const res: HttpResponse<Array<ScheduleTask>> = new HttpResponse<
                Array<ScheduleTask>
            >({
                headers: new HttpHeaders({ 'x-total-count': '0' }),
                body: [],
            });
            return of(res).pipe(delay(0));
        },
    };

    const fakedJobServiceDashboardSharedDataService = {
        _scheduleListResponse: {},
        getScheduleListResponse(): ScheduleListResponse {
            return this._scheduleListResponse;
        },
        retrieveScheduleListResponse() {
            return of({});
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ScheduleCardComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: JobserviceService,
                    useValue: fakedJobserviceService,
                },
                {
                    provide: ScheduleService,
                    useValue: fakedScheduleService,
                },
                {
                    provide: JobServiceDashboardSharedDataService,
                    useValue: fakedJobServiceDashboardSharedDataService,
                },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(ScheduleCardComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should show right status and right total count', () => {
        component.loadingStatus = false;
        fixture.detectChanges();
        const totalDiv: HTMLDivElement =
            fixture.nativeElement.querySelector('.duration');
        expect(totalDiv.innerText).toContain('0');
        const statusSpan: HTMLSpanElement =
            fixture.nativeElement.querySelector('.status');
        expect(statusSpan.innerText).toEqual(ScheduleStatusString.RUNNING);
    });
});
