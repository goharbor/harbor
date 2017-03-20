import { Component, ViewChild, OnInit } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { NgForm } from '@angular/forms';

import { PasswordSettingService } from './password-setting.service';
import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';
import { errorHandler, accessErrorHandler } from '../../shared/shared.utils';
import { AlertType } from '../../shared/shared.const';
import { MessageService } from '../../global-message/message.service';

@Component({
    selector: 'reset-password',
    templateUrl: "reset-password.component.html",
    styleUrls: ['password.component.css', '../../common.css']
})
export class ResetPasswordComponent implements OnInit{
    opened: boolean = true;
    private onGoing: boolean = false;
    private password: string = "";
    private validationState: any = {};
    private resetUuid: string = "";
    private resetOk: boolean = false;

    @ViewChild("resetPwdForm") resetPwdForm: NgForm;
    @ViewChild(InlineAlertComponent)
    private inlineAlert: InlineAlertComponent;

    constructor(
        private pwdService: PasswordSettingService,
        private route: ActivatedRoute,
        private msgService: MessageService,
        private router: Router) { }

    ngOnInit(): void {
        this.route.queryParams.subscribe(params => this.resetUuid = params["reset_uuid"] || "");
    }

    public get showProgress(): boolean {
        return this.onGoing;
    }

    public get isValid(): boolean {
        return this.resetPwdForm && this.resetPwdForm.valid && this.samePassword();
    }

    public getValidationState(key: string): boolean {
        return this.validationState && 
        this.validationState[key] &&
        key === 'reNewPassword'?this.samePassword():true;
    }

    public open(): void {
        this.resetOk = false;
        this.opened = true;
        this.resetPwdForm.resetForm();
    }

    public close(): void {
        this.opened = false;
    }

    public send(): void {
        //If already reset password ok, navigator to sign-in
        if(this.resetOk){
            this.router.navigate(['sign-in']);
            return;
        }

        //Double confirm to avoid improper situations
        if (!this.password) {
            return;
        }

        if (!this.isValid) {
            return;
        }

        this.onGoing = true;
        this.pwdService.resetPassword(this.resetUuid, this.password)
        .then(() => {
            this.onGoing = false;
            this.resetOk = true;
            this.inlineAlert.showInlineSuccess({message:'RESET_PWD.RESET_OK'});
        })
        .catch(error => {
            this.onGoing = false;
            if(accessErrorHandler(error, this.msgService)){
                this.close();
            }else{
                this.inlineAlert.showInlineError(errorHandler(error));
            }
        });
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