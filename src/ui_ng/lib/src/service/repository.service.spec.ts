import { TestBed, inject } from '@angular/core/testing';

import { RepositoryService, RepositoryDefaultService } from './repository.service';
import { SharedModule } from '../shared/shared.module';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';

describe('RepositoryService', () => {
  beforeEach(() => {
    const mockConfig: IServiceConfig = {
      repositoryBaseEndpoint: "/api/repositories/testing"
    };

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
  });

  it('should be initialized', inject([RepositoryDefaultService], (service: RepositoryService) => {
    expect(service).toBeTruthy();
  }));
});
