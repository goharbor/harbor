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
