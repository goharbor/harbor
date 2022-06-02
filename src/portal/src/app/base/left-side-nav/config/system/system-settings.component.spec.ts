import {
    ComponentFixture,
    ComponentFixtureAutoDetect,
    TestBed,
} from '@angular/core/testing';
import { SystemSettingsComponent } from './system-settings.component';
import { SystemInfoService } from '../../../../shared/services';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { of } from 'rxjs';
import { Configuration, StringValueItem } from '../config';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { ConfigService } from '../config.service';
import { AppConfigService } from '../../../../services/app-config.service';
describe('SystemSettingsComponent', () => {
    let component: SystemSettingsComponent;
    let fixture: ComponentFixture<SystemSettingsComponent>;
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
        getSystemInfo() {
            return of({});
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
    const fakeConfigService = {
        config: new Configuration(),
        getConfig() {
            return this.config;
        },
        setConfig(c) {
            this.config = c;
        },
        getOriginalConfig() {
            return new Configuration();
        },
        getLoadingConfigStatus() {
            return false;
        },
        confirmUnsavedChanges() {},
        updateConfig() {},
        resetConfig() {},
    };
    const fakedAppConfigService = {
        getConfig() {
            return {};
        },
        load() {
            return of(null);
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            providers: [
                { provide: AppConfigService, useValue: fakedAppConfigService },
                { provide: ConfigService, useValue: fakeConfigService },
                { provide: ErrorHandler, useValue: fakedErrorHandler },
                {
                    provide: SystemInfoService,
                    useValue: fakedSystemInfoService,
                },
                // open auto detect
                { provide: ComponentFixtureAutoDetect, useValue: true },
            ],
            declarations: [SystemSettingsComponent],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(SystemSettingsComponent);
        component = fixture.componentInstance;
        component.currentConfig.auth_mode = new StringValueItem(
            'db_auth',
            false
        );
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('cancel button should works', () => {
        const spy: jasmine.Spy = spyOn(
            fakeConfigService,
            'confirmUnsavedChanges'
        ).and.returnValue(undefined);
        component.systemAllowlist.items.push({ cve_id: 'CVE-2019-456' });
        const readOnly: HTMLElement =
            fixture.nativeElement.querySelector('#repoReadOnly');
        readOnly.click();
        fixture.detectChanges();
        const cancel: HTMLButtonElement = fixture.nativeElement.querySelector(
            '#config_system_cancel'
        );
        cancel.click();
        fixture.detectChanges();
        expect(spy.calls.count()).toEqual(1);
    });
    it('save button should works', () => {
        component.systemAllowlist.items[0].cve_id = 'CVE-2019-789';
        const readOnly: HTMLElement =
            fixture.nativeElement.querySelector('#repoReadOnly');
        readOnly.click();
        fixture.detectChanges();
        const save: HTMLButtonElement = fixture.nativeElement.querySelector(
            '#config_system_save'
        );
        save.click();
        fixture.detectChanges();
        expect(component.systemAllowlistOrigin.items[0].cve_id).toEqual(
            'CVE-2019-789'
        );
    });
});
