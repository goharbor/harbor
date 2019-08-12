// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { Component, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';

import { PasswordSettingService } from '../password-setting.service';
import { InlineAlertComponent } from '../../../shared/inline-alert/inline-alert.component';

@Component({
    selector: 'forgot-password',
    templateUrl: "forgot-password.component.html",
    styleUrls: ['./forgot-password.component.scss', '../password-setting.component.scss', '../../../common.scss']
})
export class ForgotPasswordComponent {
    opened: boolean = false;
    onGoing: boolean = false;
    email: string = "";
    validationState: boolean = true;
    isSuccess: boolean = false;

    @ViewChild("forgotPasswordFrom", {static: false}) forgotPwdForm: NgForm;
    @ViewChild(InlineAlertComponent, {static: false})
    inlineAlert: InlineAlertComponent;

    constructor(private pwdService: PasswordSettingService) { }

    public get showProgress(): boolean {
        return this.onGoing;
    }

    public get isValid(): boolean {
        return this.forgotPwdForm && this.forgotPwdForm.valid ;
    }

    public get btnCancelCaption(): string {
        if (this.isSuccess) {
            return "BUTTON.CLOSE";
        }

        return "BUTTON.CANCEL";
    }

    public open(): void {
        // Clear state data
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
        // Double confirm to avoid improper situations
        if (!this.email) {
            return;
        }

        if (!this.isValid) {
            return;
        }

        this.onGoing = true;
        this.pwdService.sendResetPasswordMail(this.email)
            .subscribe(response => {
                this.onGoing = false;
                this.isSuccess = true;
                this.inlineAlert.showInlineSuccess({
                    message: "RESET_PWD.SUCCESS"
                });
            }, error => {
                this.onGoing = false;
                this.inlineAlert.showInlineError(error);
            });

    }

    public handleValidation(flag: boolean): void {
        if (flag) {
            this.validationState = true;
        } else {
            this.validationState = this.isValid;
        }
    }
}
