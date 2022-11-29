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
