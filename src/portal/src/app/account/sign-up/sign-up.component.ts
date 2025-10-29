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
import { Component, EventEmitter, Output, ViewChild } from '@angular/core';
import { Modal } from '../../shared/services';
import { NewUserFormComponent } from '../../shared/components/new-user-form/new-user-form.component';
import { User } from '../../base/left-side-nav/user/user';
import { SessionService } from '../../shared/services/session.service';
import { UserService } from '../../base/left-side-nav/user/user.service';
import { InlineAlertComponent } from '../../shared/components/inline-alert/inline-alert.component';

@Component({
    selector: 'sign-up',
    templateUrl: 'sign-up.component.html',
    styleUrls: ['../../common.scss'],
})
export class SignUpComponent {
    opened: boolean = false;
    staticBackdrop: boolean = true;
    error: any;
    onGoing: boolean = false;
    formValueChanged: boolean = false;

    @Output() userCreation = new EventEmitter<User>();

    constructor(
        private session: SessionService,
        private userService: UserService
    ) {}

    @ViewChild(NewUserFormComponent, { static: true })
    newUserForm: NewUserFormComponent;

    @ViewChild(InlineAlertComponent)
    inlineAlert: InlineAlertComponent;

    @ViewChild(Modal)
    modal: Modal;

    getNewUser(): User {
        return this.newUserForm.getData();
    }

    public get inProgress(): boolean {
        return this.onGoing;
    }

    public get isValid(): boolean {
        return this.newUserForm.isValid && this.error == null;
    }

    formValueChange(flag: boolean): void {
        if (flag) {
            this.formValueChanged = true;
        }
        if (this.error != null) {
            this.error = null; // clear error
        }
        this.inlineAlert.close(); // Close alert if being shown
    }

    open(): void {
        // Reset state
        this.newUserForm.reset();
        this.formValueChanged = false;
        this.error = null;
        this.onGoing = false;
        this.inlineAlert.close();

        this.modal.open();
    }

    close(): void {
        if (this.formValueChanged) {
            if (this.newUserForm.isEmpty()) {
                this.opened = false;
            } else {
                // Need user confirmation
                this.inlineAlert.showInlineConfirmation({
                    message: 'ALERT.FORM_CHANGE_CONFIRMATION',
                });
            }
        } else {
            this.opened = false;
        }
    }

    confirmCancel($event: any): void {
        this.opened = false;
        this.modal.close();
    }

    // Create new user
    create(): void {
        // Double confirm everything is ok
        // Form is valid
        if (!this.isValid) {
            return;
        }

        // We have new user data
        let u = this.getNewUser();
        if (!u) {
            return;
        }

        // Start process
        this.onGoing = true;

        this.userService.addUser(u).subscribe(
            () => {
                this.onGoing = false;
                this.opened = false;
                this.modal.close();
                this.userCreation.emit(u);
            },
            error => {
                this.onGoing = false;
                this.error = error;
                this.inlineAlert.showInlineError(error);
            }
        );
    }
}
