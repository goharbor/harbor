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
import { Component, ViewChild, Output, EventEmitter } from '@angular/core';

import { NewUserFormComponent } from '../shared/new-user-form/new-user-form.component';
import { User } from './user';

import { SessionService } from '../shared/session.service';
import { UserService } from './user.service';
import { InlineAlertComponent } from '../shared/inline-alert/inline-alert.component';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';

@Component({
    selector: "new-user-modal",
    templateUrl: "new-user-modal.component.html",
    styleUrls: ['../common.scss']
})

export class NewUserModalComponent {
    opened: boolean = false;
    error: any;
    onGoing: boolean = false;
    formValueChanged: boolean = false;

    @Output() addNew = new EventEmitter<User>();

    constructor(private session: SessionService,
        private userService: UserService,
        private msgHandler: MessageHandlerService) { }

    @ViewChild(NewUserFormComponent, {static: true})
    newUserForm: NewUserFormComponent;
    @ViewChild(InlineAlertComponent, {static: false})
    inlineAlert: InlineAlertComponent;

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
        if (this.error != null) {
            this.error = null; // clear error
        }

        this.formValueChanged = true;
        this.inlineAlert.close();
    }

    open(): void {
        this.newUserForm.reset(); // Reset form
        this.formValueChanged = false;
        this.onGoing = false;
        this.error = null;
        this.inlineAlert.close();

        this.opened = true;
    }

    close(): void {
        if (this.formValueChanged) {
            if (this.newUserForm.isEmpty()) {
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

    confirmCancel(event: boolean): void {
        this.opened = false;
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

        // Session is ok and role is matched
        let account = this.session.getCurrentUser();
        if (!account || !account.has_admin_role) {
            return;
        }

        // Start process
        this.onGoing = true;

        this.userService.addUser(u)
            .subscribe(() => {
                this.onGoing = false;
                // TODO:
                // As no response data returned, can not add it to list directly

                this.addNew.emit(u);
                this.opened = false;
                this.msgHandler.showSuccess("USER.SAVE_SUCCESS");
            }, error => {
                this.onGoing = false;
                this.error = error;
                if (this.msgHandler.isAppLevel(error)) {
                    this.msgHandler.handleError(error);
                    this.opened = false;
                } else {
                    this.inlineAlert.showInlineError(error);
                }
            });
    }
}
