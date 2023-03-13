import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SystemInfoService } from '../../../../shared/services';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { of } from 'rxjs';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { SecurityComponent } from './security.component';
import { LOCALE_ID } from '@angular/core';
import { registerLocaleData } from '@angular/common';
import locale_en from '@angular/common/locales/en';
describe('SecurityComponent', () => {
    let component: SecurityComponent;
    let fixture: ComponentFixture<SecurityComponent>;
    const mockedAllowlist = {
        id: 1,
        project_id: 1,
        expires_at: null,
        items: [{ cve_id: 'CVE-2019-1234' }],
    };
    const fakedSystemInfoService = {
        getSystemAllowlist() {
            return of(mockedAllowlist);
        },
        updateSystemAllowlist() {
            return of(true);
        },
    };
    const fakedErrorHandler = {
        info() {
            return null;
        },
    };
    registerLocaleData(locale_en, 'en-us');
    beforeEach(() => {
        TestBed.overrideComponent(SecurityComponent, {
            set: {
                providers: [
                    {
                        provide: LOCALE_ID,
                        useValue: 'en-us',
                    },
                ],
            },
        });
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            providers: [
                { provide: ErrorHandler, useValue: fakedErrorHandler },
                {
                    provide: SystemInfoService,
                    useValue: fakedSystemInfoService,
                },
            ],
            declarations: [SecurityComponent],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(SecurityComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('cancel button and save button should enable', async () => {
        component.systemAllowlist.items.push({ cve_id: 'CVE-2019-456' });
        fixture.detectChanges();
        await fixture.whenStable();
        const cancel: HTMLButtonElement =
            fixture.nativeElement.querySelector('#security_cancel');
        expect(cancel.disabled).toBeFalse();
        const save: HTMLButtonElement =
            fixture.nativeElement.querySelector('#security_save');
        expect(save.disabled).toBeFalse();
    });
    it('save button should works', async () => {
        component.systemAllowlist.items[0].cve_id = 'CVE-2019-789';
        fixture.detectChanges();
        await fixture.whenStable();
        const save: HTMLButtonElement =
            fixture.nativeElement.querySelector('#security_save');
        save.click();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.systemAllowlistOrigin.items[0].cve_id).toEqual(
            'CVE-2019-789'
        );
    });
});
