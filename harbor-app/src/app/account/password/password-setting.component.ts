import { Component, ViewChild, AfterViewChecked } from '@angular/core';
import { Router } from '@angular/router';
import { NgForm } from '@angular/forms';

import { PasswordSettingService } from './password-setting.service';
import { SessionService } from '../../shared/session.service';
import { AlertType } from '../../shared/shared.const';
import { MessageService } from '../../global-message/message.service';
import { errorHandler } from '../../shared/shared.utils';

@Component({
    selector: 'password-setting',
    templateUrl: "password-setting.component.html"
})
export class PasswordSettingComponent implements AfterViewChecked {
    opened: boolean = false;
    oldPwd: string = "";
    newPwd: string = "";
    reNewPwd: string = "";
    alertClose: boolean = true;
    error: any = null;

    private formValueChanged: boolean = false;
    private onCalling: boolean = false;

    pwdFormRef: NgForm;
    @ViewChild("changepwdForm") pwdForm: NgForm;
    constructor(
        private passwordService: PasswordSettingService,
        private session: SessionService,
        private msgService: MessageService) { }

    //If form is valid
    public get isValid(): boolean {
        if (this.pwdForm && this.pwdForm.form.get("newPassword")) {
            return this.pwdForm.valid &&
                (this.pwdForm.form.get("newPassword").value === this.pwdForm.form.get("reNewPassword").value) &&
                this.error === null;
        }
        return false;
    }

    public get valueChanged(): boolean {
        return this.formValueChanged;
    }

    public get showProgress(): boolean {
        return this.onCalling;
    }

    public get errorMessage(): string {
        return errorHandler(this.error);
    }

    ngAfterViewChecked() {
        if (this.pwdFormRef != this.pwdForm) {
            this.pwdFormRef = this.pwdForm;
            if (this.pwdFormRef) {
                this.pwdFormRef.valueChanges.subscribe(data => {
                    this.formValueChanged = true;
                    this.error = null;
                    this.alertClose = true;
                });
            }
        }
    }

    //Open modal dialog
    open(): void {
        this.opened = true;
        this.pwdForm.reset();
    }

    //Close the moal dialog
    close(): void {
        this.opened = false;
    }

    //handle the ok action
    doOk(): void {
        if (this.onCalling) {
            return;//To avoid duplicate click events
        }

        if (!this.isValid) {
            return;//Double confirm
        }

        //Double confirm session is valid
        let cUser = this.session.getCurrentUser();
        if (!cUser) {
            return;
        }

        //Call service
        this.onCalling = true;

        this.passwordService.changePassword(cUser.user_id,
            {
                new_password: this.pwdForm.value.newPassword,
                old_password: this.pwdForm.value.oldPassword
            })
            .then(() => {
                this.onCalling = false;
                this.close();
                this.msgService.announceMessage(200, "CHANGE_PWD.SAVE_SUCCESS", AlertType.SUCCESS);
            })
            .catch(error => {
                this.onCalling = false;
                this.error = error;
                this.alertClose = false;
            });
    }
}