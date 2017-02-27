import { Component, ViewChild, Output,EventEmitter } from '@angular/core';
import { NgForm } from '@angular/forms';

import { NewUserFormComponent } from './new-user-form.component';
import { User } from './user';

import { SessionService } from '../shared/session.service';
import { UserService } from './user.service';

@Component({
    selector: "new-user-modal",
    templateUrl: "new-user-modal.component.html"
})

export class NewUserModalComponent {
    opened: boolean = false;
    alertClose: boolean = true;
    private error: any;
    private onGoing: boolean = false;

    @Output() addNew = new EventEmitter<User>();

    constructor(private session: SessionService,
        private userService: UserService) { }

    @ViewChild(NewUserFormComponent)
    private newUserForm: NewUserFormComponent;

    private getNewUser(): User {
        return this.newUserForm.getData();
    }

    public get inProgress(): boolean {
        return this.onGoing;
    }

    public get isValid(): boolean {
        return this.newUserForm.isValid;
    }

    public get errorMessage(): string {
        if (this.error) {
            if (this.error.message) {
                return this.error.message;
            } else {
                if (this.error._body) {
                    return this.error._body;
                }
            }
        }
        return "";
    }

    formValueChange(flag: boolean): void {
        if (!this.alertClose) {
            this.alertClose = true;//If alert is shown, then close it
        }
    }

    open(): void {
        this.opened = true;
    }

    close(): void {
        this.newUserForm.reset();//Reset form
        this.opened = false;
    }

    //Create new user
    create(): void {
        //Double confirm everything is ok
        //Form is valid
        if(!this.isValid){
            return;
        }

        //We have new user data
        let u = this.getNewUser();
        if(!u){
            return;
        }

        //Session is ok and role is matched
        let account = this.session.getCurrentUser();
        if(!account || account.has_admin_role === 0){
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
            this.close();
        })
        .catch(error => {
            this.onGoing = false;
            this.error = error;
        });
    }
}