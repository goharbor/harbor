import {
    ComponentFixture,
    ComponentFixtureAutoDetect,
    fakeAsync,
    TestBed,
    tick,
} from '@angular/core/testing';
import { ClrLoadingState } from '@clr/angular';
import { NewScannerModalComponent } from './new-scanner-modal.component';
import { MessageHandlerService } from '../../../../../shared/services/message-handler.service';
import { NewScannerFormComponent } from '../new-scanner-form/new-scanner-form.component';
import { FormBuilder } from '@angular/forms';
import { of, Subscription } from 'rxjs';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { Scanner } from '../scanner';
import { ScannerService } from '../../../../../../../ng-swagger-gen/services/scanner.service';

describe('NewScannerModalComponent', () => {
    let component: NewScannerModalComponent;
    let fixture: ComponentFixture<NewScannerModalComponent>;

    let mockScanner1: Scanner = {
        name: 'test1',
        description: 'just a sample',
        url: 'http://168.0.0.1',
        auth: '',
    };
    let fakedConfigScannerService = {
        listScanners() {
            return of([mockScanner1]);
        },
        pingScanner() {
            return of(true).pipe(delay(200));
        },
        createScanner() {
            return of(true).pipe(delay(200));
        },
        updateScanner() {
            return of(true).pipe(delay(200));
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [NewScannerFormComponent, NewScannerModalComponent],
            providers: [
                MessageHandlerService,
                {
                    provide: ScannerService,
                    useValue: fakedConfigScannerService,
                },
                FormBuilder,
                // open auto detect
                { provide: ComponentFixtureAutoDetect, useValue: true },
            ],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(NewScannerModalComponent);
        component = fixture.componentInstance;
        component.opened = true;
        component.newScannerFormComponent.checkNameSubscribe =
            new Subscription();
        component.newScannerFormComponent.checkEndpointUrlSubscribe =
            new Subscription();
        fixture.detectChanges();
    });
    afterEach(() => {
        if (
            component &&
            component.newScannerFormComponent &&
            component.newScannerFormComponent.checkNameSubscribe
        ) {
            component.newScannerFormComponent.checkNameSubscribe.unsubscribe();
            component.newScannerFormComponent.checkNameSubscribe = null;
        }
        if (
            component &&
            component.newScannerFormComponent &&
            component.newScannerFormComponent.checkEndpointUrlSubscribe
        ) {
            component.newScannerFormComponent.checkEndpointUrlSubscribe.unsubscribe();
            component.newScannerFormComponent.checkEndpointUrlSubscribe = null;
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
        component.originValue = mockScanner1;
        component.editScanner = {};
        // input same value to origin
        fixture.nativeElement.querySelector('#scanner-name').value = 'test2';
        fixture.nativeElement.querySelector('#description').value =
            'just a sample';
        fixture.nativeElement.querySelector('#scanner-endpoint').value =
            'http://168.0.0.1';
        fixture.nativeElement.querySelector('#scanner-authorization').value =
            '';
        fixture.nativeElement
            .querySelector('#scanner-name')
            .dispatchEvent(new Event('input'));
        fixture.nativeElement
            .querySelector('#description')
            .dispatchEvent(new Event('input'));
        fixture.nativeElement
            .querySelector('#scanner-endpoint')
            .dispatchEvent(new Event('input'));
        fixture.nativeElement
            .querySelector('#scanner-authorization')
            .dispatchEvent(new Event('input'));
        // save button should not be disabled
        expect(component.validForSaving).toBeTruthy();
        fixture.nativeElement.querySelector('#scanner-name').value = 'test3';
        fixture.nativeElement
            .querySelector('#scanner-name')
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
        let nameInput = fixture.nativeElement.querySelector('#scanner-name');
        nameInput.value = 'test2';
        nameInput.dispatchEvent(new Event('input'));
        let urlInput = fixture.nativeElement.querySelector('#scanner-endpoint');
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
        fixture.nativeElement.querySelector('#scanner-name').value = 'test2';
        fixture.nativeElement.querySelector('#scanner-endpoint').value =
            'http://168.0.0.1';
        let authInput = fixture.nativeElement.querySelector(
            '#scanner-authorization'
        );
        authInput.value = 'Basic';
        authInput.dispatchEvent(new Event('change'));
        let usernameInput =
            fixture.nativeElement.querySelector('#scanner-username');
        let passwordInput =
            fixture.nativeElement.querySelector('#scanner-password');
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
