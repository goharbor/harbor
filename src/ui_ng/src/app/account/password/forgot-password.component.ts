import { Component, ViewChild } from '@angular/core';
import { Router } from '@angular/router';
import { NgForm } from '@angular/forms';

import { PasswordSettingService } from './password-setting.service';
import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';

@Component({
    selector: 'forgot-password',
    templateUrl: "forgot-password.component.html",
    styleUrls: ['password.component.css', '../../common.css']
})
export class ForgotPasswordComponent {
    opened: boolean = false;
    private onGoing: boolean = false;
    private email: string = "";
    private validationState: boolean = true;
    private isSuccess: boolean = false;

    @ViewChild("forgotPasswordFrom") forgotPwdForm: NgForm;
    @ViewChild(InlineAlertComponent)
    private inlineAlert: InlineAlertComponent;

    constructor(private pwdService: PasswordSettingService) { }

    public get showProgress(): boolean {
        return this.onGoing;
    }

    public get isValid(): boolean {
        return this.forgotPwdForm && this.forgotPwdForm.valid ;
    }

    public get btnCancelCaption(): string {
        if(this.isSuccess){
            return "BUTTON.CLOSE";
        }

        return "BUTTON.CANCEL";
    }

    public open(): void {
        //Clear state data
        this.validationState = true;
        this.isSuccess = false;
        this.onGoing = false;
        this.email = "";
        this.forgotPwdForm.resetForm();
        this.inlineAlert.close();

        this.opened = true;
    }

    public close(): void {
        this.opened = false;
    }

    public send(): void {
        //Double confirm to avoid improper situations
        if (!this.email) {
            return;
        }

        if (!this.isValid) {
            return;
        }

        this.onGoing = true;
        this.pwdService.sendResetPasswordMail(this.email)
            .then(response => {
                this.onGoing = false;
                this.isSuccess = true;
                this.inlineAlert.showInlineSuccess({
                    message: "RESET_PWD.SUCCESS"
                });
            })
            .catch(error => {
                this.onGoing = false;
                this.inlineAlert.showInlineError(error);
            })

    }

    public handleValidation(flag: boolean): void {
        if (flag) {
            this.validationState = true;
        } else {
            this.validationState = this.isValid;
        }
    }
}