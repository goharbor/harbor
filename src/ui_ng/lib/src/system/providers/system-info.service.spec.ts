import { TestBed, inject } from '@angular/core/testing';

import { SystemInfoService } from './system-info.service';

describe('SysteninfoService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [SystemInfoService]
    });
  });

  it('should ...', inject([SystemInfoService], (service: SystemInfoService) => {
    expect(service).toBeTruthy();
  }));
});
