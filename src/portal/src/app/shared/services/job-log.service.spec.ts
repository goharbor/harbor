import { TestBed, inject } from '@angular/core/testing';
import { JobLogService, JobLogDefaultService } from './job-log.service';
import { SharedTestingModule } from '../shared.module';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('JobLogService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule, HttpClientTestingModule],
            providers: [
                JobLogDefaultService,
                {
                    provide: JobLogService,
                    useClass: JobLogDefaultService,
                },
            ],
        });
    });

    it('should be initialized', inject(
        [JobLogDefaultService],
        (service: JobLogService) => {
            expect(service).toBeTruthy();
        }
    ));
});
