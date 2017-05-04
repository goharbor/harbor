import { TestBed, inject } from '@angular/core/testing';

import { EndpointService, EndpointDefaultService } from './endpoint.service';

describe('EndpointService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
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
