import { TestBed, inject } from '@angular/core/testing';

import { ReplicationService, ReplicationDefaultService } from './replication.service';
import { SharedModule } from '../utils/shared/shared.module';
import { SERVICE_CONFIG, IServiceConfig } from '../entities/service.config';

describe('ReplicationService', () => {
  const mockConfig: IServiceConfig = {
    replicationRuleEndpoint: "/api/policies/replication/testing",
    replicationBaseEndpoint: "/api/replication/testing"
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
    expect(config.replicationRuleEndpoint).toEqual("/api/policies/replication/testing");
    expect(config.replicationBaseEndpoint).toEqual("/api/replication/testing");
  });
});
