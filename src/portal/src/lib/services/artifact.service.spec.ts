import { TestBed, inject } from '@angular/core/testing';

import { SharedModule } from '../utils/shared/shared.module';
import { SERVICE_CONFIG, IServiceConfig } from '../entities/service.config';
import { TagService, TagDefaultService } from './tag.service';


describe('TagService', () => {
  // let mockTags: Tag[] = [
  //   {
  //     "digest": "sha256:e5c82328a509aeb7c18c1d7fb36633dc638fcf433f651bdcda59c1cc04d3ee55",
  //     "name": "1.11.5",
  //     "size": "2049",
  //     "architecture": "amd64",
  //     "os": "linux",
  //     "docker_version": "1.12.3",
  //     "author": "NGINX Docker Maintainers \"docker-maint@nginx.com\"",
  //     "created": new Date("2016-11-08T22:41:15.912313785Z"),
  //     "signature": null,
  //     'labels': []
  //   }
  // ];

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
