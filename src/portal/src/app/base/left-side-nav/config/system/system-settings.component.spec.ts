import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SystemSettingsComponent } from './system-settings.component';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { of } from 'rxjs';
import { Configuration } from '../config';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { ConfigService } from '../config.service';
import { AppConfigService } from '../../../../services/app-config.service';

describe('SystemSettingsComponent', () => {
    let component: SystemSettingsComponent;
    let fixture: ComponentFixture<SystemSettingsComponent>;
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
        saveConfiguration() {
            return of(null);
        },
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
            ],
            declarations: [SystemSettingsComponent],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(SystemSettingsComponent);
        component = fixture.componentInstance;
        fixture.autoDetectChanges(true);
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('cancel button should work', () => {
        const spy: jasmine.Spy = spyOn(component, 'cancel').and.returnValue(
            undefined
        );
        const cancel: HTMLButtonElement = fixture.nativeElement.querySelector(
            '#config_system_cancel'
        );
        cancel.dispatchEvent(new Event('click'));
        expect(spy.calls.count()).toEqual(1);
    });
    it('save button should work', () => {
        const input = fixture.nativeElement.querySelector('#robotNamePrefix');
        input.value = 'test';
        input.dispatchEvent(new Event('input'));
        const save: HTMLButtonElement = fixture.nativeElement.querySelector(
            '#config_system_save'
        );
        save.dispatchEvent(new Event('click'));
        expect(input.value).toEqual('test');
    });
});
