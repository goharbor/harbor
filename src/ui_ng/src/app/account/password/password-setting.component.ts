import { Component, ViewChild, AfterViewChecked } from '@angular/core';
import { Router } from '@angular/router';
import { NgForm } from '@angular/forms';

import { PasswordSettingService } from './password-setting.service';
import { SessionService } from '../../shared/session.service';

@Component({
    selector: 'password-setting',
    templateUrl: "password-setting.component.html"
})
export class PasswordSettingComponent implements AfterViewChecked {
    opened: boolean = false;
    oldPwd: string = "";
    newPwd: string = "";
    reNewPwd: string = "";

    private formValueChanged: boolean = false;
    private onCalling: boolean = false;

    pwdFormRef: NgForm;
    @ViewChild("changepwdForm") pwdForm: NgForm;
    constructor(private passwordService: PasswordSettingService, private session: SessionService){}

    //If form is valid
    public get isValid(): boolean {
        if (this.pwdForm && this.pwdForm.form.get("newPassword")) {
            return this.pwdForm.valid &&
                this.pwdForm.form.get("newPassword").value === this.pwdForm.form.get("reNewPassword").value;
        }
        return false;
    }

    public get valueChanged(): boolean {
        return this.formValueChanged;
    }

    public get showProgress(): boolean {
        return this.onCalling;
    }

    ngAfterViewChecked() {
        if (this.pwdFormRef != this.pwdForm) {
            this.pwdFormRef = this.pwdForm;
            if (this.pwdFormRef) {
                this.pwdFormRef.valueChanges.subscribe(data => {
                    this.formValueChanged = true;
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
        if(!cUser){
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
        })
        .catch(error => {
            this.onCalling = false;
            console.error(error);//TODO:
        });
        //TODO:publish the successful message to general messae box
    }
}