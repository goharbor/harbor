import { TestBed, inject } from '@angular/core/testing';

import { ScanningResultService, ScanningResultDefaultService } from './scanning.service';
import { SharedModule } from '../utils/shared/shared.module';
import { SERVICE_CONFIG, IServiceConfig } from '../entities/service.config';

describe('ScanningResultService', () => {
  const mockConfig: IServiceConfig = {
    vulnerabilityScanningBaseEndpoint: "/api/vulnerability/testing"
  };

  let config: IServiceConfig;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        ScanningResultDefaultService,
        {
          provide: ScanningResultService,
          useClass: ScanningResultDefaultService
        }, {
          provide: SERVICE_CONFIG,
          useValue: mockConfig
        }]
    });

    config = TestBed.get(SERVICE_CONFIG);
  });

  it('should be initialized', inject([ScanningResultDefaultService], (service: ScanningResultService) => {
    expect(service).toBeTruthy();
  }));

  it('should inject the right config', () => {
    expect(config).toBeTruthy();
    expect(config.vulnerabilityScanningBaseEndpoint).toEqual("/api/vulnerability/testing");
  });
});
