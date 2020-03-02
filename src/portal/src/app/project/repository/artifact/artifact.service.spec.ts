import { TestBed, inject } from '@angular/core/testing';
import { IServiceConfig, SERVICE_CONFIG } from "../../../../lib/entities/service.config";
import { SharedModule } from "../../../../lib/utils/shared/shared.module";
import { TagDefaultService, TagService } from "../../../../lib/services";
import { CURRENT_BASE_HREF } from "../../../../lib/utils/utils";

describe('TagService', () => {

  const mockConfig: IServiceConfig = {
    repositoryBaseEndpoint: CURRENT_BASE_HREF + "/repositories/testing"
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
