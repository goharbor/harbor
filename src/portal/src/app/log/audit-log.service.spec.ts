import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { AuditLogService } from './audit-log.service';

describe('AuditLogService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule
      ],
      providers: [AuditLogService]
    });
  });

  it('should be created', inject([AuditLogService], (service: AuditLogService) => {
    expect(service).toBeTruthy();
  }));
});
