import { TestBed, inject } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { SystemAdminGuard } from './system-admin-activate.service';
import { AppConfigService } from '../../app-config.service';
import { SessionService } from '../../shared/session.service';

describe('SystemAdminGuard', () => {
  const fakeAppConfigService = null;
  const fakeSessionService = null;
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        RouterTestingModule
      ],
      providers: [
        SystemAdminGuard,
        { provide: AppConfigService, useValue: fakeAppConfigService },
        { provide: SessionService, useValue: fakeSessionService }
      ]
    });
  });

  it('should be created', inject([SystemAdminGuard], (service: SystemAdminGuard) => {
    expect(service).toBeTruthy();
  }));
});
