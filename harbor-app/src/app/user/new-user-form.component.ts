import { Component, ViewChild, AfterViewChecked, Output, EventEmitter } from '@angular/core';
import { NgForm } from '@angular/forms';

import { User } from './user';

@Component({
    selector: 'new-user-form',
    templateUrl: 'new-user-form.component.html'
})

export class NewUserFormComponent implements AfterViewChecked {
    newUser: User = new User();
    confirmedPwd: string = "";

    newUserFormRef: NgForm;
    @ViewChild("newUserFrom") newUserForm: NgForm;

    //Notify the form value changes
    @Output() valueChange = new EventEmitter<boolean>();

    public get isValid(): boolean {
        let pwdEqualStatus = true;
        if(this.newUserForm.controls["confirmPassword"] && 
        this.newUserForm.controls["newPassword"]){
            pwdEqualStatus = this.newUserForm.controls["confirmPassword"].value === this.newUserForm.controls["newPassword"].value;
        }
        return this.newUserForm && 
        this.newUserForm.valid && pwdEqualStatus;
    }

    ngAfterViewChecked(): void {
        if (this.newUserFormRef != this.newUserForm) {
            this.newUserFormRef = this.newUserForm;
            if (this.newUserFormRef) {
                this.newUserFormRef.valueChanges.subscribe(data => {
                    this.valueChange.emit(true);
                });
            }
        }
    }

    //Return the current user data
    getData(): User {
        return this.newUser;
    }

    reset(): void {
        if(this.newUserForm){
            this.newUserForm.reset();
        }
    }
}