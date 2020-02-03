import { TestBed, inject } from '@angular/core/testing';

import { SystemInfoService, SystemInfoDefaultService } from './system-info.service';
import { SharedModule } from '../utils/shared/shared.module';
import { SERVICE_CONFIG, IServiceConfig } from '../entities/service.config';

describe('SystemInfoService', () => {
  const mockConfig: IServiceConfig = {
    systemInfoEndpoint: "/api/systeminfo/testing"
  };

  let config: IServiceConfig;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        SystemInfoDefaultService,
        {
          provide: SystemInfoService,
          useClass: SystemInfoDefaultService
        }, {
          provide: SERVICE_CONFIG,
          useValue: mockConfig
        }]
    });

    config = TestBed.get(SERVICE_CONFIG);
  });

  it('should be initialized', inject([SystemInfoDefaultService], (service: SystemInfoService) => {
    expect(service).toBeTruthy();
  }));

  it('should inject the right config', () => {
    expect(config).toBeTruthy();
    expect(config.systemInfoEndpoint).toEqual("/api/systeminfo/testing");
  });
});
