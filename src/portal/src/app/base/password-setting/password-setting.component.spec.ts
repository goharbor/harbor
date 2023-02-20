import { ComponentFixture, TestBed } from '@angular/core/testing';
import { PasswordSettingService } from './password-setting.service';
import { SessionService } from '../../shared/services/session.service';
import { MessageHandlerService } from '../../shared/services/message-handler.service';
import { PasswordSettingComponent } from './password-setting.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { InlineAlertComponent } from '../../shared/components/inline-alert/inline-alert.component';
import { SharedTestingModule } from '../../shared/shared.module';

describe('PasswordSettingComponent', () => {
    let component: PasswordSettingComponent;
    let fixture: ComponentFixture<PasswordSettingComponent>;
    let fakePasswordSettingService = {
        changePassword: () => of(null),
    };
    let fakeSessionService = {
        getCurrentUser: () => true,
    };
    let fakeMessageHandlerService = {
        showSuccess: () => {},
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [PasswordSettingComponent, InlineAlertComponent],
            providers: [
                {
                    provide: PasswordSettingService,
                    useValue: fakePasswordSettingService,
                },
                { provide: SessionService, useValue: fakeSessionService },
                {
                    provide: MessageHandlerService,
                    useValue: fakeMessageHandlerService,
                },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(PasswordSettingComponent);
        component = fixture.componentInstance;
        component.inlineAlert =
            TestBed.createComponent(InlineAlertComponent).componentInstance;
        component.oldPwd = 'Harbor12345';
        component.open();
        fixture.autoDetectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should verify new Password invalid', async () => {
        await fixture.whenStable();
        const newPasswordInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#newPassword');
        newPasswordInput.value = 'HarborHarbor';
        newPasswordInput.dispatchEvent(new Event('input'));
        await fixture.whenStable();
        newPasswordInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        const newPasswordInputError: any =
            fixture.nativeElement.querySelector('#newPassword-error');
        expect(newPasswordInputError.innerText).toEqual('TOOLTIP.PASSWORD');
    });
    it('should verify new Password valid', async () => {
        await fixture.whenStable();
        const newPasswordInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#newPassword');
        newPasswordInput.value = 'Harbor123456';
        newPasswordInput.dispatchEvent(new Event('input'));
        await fixture.whenStable();
        newPasswordInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        const newPasswordInputError: any =
            fixture.nativeElement.querySelector('#newPassword-error');
        expect(newPasswordInputError).toBeNull();
    });
    it('should verify comfirm Password invalid', async () => {
        await fixture.whenStable();
        const newPasswordInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#newPassword');
        newPasswordInput.value = 'Harbor123456';
        newPasswordInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        const reNewPasswordInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#reNewPassword');
        reNewPasswordInput.value = 'Harbor12345';
        reNewPasswordInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        const reNewPasswordInputError: any =
            fixture.nativeElement.querySelector('#reNewPassword-error');
        expect(reNewPasswordInputError.innerText).toEqual(
            'TOOLTIP.CONFIRM_PWD'
        );
    });
    it('should verify comfirm Password valid', async () => {
        await fixture.whenStable();
        const newPasswordInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#newPassword');
        newPasswordInput.value = 'Harbor123456';
        newPasswordInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        const reNewPasswordInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#reNewPassword');
        reNewPasswordInput.value = 'Harbor123456';
        reNewPasswordInput.dispatchEvent(new Event('input'));
        reNewPasswordInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        const reNewPasswordInputError: any =
            fixture.nativeElement.querySelector('#reNewPassword-error');
        expect(reNewPasswordInputError).toBeNull();
    });
    it('should save new password', async () => {
        await fixture.whenStable();
        const newPasswordInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#newPassword');
        newPasswordInput.value = 'Harbor123456';
        newPasswordInput.dispatchEvent(new Event('input'));
        newPasswordInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        const reNewPasswordInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#reNewPassword');
        reNewPasswordInput.value = 'Harbor123456';
        reNewPasswordInput.dispatchEvent(new Event('input'));
        reNewPasswordInput.dispatchEvent(new Event('blur'));
        await fixture.whenStable();
        const okBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#ok-btn');
        okBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();

        const newPasswordInput1: HTMLInputElement =
            fixture.nativeElement.querySelector('#newPassword');
        expect(newPasswordInput1).toBeNull();
    });
});
