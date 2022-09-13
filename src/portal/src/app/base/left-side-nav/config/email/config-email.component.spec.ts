import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { ConfigurationService } from '../../../../services/config.service';
import { ConfigurationEmailComponent } from './config-email.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { ConfigService } from '../config.service';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { Configuration } from '../config';

describe('ConfigurationEmailComponent', () => {
    let component: ConfigurationEmailComponent;
    let fixture: ComponentFixture<ConfigurationEmailComponent>;
    let fakeConfigurationService = {
        saveConfiguration: () => of(null),
        testMailServer: () => of(null),
    };
    let fakeMessageHandlerService = {
        showSuccess: () => null,
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
        updateConfig() {},
        resetConfig() {},
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ConfigurationEmailComponent],
            providers: [
                {
                    provide: MessageHandlerService,
                    useValue: fakeMessageHandlerService,
                },
                { provide: ConfigService, useValue: fakeConfigService },
                {
                    provide: ConfigurationService,
                    useValue: fakeConfigurationService,
                },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationEmailComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should save configuration', async () => {
        component.currentConfig.email_host.value = 'smtp.mydomain.com';
        component.currentConfig.email_port.value = 25;
        component.currentConfig.email_from.value = 'smtp.mydomain.com';
        await fixture.whenStable();
        const configEmailSaveBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#config_email_save');
        component.onGoing = true;
        configEmailSaveBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        expect(component.onGoing).toBeFalsy();
    });
    it('should ping test server', async () => {
        component.currentConfig.email_host.value = 'smtp.mydomain.com';
        component.currentConfig.email_port.value = 25;
        component.currentConfig.email_from.value = 'smtp.mydomain.com';
        await fixture.whenStable();
        const pingTestBtn = fixture.nativeElement.querySelector('#ping-test');
        expect(pingTestBtn).toBeTruthy();
        pingTestBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        expect(component.testingMailOnGoing).toBeFalsy();
    });
});
