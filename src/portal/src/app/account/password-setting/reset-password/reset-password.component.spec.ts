import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ResetPasswordComponent } from './reset-password.component';
import { FormsModule } from '@angular/forms';
import { PasswordSettingService } from '../password-setting.service';
import { RouterTestingModule } from '@angular/router/testing';
import { MessageHandlerService } from '../../../shared/message-handler/message-handler.service';

describe('ResetPasswordComponent', () => {
    let component: ResetPasswordComponent;
    let fixture: ComponentFixture<ResetPasswordComponent>;
    let fakePasswordSettingService = null;
    let fakeMessageHandlerService = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule
            ],
            declarations: [ResetPasswordComponent],
            providers: [
                TranslateService,
                { provide: PasswordSettingService, useValue: fakePasswordSettingService },
                { provide: MessageHandlerService, useValue: fakeMessageHandlerService },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ResetPasswordComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
