import { TestBed, inject } from '@angular/core/testing';

import { JobLogService, JobLogDefaultService } from './job-log.service';
import { SharedModule } from '../utils/shared/shared.module';
import { SERVICE_CONFIG, IServiceConfig } from '../entities/service.config';
import { CURRENT_BASE_HREF } from "../utils/utils";

describe('JobLogService', () => {
  const mockConfig: IServiceConfig = {
    replicationBaseEndpoint: CURRENT_BASE_HREF + "/replication/testing",
    scanJobEndpoint: CURRENT_BASE_HREF + "/jobs/scan/testing"
  };

  let config: IServiceConfig;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        JobLogDefaultService,
        {
          provide: JobLogService,
          useClass: JobLogDefaultService
        }, {
          provide: SERVICE_CONFIG,
          useValue: mockConfig
        }]
    });

    config = TestBed.get(SERVICE_CONFIG);
  });

  it('should be initialized', inject([JobLogDefaultService], (service: JobLogService) => {
    expect(service).toBeTruthy();
    expect(config.replicationBaseEndpoint).toEqual(CURRENT_BASE_HREF + "/replication/testing");
    expect(config.scanJobEndpoint).toEqual(CURRENT_BASE_HREF + "/jobs/scan/testing");
  }));
});
