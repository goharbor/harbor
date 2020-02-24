import { TestBed, inject } from '@angular/core/testing';

import { RepositoryService, RepositoryDefaultService } from './repository.service';
import { IServiceConfig, SERVICE_CONFIG } from "../../../lib/entities/service.config";
import { SharedModule } from "../../../lib/utils/shared/shared.module";
import { CURRENT_BASE_HREF } from "../../../lib/utils/utils";


describe('RepositoryService', () => {
  const mockConfig: IServiceConfig = {
    repositoryBaseEndpoint: CURRENT_BASE_HREF + "/repositories/testing"
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
    expect(config.repositoryBaseEndpoint).toEqual(CURRENT_BASE_HREF + "/repositories/testing");
  });

});
