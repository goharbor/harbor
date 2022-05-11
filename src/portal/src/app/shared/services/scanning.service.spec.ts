import { TestBed, inject } from '@angular/core/testing';
import {
    ScanningResultService,
    ScanningResultDefaultService,
} from './scanning.service';
import { SharedTestingModule } from '../shared.module';
import { HttpClientTestingModule } from '@angular/common/http/testing';

describe('ScanningResultService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule, HttpClientTestingModule],
            providers: [
                ScanningResultDefaultService,
                {
                    provide: ScanningResultService,
                    useClass: ScanningResultDefaultService,
                },
            ],
        });
    });

    it('should be initialized', inject(
        [ScanningResultDefaultService],
        (service: ScanningResultService) => {
            expect(service).toBeTruthy();
        }
    ));
});
