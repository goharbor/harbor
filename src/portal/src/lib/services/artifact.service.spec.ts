import { TestBed, inject } from '@angular/core/testing';

import { SharedModule } from '../utils/shared/shared.module';
import { SERVICE_CONFIG, IServiceConfig } from '../entities/service.config';
import { TagService, TagDefaultService } from './tag.service';


describe('TagService', () => {

  const mockConfig: IServiceConfig = {
    repositoryBaseEndpoint: "/api/repositories/testing"
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        TagDefaultService,
        {
          provide: TagService,
          useClass: TagDefaultService
        }, {
          provide: SERVICE_CONFIG,
          useValue: mockConfig
        }]
    });
  });

  it('should be initialized', inject([TagDefaultService], (service: TagService) => {
    expect(service).toBeTruthy();
  }));

});
