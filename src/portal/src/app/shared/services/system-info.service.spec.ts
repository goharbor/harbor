import { TestBed, inject } from '@angular/core/testing';
import {
    SystemInfoService,
    SystemInfoDefaultService,
} from './system-info.service';
import { SharedTestingModule } from '../shared.module';

describe('SystemInfoService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            providers: [
                SystemInfoDefaultService,
                {
                    provide: SystemInfoService,
                    useClass: SystemInfoDefaultService,
                },
            ],
        });
    });

    it('should be initialized', inject(
        [SystemInfoDefaultService],
        (service: SystemInfoService) => {
            expect(service).toBeTruthy();
        }
    ));
});
