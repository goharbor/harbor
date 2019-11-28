import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { ConfirmMessageHandler } from '../config.msg.utils';
import { AppConfigService } from '../../app-config.service';
import { ConfigurationService } from '../config.service';
import { ConfigurationAuthComponent } from './config-auth.component';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { of } from 'rxjs';
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { SystemInfoService } from "../../../lib/services";

describe('ConfigurationAuthComponent', () => {
    let component: ConfigurationAuthComponent;
    let fixture: ComponentFixture<ConfigurationAuthComponent>;
    let fakeMessageHandlerService = null;
    let fakeConfigurationService = null;
    let fakeAppConfigService = null;
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
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
