import { Component, ViewChild, Output, EventEmitter } from '@angular/core';
import { NgForm } from '@angular/forms';

import { NewUserFormComponent } from '../shared/new-user-form/new-user-form.component';
import { User } from './user';

import { SessionService } from '../shared/session.service';
import { UserService } from './user.service';
import { errorHandler, accessErrorHandler } from '../shared/shared.utils';
import { MessageService } from '../global-message/message.service';
import { AlertType, httpStatusCode } from '../shared/shared.const';
import { InlineAlertComponent } from '../shared/inline-alert/inline-alert.component';

@Component({
    selector: "new-user-modal",
    templateUrl: "new-user-modal.component.html"
})

export class NewUserModalComponent {
    opened: boolean = false;
    private error: any;
    private onGoing: boolean = false;
    private formValueChanged: boolean = false;

    @Output() addNew = new EventEmitter<User>();

    constructor(private session: SessionService,
        private userService: UserService,
        private msgService: MessageService) { }

    @ViewChild(NewUserFormComponent)
    private newUserForm: NewUserFormComponent;
    @ViewChild(InlineAlertComponent)
    private inlineAlert: InlineAlertComponent;

    private getNewUser(): User {
        return this.newUserForm.getData();
    }

    public get inProgress(): boolean {
        return this.onGoing;
    }

    public get isValid(): boolean {
        return this.newUserForm.isValid && this.error == null;
    }

    public get errorMessage(): string {
        return errorHandler(this.error);
    }

    formValueChange(flag: boolean): void {
        if (this.error != null) {
            this.error = null;//clear error
        }

        this.formValueChanged = true;
        this.inlineAlert.close();
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

        //Session is ok and role is matched
        let account = this.session.getCurrentUser();
        if (!account || account.has_admin_role === 0) {
            return;
        }

        //Start process
        this.onGoing = true;

        this.userService.addUser(u)
            .then(() => {
                this.onGoing = false;
                //TODO:
                //As no response data returned, can not add it to list directly

                this.addNew.emit(u);
                this.opened = false;
                this.msgService.announceMessage(200, "USER.SAVE_SUCCESS", AlertType.SUCCESS);
            })
            .catch(error => {
                this.onGoing = false;
                this.error = error;
                if(accessErrorHandler(error, this.msgService)){
                    this.opened = false;
                }else{
                    this.inlineAlert.showInlineError(error);
                }
            });
    }
}