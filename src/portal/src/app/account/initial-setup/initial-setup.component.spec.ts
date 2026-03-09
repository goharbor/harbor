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
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { InitialSetupComponent } from './initial-setup.component';
import { SetupService } from '../../services/setup.service';
import { SkinableConfig } from '../../services/skinable-config.service';
import { CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { throwError as observableThrowError } from 'rxjs/internal/observable/throwError';
import { HttpErrorResponse } from '@angular/common/http';
import { RouterTestingModule } from '@angular/router/testing';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { TranslateModule, TranslateService } from '@ngx-translate/core';

describe('InitialSetupComponent', () => {
    let component: InitialSetupComponent;
    let fixture: ComponentFixture<InitialSetupComponent>;
    const mockedSetupService = {
        isSetupRequired() {
            return of(true);
        },
        setupAdminPassword(password: string) {
            return of({ ok: true });
        },
        clearCache() {},
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                RouterTestingModule,
                ClarityModule,
                FormsModule,
            ],
            declarations: [InitialSetupComponent],
            providers: [
                TranslateService,
                {
                    provide: SetupService,
                    useValue: mockedSetupService,
                },
                {
                    provide: SkinableConfig,
                    useValue: {
                        getSkinConfig: function () {
                            return {
                                loginBgImg: 'abc',
                                appTitle: 'Harbor',
                            };
                        },
                    },
                },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(InitialSetupComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should validate password strength correctly', () => {
        component.password = 'short';
        expect(component.isPasswordStrong()).toBeFalse();

        component.password = 'nouppercase1';
        expect(component.isPasswordStrong()).toBeFalse();

        component.password = 'NOLOWERCASE1';
        expect(component.isPasswordStrong()).toBeFalse();

        component.password = 'NoNumbers';
        expect(component.isPasswordStrong()).toBeFalse();

        component.password = 'Harbor12345';
        expect(component.isPasswordStrong()).toBeTrue();
    });

    it('should detect password mismatch', () => {
        component.password = 'Harbor12345';
        component.confirmPassword = 'Harbor12346';
        expect(component.passwordMismatch).toBeTrue();

        component.confirmPassword = 'Harbor12345';
        expect(component.passwordMismatch).toBeFalse();
    });

    it('should handle setup error (403 already completed)', async () => {
        const setupService = TestBed.get<SetupService>(SetupService);
        spyOn(setupService, 'setupAdminPassword').and.returnValue(
            observableThrowError(
                new HttpErrorResponse({
                    error: 'Setup already completed',
                    status: 403,
                })
            )
        );

        component.password = 'Harbor12345';
        component.confirmPassword = 'Harbor12345';
        component.submitSetup();

        fixture.detectChanges();
        await fixture.whenStable();

        expect(component.isError).toBeTrue();
        expect(component.errorMessage).toEqual(
            'INITIAL_SETUP.ERROR_ALREADY_COMPLETED'
        );
    });

    it('should handle setup error (400 weak password)', async () => {
        const setupService = TestBed.get<SetupService>(SetupService);
        spyOn(setupService, 'setupAdminPassword').and.returnValue(
            observableThrowError(
                new HttpErrorResponse({
                    error: 'Weak password',
                    status: 400,
                })
            )
        );

        component.password = 'Harbor12345';
        component.confirmPassword = 'Harbor12345';
        component.submitSetup();

        fixture.detectChanges();
        await fixture.whenStable();

        expect(component.isError).toBeTrue();
        expect(component.errorMessage).toEqual(
            'INITIAL_SETUP.ERROR_WEAK_PASSWORD'
        );
    });
});
