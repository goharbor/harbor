import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { AppConfigService } from '../../../../services/app-config.service';
import { ConfigurationService } from '../../../../services/config.service';
import { ConfigurationAuthComponent } from './config-auth.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { SystemInfoService } from '../../../../shared/services';
import { ConfigService } from '../config.service';
import { Configuration } from '../config';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('ConfigurationAuthComponent', () => {
    let component: ConfigurationAuthComponent;
    let fixture: ComponentFixture<ConfigurationAuthComponent>;
    let fakeMessageHandlerService = {
        showSuccess: () => null,
    };
    let fakeConfigurationService = {
        saveConfiguration: () => of(null),
        testLDAPServer: () => of(null),
        testOIDCServer: () => of(null),
    };
    let fakeAppConfigService = {
        load: () => of(null),
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
    let fakeSystemInfoService = {
        getSystemInfo: function () {
            return of({
                external_url: 'expectedUrl',
            });
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ConfigurationAuthComponent],
            providers: [
                {
                    provide: MessageHandlerService,
                    useValue: fakeMessageHandlerService,
                },
                {
                    provide: ConfigurationService,
                    useValue: fakeConfigurationService,
                },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: ConfigService, useValue: fakeConfigService },
                { provide: SystemInfoService, useValue: fakeSystemInfoService },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationAuthComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
