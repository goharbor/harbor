import { Component, ViewChild } from '@angular/core';
import { Router } from '@angular/router';
import { NgForm } from '@angular/forms';

import { PasswordSettingService } from './password-setting.service';
import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';

@Component({
    selector: 'reset-password',
    templateUrl: "reset-password.component.html",
    styleUrls: ['password.component.css']
})
export class ResetPasswordComponent {
    opened: boolean = true;
    private onGoing: boolean = false;
    private password: string = "";
    private validationState: any = {};

    @ViewChild("resetPwdForm") resetPwdForm: NgForm;
    @ViewChild(InlineAlertComponent)
    private inlineAlert: InlineAlertComponent;

    constructor(private pwdService: PasswordSettingService) { }

    public get showProgress(): boolean {
        return this.onGoing;
    }

    public get isValid(): boolean {
        return this.resetPwdForm && this.resetPwdForm.valid && this.samePassword();
    }

    public getValidationState(key: string): boolean {
        return this.validationState && this.validationState[key];
    }

    public open(): void {
        this.opened = true;
        this.resetPwdForm.resetForm();
    }

    public close(): void {
        this.opened = false;
    }

    public send(): void {
        //Double confirm to avoid improper situations
        if (!this.password) {
            return;
        }

        if (!this.isValid) {
            return;
        }

        this.onGoing = true;
    }

    public handleValidation(key: string, flag: boolean): void {
        if (flag) {
            if(!this.validationState[key]){
                this.validationState[key] = true;
            }
        } else {
            this.validationState[key] = this.getControlValidationState(key)
        }
    }

    private getControlValidationState(key: string): boolean {
        if (this.resetPwdForm) {
            let control = this.resetPwdForm.controls[key];
            if (control) {
                return control.valid;
            }
        }

        return false;
    }

    private samePassword(): boolean {
        if (this.resetPwdForm) {
            let control1 = this.resetPwdForm.controls["newPassword"];
            let control2 = this.resetPwdForm.controls["reNewPassword"];
            if (control1 && control2) {
                return control1.value == control2.value;
            }
        }

        return false;
    }
}