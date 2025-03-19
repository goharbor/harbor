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
import { ScheduleListComponent } from './schedule-list.component';
import { ScheduleTask } from '../../../../../../ng-swagger-gen/models/schedule-task';
import { of } from 'rxjs';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';
import { ScheduleListResponse } from '../job-service-dashboard.interface';
import { delay } from 'rxjs/operators';

describe('ScheduleListComponent', () => {
    let component: ScheduleListComponent;
    let fixture: ComponentFixture<ScheduleListComponent>;

    const mockedSchedules: ScheduleTask[] = [
        { id: 1, vendor_type: 'test1' },
        { id: 2, vendor_type: 'test2' },
    ];

    const fakedJobServiceDashboardSharedDataService = {
        _scheduleListResponse: {
            scheduleList: mockedSchedules,
            total: mockedSchedules.length,
        },
        getScheduleListResponse(): ScheduleListResponse {
            return this._scheduleListResponse;
        },
        retrieveScheduleListResponse() {
            return of(this._scheduleListResponse).pipe(delay(0));
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ScheduleListComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: JobServiceDashboardSharedDataService,
                    useValue: fakedJobServiceDashboardSharedDataService,
                },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(ScheduleListComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render list', async () => {
        await fixture.whenStable();
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(2);
    });

    it('should show the right total count', async () => {
        await fixture.whenStable();
        fixture.detectChanges();
        await fixture.whenStable();
        const span: HTMLSpanElement =
            fixture.nativeElement.querySelector('#total');
        expect(span.innerText).toEqual('2');
    });
});
