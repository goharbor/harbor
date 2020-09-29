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
import {
    Component,
    ViewChild,
    AfterViewChecked,
    Output,
    EventEmitter,
    Input,
    OnInit,
    ChangeDetectorRef,
} from '@angular/core';
import { NgForm } from '@angular/forms';

import { User } from '../../user/user';
import { isEmptyForm } from '../../shared/shared.utils';
import { SessionService } from '../../shared/session.service';

@Component({
    selector: 'new-user-form',
    templateUrl: 'new-user-form.component.html',
    styleUrls: ['./new-user-form.component.scss', '../../common.scss']
})

export class NewUserFormComponent implements AfterViewChecked, OnInit {

    @Input() isSelfRegistration = false;
    // Notify the form value changes
    @Output() valueChange = new EventEmitter<boolean>();
    @ViewChild("newUserFrom", {static: true}) newUserForm: NgForm;
    newUser: User = new User();
    newUserFormRef: NgForm;
    confirmedPwd: string;
    timerHandler: any;
    validationStateMap: any = {};
    mailAlreadyChecked: any = {};
    userNameAlreadyChecked: any = {};
    emailTooltip = 'TOOLTIP.EMAIL';
    usernameTooltip = 'TOOLTIP.USER_NAME';
    formValueChanged = false;

    checkOnGoing: any = {};
    constructor(private session: SessionService,
        private ref: ChangeDetectorRef) { }

    ngOnInit() {
        this.resetState();
    }
    resetState(): void {
        this.mailAlreadyChecked = {};
        this.userNameAlreadyChecked = {};
        this.emailTooltip = 'TOOLTIP.EMAIL';
        this.usernameTooltip = 'TOOLTIP.USER_NAME';
        this.formValueChanged = false;
        this.checkOnGoing = {
            "username": false,
            "email": false
        };
        this.validationStateMap = {
            "username": true,
            "email": true,
            "realname": true,
            "newPassword": true,
            "confirmPassword": true,
            "comment": true
        };
    }

    public isChecking(key: string): boolean {
        return !this.checkOnGoing[key];
    }

    forceRefreshView(duration: number): void {
        // Reset timer
        if (this.timerHandler) {
          clearInterval(this.timerHandler);
        }
        this.timerHandler = setInterval(() => this.ref.markForCheck(), 100);
        setTimeout(() => {
          if (this.timerHandler) {
            clearInterval(this.timerHandler);
            this.timerHandler = null;
          }
        }, duration);
      }

    getValidationState(key: string): boolean {
        return !this.validationStateMap[key];
    }

    handleValidation(key: string, flag: boolean): void {
        if (flag) {
            // Checking
            let cont = this.newUserForm.controls[key];
            if (cont) {
                this.validationStateMap[key] = cont.valid;
                // Check email existing from backend
                if (cont.valid && this.formValueChanged) {
                    // Check username from backend
                    if (key === "username" && this.newUser.username.trim() !== "") {
                        if (this.userNameAlreadyChecked[this.newUser.username.trim()]) {
                            this.validationStateMap[key] = !this.userNameAlreadyChecked[this.newUser.username.trim()].result;
                            if (!this.validationStateMap[key]) {
                                this.usernameTooltip = "TOOLTIP.USER_EXISTING";
                            }
                            return;
                        }

                        this.checkOnGoing[key] = true;
                        this.session.checkUserExisting("username", this.newUser.username)
                            .subscribe((res: boolean) => {
                                this.checkOnGoing[key] = false;
                                this.validationStateMap[key] = !res;
                                if (res) {
                                    this.usernameTooltip = "TOOLTIP.USER_EXISTING";
                                }
                                this.userNameAlreadyChecked[this.newUser.username.trim()] = {
                                    result: res
                                }; // Tag it checked
                                this.forceRefreshView(2000);
                            }, error => {
                                this.checkOnGoing[key] = false;
                                this.validationStateMap[key] = false; // Not valid @ backend
                                this.forceRefreshView(2000);
                            });
                        return;

                    }

                    // Check email from backend
                    if (key === "email" && this.newUser.email.trim() !== "") {
                        if (this.mailAlreadyChecked[this.newUser.email.trim()]) {
                            this.validationStateMap[key] = !this.mailAlreadyChecked[this.newUser.email.trim()].result;
                            if (!this.validationStateMap[key]) {
                                this.emailTooltip = "TOOLTIP.EMAIL_EXISTING";
                            }
                            return;
                        }

                        // Mail changed
                        this.checkOnGoing[key] = true;
                        this.session.checkUserExisting("email", this.newUser.email)
                            .subscribe((res: boolean) => {
                                this.checkOnGoing[key] = false;
                                this.validationStateMap[key] = !res;
                                if (res) {
                                    this.emailTooltip = "TOOLTIP.EMAIL_EXISTING";
                                }
                                this.mailAlreadyChecked[this.newUser.email.trim()] = {
                                    result: res
                                }; // Tag it checked
                                this.forceRefreshView(2000);
                            }, error => {
                                this.checkOnGoing[key] = false;
                                this.validationStateMap[key] = false; // Not valid @ backend
                                this.forceRefreshView(2000);
                            });
                        return;
                    }

                    // Check password confirmation
                    if (key === "confirmPassword" || key === "newPassword") {
                        let cpKey = key === "confirmPassword" ? "newPassword" : "confirmPassword";
                        let peerCont = this.newUserForm.controls[cpKey];
                        if (peerCont && peerCont.valid) {
                            this.validationStateMap["confirmPassword"] = cont.value === peerCont.value;
                        }
                    }
                }
            }
        } else {
            // Reset
            this.validationStateMap[key] = true;
            if (key === "email") {
                this.emailTooltip = "TOOLTIP.EMAIL";
            }

            if (key === "username") {
                this.usernameTooltip = "TOOLTIP.USER_NAME";
            }
        }
    }

    public get isValid(): boolean {
        let pwdEqualStatus = true;
        if (this.newUserForm.controls["confirmPassword"] &&
            this.newUserForm.controls["newPassword"]) {
            pwdEqualStatus = this.newUserForm.controls["confirmPassword"].value === this.newUserForm.controls["newPassword"].value;
        }
        return this.newUserForm &&
            this.newUserForm.valid &&
            pwdEqualStatus &&
            this.validationStateMap["username"] &&
            this.validationStateMap["email"]; // Backend check should be valid as well
    }

    ngAfterViewChecked(): void {
        if (this.newUserFormRef !== this.newUserForm) {
            this.newUserFormRef = this.newUserForm;
            if (this.newUserFormRef) {
                this.newUserFormRef.valueChanges.subscribe(data => {
                    this.formValueChanged = true;
                    this.valueChange.emit(true);
                });
            }
        }
    }

    // Return the current user data
    getData(): User {
        return this.newUser;
    }

    // Reset form
    reset(): void {
        this.resetState();
        if (this.newUserForm) {
            this.newUserForm.reset();
        }
    }

    // To check if form is empty
    isEmpty(): boolean {
        return isEmptyForm(this.newUserForm);
    }
}
