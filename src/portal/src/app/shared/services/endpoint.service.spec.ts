import { TestBed, inject } from '@angular/core/testing';
import { SharedTestingModule } from '../shared.module';
import { EndpointDefaultService, EndpointService } from './endpoint.service';

describe('EndpointService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            providers: [
                EndpointDefaultService,
                {
                    provide: EndpointService,
                    useClass: EndpointDefaultService,
                },
            ],
        });
    });

    it('should be initialized', inject(
        [EndpointDefaultService],
        (service: EndpointService) => {
            expect(service).toBeTruthy();
        }
    ));
});
