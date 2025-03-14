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
import { PendingCardComponent } from './pending-job-card.component';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { JobQueue } from '../../../../../../ng-swagger-gen/models/job-queue';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';

describe('PendingCardComponent', () => {
    let component: PendingCardComponent;
    let fixture: ComponentFixture<PendingCardComponent>;

    const mockedJobs: JobQueue[] = [
        {
            job_type: 'test1',
            count: 1,
        },
        {
            job_type: 'test2',
            count: 2,
        },
        {
            job_type: 'test3',
            count: 3,
        },
    ];

    const fakedJobServiceDashboardSharedDataService = {
        _jobQueues: [],
        getJobQueues(): JobQueue[] {
            return this._jobQueues;
        },
        retrieveJobQueues() {
            return of(mockedJobs).pipe(delay(0));
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [PendingCardComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: JobServiceDashboardSharedDataService,
                    useValue: fakedJobServiceDashboardSharedDataService,
                },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(PendingCardComponent);
        component = fixture.componentInstance;
        spyOn(component, 'loopGetPendingJobs').and.callFake(() => {
            fakedJobServiceDashboardSharedDataService
                .retrieveJobQueues()
                .subscribe(res => {
                    component.loading = false;
                    fakedJobServiceDashboardSharedDataService._jobQueues =
                        res.sort((a, b) => {
                            const ACount: number = a?.count | 0;
                            const BCount: number = b?.count | 0;
                            return BCount - ACount;
                        });
                });
        });
        fixture.autoDetectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render data', async () => {
        await fixture.whenStable();
        expect(component.total()).toEqual(6);
    });
});
