import { TestBed, inject } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { OidcGuard } from './oidc-guard-active.service';
import { AppConfigService } from '../../services/app-config.service';
import { of } from 'rxjs';
import { UserPermissionService } from '../services';

describe('OidcGuard', () => {
    const fakeAppConfigService = null;
    const fakeUserPermissionService = {
        getPermission() {
            return of(true);
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule],
            providers: [
                OidcGuard,
                { provide: AppConfigService, useValue: fakeAppConfigService },
                {
                    provide: UserPermissionService,
                    useValue: fakeUserPermissionService,
                },
            ],
        });
    });

    it('should ...', inject([OidcGuard], (guard: OidcGuard) => {
        expect(guard).toBeTruthy();
    }));
});
