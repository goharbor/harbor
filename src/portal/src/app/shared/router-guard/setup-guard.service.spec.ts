// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { TestBed, inject } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';
import { of, Observable } from 'rxjs';
import { SetupGuard } from './setup-guard.service';
import { SetupService } from '../../services/setup.service';
import { CommonRoutes } from '../entities/shared.const';

describe('SetupGuard', () => {
    let router: Router;
    const fakeSetupService = {
        isSetupRequired: jasmine.createSpy('isSetupRequired'),
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule],
            providers: [
                SetupGuard,
                { provide: SetupService, useValue: fakeSetupService },
            ],
        });
        router = TestBed.inject(Router);
        spyOn(router, 'navigateByUrl');
    });

    afterEach(() => {
        fakeSetupService.isSetupRequired.calls.reset();
    });

    it('should be created', inject([SetupGuard], (guard: SetupGuard) => {
        expect(guard).toBeTruthy();
    }));

    it('should allow activation when setup is required', inject(
        [SetupGuard],
        (guard: SetupGuard) => {
            fakeSetupService.isSetupRequired.and.returnValue(of(true));
            const result$ = guard.canActivate(null, null);
            if (result$ instanceof Observable) {
                result$.subscribe(allowed => {
                    expect(allowed).toBeTrue();
                    expect(router.navigateByUrl).not.toHaveBeenCalled();
                });
            }
        }
    ));

    it('should redirect to sign-in when setup is not required', inject(
        [SetupGuard],
        (guard: SetupGuard) => {
            fakeSetupService.isSetupRequired.and.returnValue(of(false));
            const result$ = guard.canActivate(null, null);
            if (result$ instanceof Observable) {
                result$.subscribe(allowed => {
                    expect(allowed).toBeFalse();
                    expect(router.navigateByUrl).toHaveBeenCalledWith(
                        CommonRoutes.EMBEDDED_SIGN_IN
                    );
                });
            }
        }
    ));
});
