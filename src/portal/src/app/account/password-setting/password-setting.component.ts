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
import { Component, ViewChild, AfterViewChecked } from '@angular/core';
import { NgForm } from '@angular/forms';

import { PasswordSettingService } from './password-setting.service';
import { SessionService } from '../../shared/session.service';
import { isEmptyForm } from '../../shared/shared.utils';
import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';

@Component({
    selector: 'password-setting',
    templateUrl: "password-setting.component.html",
    styleUrls: ['./password-setting.component.scss', '../../common.scss']
})
export class PasswordSettingComponent implements AfterViewChecked {
    opened: boolean = false;
    oldPwd: string = "";
    newPwd: string = "";
    reNewPwd: string = "";
    error: any = null;

    formValueChanged: boolean = false;
    onCalling: boolean = false;
    private validationStateMap: any = {
        "newPassword": true,
        "reNewPassword": true
    };

    pwdFormRef: NgForm;
    @ViewChild("changepwdForm", { static: false }) pwdForm: NgForm;
    @ViewChild(InlineAlertComponent, { static: false })
    inlineAlert: InlineAlertComponent;

    constructor(
        private passwordService: PasswordSettingService,
        private session: SessionService,
        private msgHandler: MessageHandlerService) { }

    // If form is valid
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

    getValidationState(key: string): boolean {
        return this.validationStateMap[key];
    }

    handleValidation(key: string, flag: boolean): void {
        if (flag) {
            // Checking
            let cont = this.pwdForm.controls[key];
            if (cont) {
                this.validationStateMap[key] = cont.valid;
                if (cont.valid) {
                    if (key === "reNewPassword" || key === "newPassword") {
                        let cpKey = key === "reNewPassword" ? "newPassword" : "reNewPassword";
                        let compareCont = this.pwdForm.controls[cpKey];
                        if (compareCont && compareCont.valid) {
                            this.validationStateMap["reNewPassword"] = cont.value === compareCont.value;
                        }
                    }
                }
            }
        } else {
            // Reset
            this.validationStateMap[key] = true;
        }
    }

    ngAfterViewChecked() {
        if (this.pwdFormRef !== this.pwdForm) {
            this.pwdFormRef = this.pwdForm;
            if (this.pwdFormRef) {
                this.pwdFormRef.valueChanges.subscribe(data => {
                    this.formValueChanged = true;
                    this.error = null;
                    this.inlineAlert.close();
                });
            }
        }
    }

    // Open modal dialog
    open(): void {
        // Reset state
        this.formValueChanged = false;
        this.onCalling = false;
        this.error = null;
        this.validationStateMap = {
            "newPassword": true,
            "reNewPassword": true
        };
        this.pwdForm.reset();
        this.inlineAlert.close();

        this.opened = true;
    }

    // Close the modal dialog
    close(): void {
        if (this.formValueChanged) {
            if (isEmptyForm(this.pwdForm)) {
                this.opened = false;
            } else {
                // Need user confirmation
                this.inlineAlert.showInlineConfirmation({
                    message: "ALERT.FORM_CHANGE_CONFIRMATION"
                });
            }
        } else {
            this.opened = false;
        }
    }

    confirmCancel($event: any): void {
        this.opened = false;
    }

    // handle the ok action
    doOk(): void {
        if (this.onCalling) {
            return; // To avoid duplicate click events
        }

        if (!this.isValid) {
            return; // Double confirm
        }

        // Double confirm session is valid
        let cUser = this.session.getCurrentUser();
        if (!cUser) {
            return;
        }

        // Call service
        this.onCalling = true;

        this.passwordService.changePassword(cUser.user_id,
            {
                new_password: this.pwdForm.value.newPassword,
                old_password: this.pwdForm.value.oldPassword
            })
            .subscribe(() => {
                this.onCalling = false;
                this.opened = false;
                this.msgHandler.showSuccess("CHANGE_PWD.SAVE_SUCCESS");
            }, error => {
                this.onCalling = false;
                this.error = error;
                if (this.msgHandler.isAppLevel(error)) {
                    this.opened = false;
                    this.msgHandler.handleError(error);
                } else {
                    // Special case for 400
                    let msg = '' + error.error;
                    if (msg && msg.includes('old_password_is_not_correct')) {
                        this.inlineAlert.showInlineError("INCONRRECT_OLD_PWD");
                    } else {
                        this.inlineAlert.showInlineError(error);
                    }
                }
            });
    }
}
