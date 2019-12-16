import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { ConfirmMessageHandler } from '../config.msg.utils';
import { AppConfigService } from '../../app-config.service';
import { ConfigurationService } from '../config.service';
import { ConfigurationAuthComponent } from './config-auth.component';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA, SimpleChange } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { of } from 'rxjs';
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { SystemInfoService } from "../../../lib/services";
import { Configuration } from '../../../lib/components/config/config';
import { clone } from '../../../lib/utils/utils';
import { CONFIG_AUTH_MODE } from '../../../lib/entities/shared.const';

describe('ConfigurationAuthComponent', () => {
    let component: ConfigurationAuthComponent;
    let fixture: ComponentFixture<ConfigurationAuthComponent>;
    let fakeMessageHandlerService = {
        showSuccess: () => null
    };
    let fakeConfigurationService = {
        saveConfiguration: () => of(null),
        testLDAPServer: () => of(null),
        testOIDCServer: () => of(null)
    };
    let fakeAppConfigService = {
        load: () => of(null)
    };
    let fakeConfirmMessageService = null;
    let fakeSystemInfoService = {
        getSystemInfo: function () {
            return of({
                external_url: "expectedUrl"
            });
        }
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                FormsModule
            ],
            declarations: [ConfigurationAuthComponent],
            providers: [
                ErrorHandler,
                TranslateService,
                { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
                { provide: ConfigurationService, useValue: fakeConfigurationService },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: ConfirmMessageHandler, useValue: fakeConfirmMessageService },
                { provide: SystemInfoService, useValue: fakeSystemInfoService }
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationAuthComponent);
        component = fixture.componentInstance;
        (component as any).originalConfig = clone(component.currentConfig);
        fixture.autoDetectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should save configuration', async () => {
        const selfRegInput: HTMLInputElement = fixture.nativeElement.querySelector("#selfReg");
        selfRegInput.dispatchEvent(new Event('click'));
        component.currentConfig.self_registration.value = true;
        await fixture.whenStable();
        const configAuthSaveBtn: HTMLButtonElement = fixture.nativeElement.querySelector("#config_auth_save");
        component.onGoing = true;
        configAuthSaveBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        expect(component.onGoing).toBeFalsy();
    });
    it('should select ldap or uaa', () => {
        component.handleOnChange({target: {value: 'ldap_auth'}});
        expect(component.currentConfig.self_registration.value).toEqual(false);
    });
    it('should ping test server when ldap', async () => {
        component.currentConfig.auth_mode.value = CONFIG_AUTH_MODE.LDAP_AUTH;
        component.currentConfig.ldap_scope.value = 123456;
        await fixture.whenStable();
        const pingTestBtn = fixture.nativeElement.querySelector("#ping-test");
        expect(pingTestBtn).toBeTruthy();
        pingTestBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        expect(component.testingOnGoing).toBeFalsy();
    });
    it('should ping test server when oidc', async () => {
        component.currentConfig.auth_mode.value = CONFIG_AUTH_MODE.OIDC_AUTH;
        await fixture.whenStable();
        const pingTestBtn = fixture.nativeElement.querySelector("#ping-test");
        expect(pingTestBtn).toBeTruthy();
        pingTestBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        expect(component.testingOnGoing).toBeFalsy();
    });
});
