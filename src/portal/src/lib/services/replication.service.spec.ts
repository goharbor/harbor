import { TestBed, inject } from '@angular/core/testing';

import { ReplicationService, ReplicationDefaultService } from './replication.service';
import { SharedModule } from '../utils/shared/shared.module';
import { SERVICE_CONFIG, IServiceConfig } from '../entities/service.config';
import { CURRENT_BASE_HREF } from "../utils/utils";

describe('ReplicationService', () => {
  const mockConfig: IServiceConfig = {
    replicationRuleEndpoint: CURRENT_BASE_HREF + "/policies/replication/testing",
    replicationBaseEndpoint: CURRENT_BASE_HREF + "/replication/testing"
  };

  let config: IServiceConfig;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        ReplicationDefaultService,
        {
          provide: ReplicationService,
          useClass: ReplicationDefaultService
        }, {
          provide: SERVICE_CONFIG,
          useValue: mockConfig
        }]
    });

    config = TestBed.get(SERVICE_CONFIG);
  });

  it('should be initialized', inject([ReplicationDefaultService], (service: ReplicationService) => {
    expect(service).toBeTruthy();
  }));

  it('should inject the right config', () => {
    expect(config).toBeTruthy();
    expect(config.replicationRuleEndpoint).toEqual(CURRENT_BASE_HREF + "/policies/replication/testing");
    expect(config.replicationBaseEndpoint).toEqual(CURRENT_BASE_HREF + "/replication/testing");
  });
});
