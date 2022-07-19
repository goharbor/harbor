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
import { Component, OnInit, ViewChild } from '@angular/core';
import { Router } from '@angular/router';

import { NewUserFormComponent } from '../../shared/components/new-user-form/new-user-form.component';
import { User } from '../../base/left-side-nav/user/user';
import { UserService } from '../../base/left-side-nav/user/user.service';
import { MessageService } from '../../shared/components/global-message/message.service';
import { AlertType } from '../../shared/entities/shared.const';

@Component({
    selector: 'sign-up-page',
    templateUrl: 'sign-up-page.component.html',
    styleUrls: ['../../common.scss'],
})
export class SignUpPageComponent implements OnInit {
    error: any;
    onGoing: boolean = false;
    formValueChanged: boolean = false;

    constructor(
        private userService: UserService,
        private msgService: MessageService,
        private router: Router
    ) {}

    @ViewChild(NewUserFormComponent)
    newUserForm: NewUserFormComponent;

    getNewUser(): User {
        return this.newUserForm.getData();
    }

    public get inProgress(): boolean {
        return this.onGoing;
    }

    public get isValid(): boolean {
        return this.newUserForm.isValid && this.error == null;
    }

    public get canBeCancelled(): boolean {
        return (
            this.formValueChanged &&
            this.newUserForm &&
            !this.newUserForm.isEmpty()
        );
    }

    ngOnInit(): void {
        this.newUserForm.reset(); // Reset form
        this.formValueChanged = false;
    }

    formValueChange(flag: boolean): void {
        if (flag) {
            this.formValueChanged = true;
        }
        if (this.error != null) {
            this.error = null; // clear error
        }
    }

    cancel(): void {
        if (this.newUserForm) {
            this.newUserForm.reset();
        }
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
                this.msgService.announceMessage(200, '', AlertType.SUCCESS);
                // Navigate to embeded sign-in
                this.router.navigate(['harbor', 'sign-in']);
            },
            error => {
                this.onGoing = false;
                this.error = error;
                this.msgService.announceMessage(
                    error.status || 500,
                    '',
                    AlertType.WARNING
                );
            }
        );
    }
}
