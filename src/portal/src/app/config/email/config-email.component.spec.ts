import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { ConfirmMessageHandler } from '../config.msg.utils';
import { ConfigurationService } from '../config.service';
import { ConfigurationEmailComponent } from './config-email.component';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { clone } from '../../../lib/utils/utils';
import { of } from 'rxjs';

describe('ConfigurationEmailComponent', () => {
    let component: ConfigurationEmailComponent;
    let fixture: ComponentFixture<ConfigurationEmailComponent>;
    let fakeConfigurationService = {
        saveConfiguration: () => of(null),
        testMailServer: () => of(null)
    };
    let fakeMessageHandlerService = {
        showSuccess: () => null
    };
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                FormsModule
            ],
            declarations: [ConfigurationEmailComponent],
            providers: [
                { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
                TranslateService,
                { provide: ConfirmMessageHandler, useValue: null },
                { provide: ConfigurationService, useValue: fakeConfigurationService }
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationEmailComponent);
        component = fixture.componentInstance;
        (component as any).originalConfig = clone(component.currentConfig);
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
        const configEmailSaveBtn: HTMLButtonElement = fixture.nativeElement.querySelector("#config_email_save");
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
        const pingTestBtn = fixture.nativeElement.querySelector("#ping-test");
        expect(pingTestBtn).toBeTruthy();
        pingTestBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        expect(component.testingMailOnGoing).toBeFalsy();
    });
});
