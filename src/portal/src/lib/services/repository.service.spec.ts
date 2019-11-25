import { TestBed, inject } from '@angular/core/testing';

import { RepositoryService, RepositoryDefaultService } from './repository.service';
import { SharedModule } from '../utils/shared/shared.module';
import { SERVICE_CONFIG, IServiceConfig } from '../entities/service.config';

describe('RepositoryService', () => {
  const mockConfig: IServiceConfig = {
    repositoryBaseEndpoint: "/api/repositories/testing"
  };

  let config: IServiceConfig;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        RepositoryDefaultService,
        {
          provide: RepositoryService,
          useClass: RepositoryDefaultService
        }, {
          provide: SERVICE_CONFIG,
          useValue: mockConfig
        }]
    });

    config = TestBed.get(SERVICE_CONFIG);
  });

  it('should be initialized', inject([RepositoryDefaultService], (service: RepositoryService) => {
    expect(service).toBeTruthy();
  }));

  it('should inject the right config', () => {
    expect(config).toBeTruthy();
    expect(config.repositoryBaseEndpoint).toEqual("/api/repositories/testing");
  });

});
