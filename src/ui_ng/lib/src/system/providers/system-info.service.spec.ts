import { TestBed, inject } from '@angular/core/testing';

import { SystemInfoService } from './system-info.service';
import { HttpModule } from '@angular/http';
import { SERVICE_CONFIG, IServiceConfig } from '../../service.config';

export const testConfig: IServiceConfig = {
  systemInfoEndpoint: "/api/systeminfo",
  repositoryBaseEndpoint: "",
  logBaseEndpoint: "",
  targetBaseEndpoint: "",
  replicationRuleEndpoint: "",
  replicationJobEndpoint: ""
};

describe('SysteninfoService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpModule],
      providers: [
        { provide: SERVICE_CONFIG, useValue: testConfig },
        SystemInfoService]
    });
  });

  it('should be initialized', inject([SystemInfoService], (service: SystemInfoService) => {
    expect(service).toBeTruthy();
  }));
});
