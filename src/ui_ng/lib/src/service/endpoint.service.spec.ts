import { TestBed, inject } from '@angular/core/testing';

import { EndpointService, EndpointDefaultService } from './endpoint.service';

describe('EndpointService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [{
        provide: EndpointService,
        useClass: EndpointDefaultService
      }]
    });
  });

  it('should ...', inject([EndpointDefaultService], (service: EndpointService) => {
    expect(service).toBeTruthy();
  }));
});
