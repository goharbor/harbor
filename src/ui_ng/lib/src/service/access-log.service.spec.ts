import { TestBed, inject } from '@angular/core/testing';

import { AccessLogService, AccessLogDefaultService } from './access-log.service';

describe('AccessLogService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        AccessLogDefaultService,
        {
          provide: AccessLogService,
          useClass: AccessLogDefaultService
        }]
    });
  });

  it('should be initialized', inject([AccessLogDefaultService], (service: AccessLogService) => {
    expect(service).toBeTruthy();
  }));
});
