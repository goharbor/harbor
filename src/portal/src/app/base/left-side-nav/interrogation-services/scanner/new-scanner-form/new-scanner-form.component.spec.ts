import {
    ComponentFixture,
    ComponentFixtureAutoDetect,
    fakeAsync,
    TestBed,
    tick,
} from '@angular/core/testing';
import { NewScannerFormComponent } from './new-scanner-form.component';
import { FormBuilder } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { of } from 'rxjs';
import { TranslateService } from '@ngx-translate/core';
import { delay } from 'rxjs/operators';
import { ScannerService } from '../../../../../../../ng-swagger-gen/services/scanner.service';

describe('NewScannerFormComponent', () => {
    let mockScanner1 = {
        name: 'test1',
        description: 'just a sample',
        version: '1.0.0',
        url: 'http://168.0.0.1',
    };
    let component: NewScannerFormComponent;
    let fixture: ComponentFixture<NewScannerFormComponent>;
    let fakedConfigScannerService = {
        listScanners() {
            return of([mockScanner1]).pipe(delay(500));
        },
        getScannersByEndpointUrl() {
            return of([mockScanner1]).pipe(delay(500));
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [
                SharedTestingModule,
                BrowserAnimationsModule,
                ClarityModule,
            ],
            declarations: [NewScannerFormComponent],
            providers: [
                FormBuilder,
                TranslateService,
                {
                    provide: ScannerService,
                    useValue: fakedConfigScannerService,
                },
                // open auto detect
                { provide: ComponentFixtureAutoDetect, useValue: true },
            ],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(NewScannerFormComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });
    it('should creat', () => {
        expect(component).toBeTruthy();
    });
    it('should show "name is required"', () => {
        let nameInput = fixture.nativeElement.querySelector('#scanner-name');
        nameInput.value = '';
        nameInput.dispatchEvent(new Event('input'));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        let el = fixture.nativeElement.querySelector('clr-control-error');
        expect(el).toBeTruthy();
    });
    it('name should be existed', fakeAsync(() => {
        let nameInput = fixture.nativeElement.querySelector('#scanner-name');
        nameInput.value = 'test1';
        nameInput.dispatchEvent(new Event('input'));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        tick(20000);
        const el = fixture.nativeElement.querySelector('#name-error');
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(el).toBeTruthy();
        });
    }));
    it('name should be valid', fakeAsync(() => {
        let nameInput = fixture.nativeElement.querySelector('#scanner-name');
        nameInput.value = 'test2';
        nameInput.dispatchEvent(new Event('input'));
        nameInput.blur();
        nameInput.dispatchEvent(new Event('blur'));
        tick(20000);
        const el = fixture.nativeElement.querySelector('#name-error');
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(el).toBeFalsy();
        });
    }));

    it('endpoint url should be valid', fakeAsync(() => {
        let nameInput = fixture.nativeElement.querySelector('#scanner-name');
        nameInput.value = 'test2';
        let urlInput = fixture.nativeElement.querySelector('#scanner-endpoint');
        urlInput.value = 'http://168.0.0.2';
        urlInput.dispatchEvent(new Event('input'));
        urlInput.blur();
        urlInput.dispatchEvent(new Event('blur'));
        tick(20000);
        const el = fixture.nativeElement.querySelector('#endpoint-error');
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(el).toBeFalsy();
        });
    }));

    it('auth should be valid', () => {
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
        let el = fixture.nativeElement.querySelector('#pwd-error');
        expect(el).toBeFalsy();
    });
});
