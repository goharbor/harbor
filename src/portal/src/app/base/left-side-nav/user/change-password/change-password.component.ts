// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { Component, ViewChild, ChangeDetectorRef } from '@angular/core';
import { NgForm } from '@angular/forms';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { UserService } from '../user.service';
import { TranslateService } from '@ngx-translate/core';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';

@Component({
    selector: 'change-password',
    templateUrl: 'change-password.component.html',
    styleUrls: ['./change-password.component.scss', '../../../../common.scss'],
})
export class ChangePasswordComponent {
    showNewPwd: boolean = false;
    showConfirmPwd: boolean = false;
    opened: boolean = false;
    onGoing: boolean = false;
    password: string = '';
    private validationState: any = {
        newPassword: true,
        reNewPassword: true,
    };
    confirmPwd: string = '';
    userId: number;

    @ViewChild('resetPwdForm') resetPwdForm: NgForm;
    @ViewChild(InlineAlertComponent)
    inlineAlert: InlineAlertComponent;

    constructor(
        private userService: UserService,
        private msgHandler: MessageHandlerService,
        private translateService: TranslateService,
        private ref: ChangeDetectorRef
    ) {}

    public get showProgress(): boolean {
        return this.onGoing;
    }

    public get isValid(): boolean {
        return (
            this.resetPwdForm && this.resetPwdForm.valid && this.samePassword()
        );
    }

    public getValidationState(key: string): boolean {
        return this.validationState && this.validationState[key];
    }

    confirmCancel(event: boolean): void {
        this.opened = false;
    }

    public open(userId: number): void {
        this.showConfirmPwd = false;
        this.showNewPwd = false;
        this.onGoing = false;
        this.validationState = {
            newPassword: true,
            reNewPassword: true,
        };
        this.resetPwdForm.resetForm();
        this.inlineAlert.close();
        this.userId = userId;
        this.opened = true;
    }

    public close(): void {
        if (this.password || this.confirmPwd) {
            // Need user confirmation
            this.inlineAlert.showInlineConfirmation({
                message: 'ALERT.FORM_CHANGE_CONFIRMATION',
            });
        } else {
            this.opened = false;
        }
    }

    public send(): void {
        // Double confirm to avoid improper situations
        if (!this.password || !this.confirmPwd) {
            return;
        }

        if (!this.isValid) {
            return;
        }

        this.onGoing = true;
        this.userService
            .changePassword(this.userId, this.password, this.confirmPwd)
            .subscribe(
                () => {
                    this.onGoing = false;
                    this.opened = false;
                    this.msgHandler.showSuccess('USER.RESET_OK');

                    let hnd = setInterval(() => this.ref.markForCheck(), 100);
                    setTimeout(() => clearInterval(hnd), 2000);
                },
                error => {
                    this.onGoing = false;
                    if (error.status === 400) {
                        this.translateService
                            .get('USER.EXISTING_PASSWORD')
                            .subscribe(res => {
                                this.inlineAlert.showInlineError(res);
                            });
                    } else {
                        this.inlineAlert.showInlineError(error);
                    }
                    let hnd = setInterval(() => this.ref.markForCheck(), 100);
                    setTimeout(() => clearInterval(hnd), 2000);
                }
            );
    }

    public handleValidation(key: string, flag: boolean): void {
        if (!flag) {
            this.validationState[key] = true;
        } else {
            this.validationState[key] = this.getControlValidationState(key);
            if (this.validationState[key]) {
                this.validationState['reNewPassword'] = this.samePassword();
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
            let control1 = this.resetPwdForm.controls['newPassword'];
            let control2 = this.resetPwdForm.controls['reNewPassword'];
            if (control1 && control2) {
                return control1.value === control2.value;
            }
        }
        return false;
    }
}
