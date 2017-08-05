import { TestBed, inject } from '@angular/core/testing';
import { SharedModule } from '../shared/shared.module';
import { EndpointService, EndpointDefaultService } from './endpoint.service';
import { IServiceConfig, SERVICE_CONFIG } from '../service.config';


describe('EndpointService', () => {

  let mockEndpoint:IServiceConfig = {
    targetBaseEndpoint: '/api/endpoint/testing'
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        EndpointDefaultService,
        {
          provide: SERVICE_CONFIG,
          useValue: mockEndpoint
        },
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
