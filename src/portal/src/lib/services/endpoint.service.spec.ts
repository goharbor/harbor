import { TestBed, inject } from '@angular/core/testing';
import { SharedModule } from '../utils/shared/shared.module';
import { IServiceConfig, SERVICE_CONFIG } from '../entities/service.config';
import { EndpointService, EndpointDefaultService } from './endpoint.service';
import { CURRENT_BASE_HREF } from "../utils/utils";



describe('EndpointService', () => {

  let mockEndpoint: IServiceConfig = {
    targetBaseEndpoint: CURRENT_BASE_HREF + '/endpoint/testing'
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
