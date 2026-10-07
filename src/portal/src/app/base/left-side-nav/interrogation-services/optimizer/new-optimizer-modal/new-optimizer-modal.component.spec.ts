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
import {
    ComponentFixture,
    ComponentFixtureAutoDetect,
    fakeAsync,
    TestBed,
    tick,
} from '@angular/core/testing';
import { ClrLoadingState } from '@clr/angular';
import { NewOptimizerModalComponent } from './new-optimizer-modal.component';
import { MessageHandlerService } from '../../../../../shared/services/message-handler.service';
import { NewOptimizerFormComponent } from '../new-optimizer-form/new-optimizer-form.component';
import { UntypedFormBuilder } from '@angular/forms';
import { of, Subscription } from 'rxjs';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { Optimizer } from '../optimizer';
import { OptimizerService } from '../../../../../../../ng-swagger-gen/services/optimizer.service';

describe('NewOptimizerModalComponent', () => {
    let component: NewOptimizerModalComponent;
    let fixture: ComponentFixture<NewOptimizerModalComponent>;

    let mockOptimizer1: Optimizer = {
        name: 'test1',
        description: 'just a sample',
        url: 'http://168.0.0.1',
        auth: '',
    };
    let fakedConfigOptimizerService = {
        listOptimizers() {
            return of([mockOptimizer1]);
        },
        pingOptimizer() {
            return of(true).pipe(delay(200));
        },
        createOptimizer() {
            return of(true).pipe(delay(200));
        },
        updateOptimizer() {
            return of(true).pipe(delay(200));
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [NewOptimizerFormComponent, NewOptimizerModalComponent],
            providers: [
                MessageHandlerService,
                {
                    provide: OptimizerService,
                    useValue: fakedConfigOptimizerService,
                },
                UntypedFormBuilder,
                // open auto detect
                { provide: ComponentFixtureAutoDetect, useValue: true },
            ],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(NewOptimizerModalComponent);
        component = fixture.componentInstance;
        component.opened = true;
        component.newOptimizerFormComponent.checkNameSubscribe =
            new Subscription();
        component.newOptimizerFormComponent.checkEndpointUrlSubscribe =
            new Subscription();
        fixture.detectChanges();
    });
    afterEach(() => {
        if (
            component &&
            component.newOptimizerFormComponent &&
            component.newOptimizerFormComponent.checkNameSubscribe
        ) {
            component.newOptimizerFormComponent.checkNameSubscribe.unsubscribe();
            component.newOptimizerFormComponent.checkNameSubscribe = null;
        }
        if (
            component &&
            component.newOptimizerFormComponent &&
            component.newOptimizerFormComponent.checkEndpointUrlSubscribe
        ) {
            component.newOptimizerFormComponent.checkEndpointUrlSubscribe.unsubscribe();
            component.newOptimizerFormComponent.checkEndpointUrlSubscribe = null;
        }
    });
    it('should creat', () => {
        expect(component).toBeTruthy();
    });
    it('should be add mode', () => {
        component.isEdit = false;
        fixture.detectChanges();
        let el = fixture.nativeElement.querySelector('#button-add');
        expect(el).toBeTruthy();
    });
    it('should be edit mode', fakeAsync(() => {
        component.isEdit = true;
        fixture.detectChanges();
        let el = fixture.nativeElement.querySelector('#button-save');
        expect(el).toBeTruthy();
        // set origin value
        component.originValue = mockOptimizer1;
        component.editOptimizer = {};
        // input same value to origin
        fixture.nativeElement.querySelector('#optimizer-name').value = 'test2';
        fixture.nativeElement.querySelector('#description').value =
            'just a sample';
        fixture.nativeElement.querySelector('#optimizer-endpoint').value =
            'http://168.0.0.1';
        fixture.nativeElement.querySelector('#optimizer-authorization').value =
            '';
        fixture.nativeElement
            .querySelector('#optimizer-name')
            .dispatchEvent(new Event('input'));
        fixture.nativeElement
            .querySelector('#description')
            .dispatchEvent(new Event('input'));
        fixture.nativeElement
            .querySelector('#optimizer-endpoint')
            .dispatchEvent(new Event('input'));
        fixture.nativeElement
            .querySelector('#optimizer-authorization')
            .dispatchEvent(new Event('input'));
        // save button should not be disabled
        expect(component.validForSaving).toBeTruthy();
        fixture.nativeElement.querySelector('#optimizer-name').value = 'test3';
        fixture.nativeElement
            .querySelector('#optimizer-name')
            .dispatchEvent(new Event('input'));
        fixture.detectChanges();
        expect(component.validForSaving).toBeTruthy();
        el.click();
        el.dispatchEvent(new Event('click'));
        tick(10000);
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(component.opened).toBeFalsy();
        });
    }));
    it('test connection button should not be disabled', fakeAsync(() => {
        let nameInput = fixture.nativeElement.querySelector('#optimizer-name');
        nameInput.value = 'test2';
        nameInput.dispatchEvent(new Event('input'));
        let urlInput = fixture.nativeElement.querySelector('#optimizer-endpoint');
        urlInput.value = 'http://168.0.0.1';
        urlInput.dispatchEvent(new Event('input'));
        expect(component.canTestEndpoint).toBeTruthy();
        let el = fixture.nativeElement.querySelector('#button-test');
        el.click();
        el.dispatchEvent(new Event('click'));
        expect(component.checkBtnState).toBe(ClrLoadingState.LOADING);
        tick(10000);
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(component.checkBtnState).toBe(ClrLoadingState.SUCCESS);
        });
    }));
    it('add button should not be disabled', fakeAsync(() => {
        fixture.nativeElement.querySelector('#optimizer-name').value = 'test2';
        fixture.nativeElement.querySelector('#optimizer-endpoint').value =
            'http://168.0.0.1';
        let authInput = fixture.nativeElement.querySelector(
            '#optimizer-authorization'
        );
        authInput.value = 'Basic';
        authInput.dispatchEvent(new Event('change'));
        let usernameInput =
            fixture.nativeElement.querySelector('#optimizer-username');
        let passwordInput =
            fixture.nativeElement.querySelector('#optimizer-password');
        expect(usernameInput).toBeTruthy();
        expect(passwordInput).toBeTruthy();
        usernameInput.value = 'user';
        passwordInput.value = '12345';
        usernameInput.dispatchEvent(new Event('input'));
        passwordInput.dispatchEvent(new Event('input'));
        let el = fixture.nativeElement.querySelector('#button-add');
        expect(component.valid).toBeFalsy();
        el.click();
        el.dispatchEvent(new Event('click'));
        tick(10000);
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(component.opened).toBeFalsy();
        });
    }));
});
