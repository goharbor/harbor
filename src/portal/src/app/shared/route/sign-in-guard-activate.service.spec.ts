import { TestBed, async, inject } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { SignInGuard } from './sign-in-guard-activate.service';
import { SessionService } from '../../shared/session.service';
import { of } from 'rxjs';
import { UserPermissionService } from "../../../lib/services";

describe('SignInGuard', () => {
  const fakeUserPermissionService = {
    getPermission() {
      return of(true);
    }
  };
  const fakeSessionService = null;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        RouterTestingModule
      ],
      providers: [
        SignInGuard,
        { provide: UserPermissionService, useValue: fakeUserPermissionService },
        { provide: SessionService, useValue: fakeSessionService },
      ]
    });
  });

  it('should ...', inject([SignInGuard], (guard: SignInGuard) => {
    expect(guard).toBeTruthy();
  }));
});

