import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { PasswordSettingService } from './password-setting.service';
import { SessionService } from '../../shared/session.service';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { PasswordSettingComponent } from './password-setting.component';
import { ClarityModule } from "@clr/angular";
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { FormsModule } from '@angular/forms';
import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';

describe('PasswordSettingComponent', () => {
    let component: PasswordSettingComponent;
    let fixture: ComponentFixture<PasswordSettingComponent>;
    let fakePasswordSettingService = null;
    let fakeSessionService = null;
    let fakeMessageHandlerService = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule
            ],
            declarations: [PasswordSettingComponent, InlineAlertComponent],
            providers: [
                TranslateService,
                { provide: PasswordSettingService, useValue: fakePasswordSettingService },
                { provide: SessionService, useValue: fakeSessionService },
                { provide: MessageHandlerService, useValue: fakeMessageHandlerService }
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(PasswordSettingComponent);
        component = fixture.componentInstance;
        component.inlineAlert =
        TestBed.createComponent(InlineAlertComponent).componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
