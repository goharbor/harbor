import { TestBed, inject } from '@angular/core/testing';

import { AccessLogService, AccessLogDefaultService } from './access-log.service';
import { SharedModule } from '../shared/shared.module';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';

describe('AccessLogService', () => {
  beforeEach(() => {
    const mockConfig:IServiceConfig = {
      logBaseEndpoint:"/api/logs/testing"
    };

    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        AccessLogDefaultService,
        {
          provide: AccessLogService,
          useClass: AccessLogDefaultService
        },{
          provide: SERVICE_CONFIG,
          useValue: mockConfig
        }]
    });
  });

  it('should be initialized', inject([AccessLogDefaultService], (service: AccessLogService) => {
    expect(service).toBeTruthy();
  }));
});
