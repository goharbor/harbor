import { TestBed, inject } from '@angular/core/testing';

import { AccessLogService, AccessLogDefaultService } from './access-log.service';

describe('AccessLogService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [{
        provide: AccessLogService,
        useClass: AccessLogDefaultService
      }]
    });
  });

  it('should ...', inject([AccessLogDefaultService], (service: AccessLogService) => {
    expect(service).toBeTruthy();
  }));
});
