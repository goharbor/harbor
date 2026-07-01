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
import { AddRobotComponent } from './add-robot.component';
import { of } from 'rxjs';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { delay } from 'rxjs/operators';
import { RobotService } from '../../../../../../ng-swagger-gen/services/robot.service';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('AddRobotComponent', () => {
    let component: AddRobotComponent;
    let fixture: ComponentFixture<AddRobotComponent>;
    const fakedRobotService = {
        ListRobot() {
            return of([]).pipe(delay(0));
        },
    };
    const fakedMessageHandlerService = {
        showSuccess() {},
        error() {},
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [AddRobotComponent],
            imports: [SharedTestingModule],
            providers: [
                OperationService,
                { provide: RobotService, useValue: fakedRobotService },
                {
                    provide: MessageHandlerService,
                    useValue: fakedMessageHandlerService,
                },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddRobotComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    describe('secret validation', () => {
        it('validateSecret() should set isSecretDirty and populate errors for invalid secret', () => {
            component.userProvidedSecret = 'abc';
            component.validateSecret();
            expect(component.isSecretDirty).toBe(true);
            expect(component.secretValidationErrors.length).toBeGreaterThan(0);
        });

        it('validateSecret() should clear errors for valid secret', () => {
            component.userProvidedSecret = 'Harbor12345';
            component.validateSecret();
            expect(component.secretValidationErrors.length).toBe(0);
        });

        it('validateSecret() should clear errors when secret is empty', () => {
            component.userProvidedSecret = '';
            component.validateSecret();
            expect(component.secretValidationErrors.length).toBe(0);
        });

        it('toggleSecretVisibility() should toggle showSecretPassword flag', () => {
            expect(component.showSecretPassword).toBe(false);
            component.toggleSecretVisibility();
            expect(component.showSecretPassword).toBe(true);
            component.toggleSecretVisibility();
            expect(component.showSecretPassword).toBe(false);
        });

        it('isSecretInputValid() should return falsy for empty secret', () => {
            component.userProvidedSecret = '';
            expect(component.isSecretInputValid()).toBeFalsy();
        });

        it('isSecretInputValid() should return falsy for invalid secret', () => {
            component.userProvidedSecret = 'abc';
            expect(component.isSecretInputValid()).toBeFalsy();
        });

        it('isSecretInputValid() should return truthy for valid secret', () => {
            component.userProvidedSecret = 'Harbor12345';
            expect(component.isSecretInputValid()).toBeTruthy();
        });

        it('secretsMatch() should return falsy when confirm secret is empty', () => {
            component.userProvidedSecret = 'Harbor12345';
            component.userProvidedSecretConfirm = '';
            expect(component.secretsMatch()).toBeFalsy();
        });

        it('secretsMatch() should return falsy when secrets do not match', () => {
            component.userProvidedSecret = 'Harbor12345';
            component.userProvidedSecretConfirm = 'Harbor123456';
            expect(component.secretsMatch()).toBeFalsy();
        });

        it('secretsMatch() should return truthy when secrets match', () => {
            component.userProvidedSecret = 'Harbor12345';
            component.userProvidedSecretConfirm = 'Harbor12345';
            expect(component.secretsMatch()).toBeTruthy();
        });
    });
});
