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
import { TestBed } from '@angular/core/testing';
import { JobServiceDashboardSharedDataService } from './job-service-dashboard-shared-data.service';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('JobServiceDashboardSharedDataService', () => {
    let service: JobServiceDashboardSharedDataService;

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            providers: [JobServiceDashboardSharedDataService],
        });
        service = TestBed.inject(JobServiceDashboardSharedDataService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should have initial value', () => {
        expect(service.getAllWorkers().length).toEqual(0);
        expect(service.getJobQueues().length).toEqual(0);
    });
});
