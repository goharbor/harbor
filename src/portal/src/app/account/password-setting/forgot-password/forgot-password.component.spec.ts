import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { FormsModule } from '@angular/forms';
import { ForgotPasswordComponent } from './forgot-password.component';
import { ClarityModule } from "@clr/angular";
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { PasswordSettingService } from '../password-setting.service';

describe('ForgotPasswordComponent', () => {
    let component: ForgotPasswordComponent;
    let fixture: ComponentFixture<ForgotPasswordComponent>;
    let fakePasswordSettingService = null;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [ForgotPasswordComponent],
            imports: [
                FormsModule,
                ClarityModule,
                TranslateModule.forRoot()
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                TranslateService,
                { provide: PasswordSettingService, useValue: fakePasswordSettingService }
            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(ForgotPasswordComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
