import { TestBed, inject } from '@angular/core/testing';
import {
  CanActivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  CanActivateChild,
  NavigationExtras
} from '@angular/router';
import { RouterTestingModule } from '@angular/router/testing';
import { SessionService } from '../../shared/session.service';
import { AppConfigService } from '../../app-config.service';
import { MessageHandlerService } from '../message-handler/message-handler.service';
import { SearchTriggerService } from '../../base/global-search/search-trigger.service';
import { AuthCheckGuard } from './auth-user-activate.service';

describe('AuthCheckGuard', () => {
  const fakeSessionService = null;
  const fakeAppConfigService = null;
  const fakeMessageHandlerService = null;
  const fakeSearchTriggerService = null;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        RouterTestingModule
      ],
      providers: [
        AuthCheckGuard,
        { provide: SessionService, useValue: fakeSessionService },
        { provide: AppConfigService, useValue: fakeAppConfigService },
        { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
        { provide: SearchTriggerService, useValue: fakeSearchTriggerService },
      ]
    });
  });

  it('should be created', inject([AuthCheckGuard], (service: AuthCheckGuard) => {
    expect(service).toBeTruthy();
  }));
});
