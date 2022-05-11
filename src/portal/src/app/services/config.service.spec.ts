import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { ConfigurationService } from './config.service';

describe('ConfigService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [ConfigurationService],
        });
    });

    it('should be created', inject(
        [ConfigurationService],
        (service: ConfigurationService) => {
            expect(service).toBeTruthy();
        }
    ));
});
