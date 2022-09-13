import { TestBed, inject } from '@angular/core/testing';
import { SharedTestingModule } from '../shared/shared.module';
import { HarborTranslateLoaderService } from './harbor-translate-loader.service';

describe('ConfigService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            providers: [HarborTranslateLoaderService],
        });
    });

    it('should be created', inject(
        [HarborTranslateLoaderService],
        (service: HarborTranslateLoaderService) => {
            expect(service).toBeTruthy();
        }
    ));
});
