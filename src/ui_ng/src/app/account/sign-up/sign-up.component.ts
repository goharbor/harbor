import { Component, Output, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';

import { NewUserFormComponent } from '../../shared/new-user-form/new-user-form.component';
import { User } from '../../user/user';

import { SessionService } from '../../shared/session.service';
import { UserService } from '../../user/user.service';
import { errorHandler } from '../../shared/shared.utils';
import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';

@Component({
    selector: 'sign-up',
    templateUrl: "sign-up.component.html"
})
export class SignUpComponent {
    opened: boolean = false;
    staticBackdrop: boolean = true;
    private error: any;
    private onGoing: boolean = false;
    private formValueChanged: boolean = false;

    constructor(
        private session: SessionService,
        private userService: UserService) { }

    @ViewChild(NewUserFormComponent)
    private newUserForm: NewUserFormComponent;

    @ViewChild(InlineAlertComponent)
    private inlienAlert: InlineAlertComponent;

    private getNewUser(): User {
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
            this.error = null;//clear error
        }
        this.inlienAlert.close();//Close alert if being shown
    }

    open(): void {
        this.newUserForm.reset();//Reset form
        this.formValueChanged = false;
        this.opened = true;
    }

    close(): void {
        if (this.formValueChanged) {
            if (this.newUserForm.isEmpty()) {
                this.opened = false;
            } else {
                //Need user confirmation
                this.inlienAlert.showInlineConfirmation({
                    message: "ALERT.FORM_CHANGE_CONFIRMATION"
                });
            }
        } else {
            this.opened = false;
        }
    }

    confirmCancel(): void {
        this.opened = false;
    }

    //Create new user
    create(): void {
        //Double confirm everything is ok
        //Form is valid
        if (!this.isValid) {
            return;
        }

        //We have new user data
        let u = this.getNewUser();
        if (!u) {
            return;
        }

        //Start process
        this.onGoing = true;

        this.userService.addUser(u)
            .then(() => {
                this.onGoing = false;
                this.close();
            })
            .catch(error => {
                this.onGoing = false;
                this.error = error;
                this.inlienAlert.showInlineError(error);
            });
    }
}