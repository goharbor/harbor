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
import { SharedTestingModule } from '../../../../shared/shared.module';
import { of } from 'rxjs';
import { WorkerCardComponent } from './worker-card.component';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { Worker } from '../../../../../../ng-swagger-gen/models/worker';
import { ScheduleListResponse } from '../job-service-dashboard.interface';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';

describe('WorkerCardComponent', () => {
    let component: WorkerCardComponent;
    let fixture: ComponentFixture<WorkerCardComponent>;

    const mockedWorkers: Worker[] = [
        { id: '1', job_id: '1', job_name: 'test1' },
        { id: '2', job_id: '2', job_name: 'test2' },
    ];

    const fakedJobServiceDashboardSharedDataService = {
        _allWorkers: mockedWorkers,
        getAllWorkers(): ScheduleListResponse {
            return this._allWorkers;
        },
        retrieveAllWorkers() {
            return of(mockedWorkers);
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [NO_ERRORS_SCHEMA],
            declarations: [WorkerCardComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: JobServiceDashboardSharedDataService,
                    useValue: fakedJobServiceDashboardSharedDataService,
                },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(WorkerCardComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should init timeout', () => {
        expect(component.statusTimeout).toBeTruthy();
    });

    it('should get workers', () => {
        expect(component.busyWorkers.length).toEqual(2);
    });
});
