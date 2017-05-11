import { TestBed, inject } from '@angular/core/testing';
import { SharedModule } from '../shared/shared.module';
import { EndpointService, EndpointDefaultService } from './endpoint.service';

describe('EndpointService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        EndpointDefaultService,
        {
          provide: EndpointService,
          useClass: EndpointDefaultService
        }]
    });
  });

  it('should be initialized', inject([EndpointDefaultService], (service: EndpointService) => {
    expect(service).toBeTruthy();
  }));
});
