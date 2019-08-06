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
import { Component, ViewChild, OnInit } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { NgForm } from '@angular/forms';

import { PasswordSettingService } from '../password-setting.service';
import { InlineAlertComponent } from '../../../shared/inline-alert/inline-alert.component';
import { MessageHandlerService } from '../../../shared/message-handler/message-handler.service';
import { CommonRoutes } from '@harbor/ui';

@Component({
    selector: 'reset-password',
    templateUrl: "reset-password.component.html",
    styleUrls: ['./reset-password.component.scss', '../password-setting.component.scss', '../../../common.scss']
})
export class ResetPasswordComponent implements OnInit {
    opened: boolean = true;
    onGoing: boolean = false;
    password: string = "";
    private validationState: any = {
        "newPassword": true,
        "reNewPassword": true
    };
    resetUuid: string = "";
    resetOk: boolean = false;
    confirmPwd: string = "";

    @ViewChild("resetPwdForm", { static: false }) resetPwdForm: NgForm;
    @ViewChild(InlineAlertComponent, { static: false })
    inlineAlert: InlineAlertComponent;

    constructor(
        private pwdService: PasswordSettingService,
        private route: ActivatedRoute,
        private msgHandler: MessageHandlerService,
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

    public get btnCancelCaption(): string {
        if (!this.resetOk) {
            return 'BUTTON.CANCEL';
        } else {
            return 'BUTTON.CLOSE';
        }
    }

    public getValidationState(key: string): boolean {
        return this.validationState &&
            this.validationState[key];
    }

    public open(): void {
        this.resetOk = false;
        this.onGoing = false;
        this.validationState = {
            "newPassword": true,
            "reNewPassword": true
        };
        this.resetPwdForm.resetForm();
        this.inlineAlert.close();

        this.opened = true;
    }

    public close(): void {
        // If already reset password ok, navigator to sign-in
        if (this.resetOk) {
            this.router.navigateByUrl(CommonRoutes.EMBEDDED_SIGN_IN);
        }
        this.opened = false;
    }

    public send(): void {
        // Double confirm to avoid improper situations
        if (!this.password) {
            return;
        }

        if (!this.isValid) {
            return;
        }

        this.onGoing = true;
        this.pwdService.resetPassword(this.resetUuid, this.password)
            .subscribe(() => {
                this.onGoing = false;
                this.resetOk = true;
                this.inlineAlert.showInlineSuccess({ message: 'RESET_PWD.RESET_OK' });
            }, error => {
                this.onGoing = false;
                if (this.msgHandler.isAppLevel(error)) {
                    this.close();
                } else {
                    this.inlineAlert.showInlineError(error);
                }
            });
    }

    public handleValidation(key: string, flag: boolean): void {
        if (!flag) {
            this.validationState[key] = true;
        } else {
            this.validationState[key] = this.getControlValidationState(key);
            if (this.validationState[key]) {
                this.validationState["reNewPassword"] = this.samePassword();
            }
        }
    }

    getControlValidationState(key: string): boolean {
        if (this.resetPwdForm) {
            let control = this.resetPwdForm.controls[key];
            if (control) {
                return control.valid;
            }
        }

        return false;
    }

    samePassword(): boolean {
        if (this.resetPwdForm) {
            let control1 = this.resetPwdForm.controls["newPassword"];
            let control2 = this.resetPwdForm.controls["reNewPassword"];
            if (control1 && control2) {
                return control1.value === control2.value;
            }
        }

        return false;
    }
}
