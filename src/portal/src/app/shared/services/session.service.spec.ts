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
import { TestBed, inject } from '@angular/core/testing';
import {
    HttpTestingController,
    provideHttpClientTesting,
} from '@angular/common/http/testing';
import { SessionService } from './session.service';
import { JobServiceDashboardHealthCheckService } from '../../base/left-side-nav/job-service-dashboard/job-service-dashboard-health-check.service';
import { CURRENT_BASE_HREF } from '../units/utils';
import {
    provideHttpClient,
    withInterceptorsFromDi,
} from '@angular/common/http';

describe('SessionService', () => {
    const healthCheck = jasmine.createSpyObj('health', ['checkHealth']);

    beforeEach(() => {
        healthCheck.checkHealth.calls.reset();
        TestBed.configureTestingModule({
            imports: [],
            providers: [
                SessionService,
                provideHttpClient(withInterceptorsFromDi()),
                provideHttpClientTesting(),
                {
                    provide: JobServiceDashboardHealthCheckService,
                    useValue: healthCheck,
                },
            ],
        });
    });

    function retrieveUser(sysadmin: boolean) {
        TestBed.inject(SessionService).retrieveUser().subscribe();
        TestBed.inject(HttpTestingController)
            .expectOne(CURRENT_BASE_HREF + '/users/current')
            .flush({ user_id: 2, username: 'user', sysadmin_flag: sysadmin });
    }

    it('should check the job service queue health for a system admin', () => {
        retrieveUser(true);
        expect(healthCheck.checkHealth).toHaveBeenCalled();
    });

    it('should not check the job service queue health for other users', () => {
        retrieveUser(false);
        expect(healthCheck.checkHealth).not.toHaveBeenCalled();
    });

    it('should be created', inject(
        [SessionService],
        (service: SessionService) => {
            expect(service).toBeTruthy();
        }
    ));
});
