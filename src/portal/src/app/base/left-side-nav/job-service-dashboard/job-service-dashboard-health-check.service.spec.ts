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
import { JobServiceDashboardHealthCheckService } from './job-service-dashboard-health-check.service';
import { JobserviceService } from '../../../../../ng-swagger-gen/services/jobservice.service';
import { of } from 'rxjs';

describe('JobServiceDashboardHealthCheckService', () => {
    let service: JobServiceDashboardHealthCheckService;

    const fakedJobserviceService = {
        listJobQueues() {
            return of({});
        },
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [
                {
                    provide: JobserviceService,
                    useValue: fakedJobserviceService,
                },
            ],
        });
        service = TestBed.inject(JobServiceDashboardHealthCheckService);
    });

    it('should be created', () => {
        expect(service).toBeTruthy();
    });

    it('should return false when hasUnhealthyQueue is called', () => {
        expect(service.hasUnhealthyQueue()).toBeFalsy();
    });

    it('should return false when hasManuallyClosed is called', () => {
        expect(service.hasManuallyClosed()).toBeFalsy();
    });

    it('should return false when hasUnhealthyQueue is called', () => {
        expect(service.hasUnhealthyQueue()).toBeFalsy();
    });
});
