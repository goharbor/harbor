import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { Input, ViewChild, AfterViewChecked } from '@angular/core';
import { NgForm } from '@angular/forms';

import { SignInService } from './sign-in.service';
import { SignInCredential } from './sign-in-credential'

//Define status flags for signing in states
export const signInStatusNormal = 0;
export const signInStatusOnGoing = 1;
export const signInStatusError = -1;

@Component({
    selector: 'sign-in',
    templateUrl: "sign-in.component.html",
    styleUrls: ['sign-in.component.css'],

    providers: [SignInService]
})

export class SignInComponent implements AfterViewChecked {
    //Form reference
    signInForm: NgForm;
    @ViewChild('signInForm') currentForm: NgForm;

    //Status flag
    signInStatus: number = 0;

    //Initialize sign in credential
    @Input() signInCredential: SignInCredential = {
        principal: "",
        password: ""
    };

    constructor(
        private signInService: SignInService,
        private router: Router
    ) { }

    //For template accessing
    get statusError(): number {
        return signInStatusError;
    }

    get statusOnGoing(): number {
        return signInStatusOnGoing;
    }

    //Validate the related fields
    private validate(): boolean {
        return true;
        //return this.signInForm.valid;
    }

    //Hande form values changes
    private formChanged() {
        if (this.currentForm === this.signInForm) {
            return;
        }
        this.signInForm = this.currentForm;
        if (this.signInForm) {
            this.signInForm.valueChanges
                .subscribe(data => {
                    this.updateState();
                });
        }

    }

    //Implement interface
    //Watch the view change only when view is in error state
    ngAfterViewChecked() {
        if (this.signInStatus === signInStatusError) {
            this.formChanged();
        }
    }

    //Update the status if we have done some changes
    updateState(): void {
        if (this.signInStatus === signInStatusError) {
            this.signInStatus = signInStatusNormal; //reset
        }
    }

    //Trigger the signin action
    signIn(): void {
        //Should validate input firstly
        if (!this.validate()) {
            console.info("return");
            return;
        }

        //Start signing in progress
        this.signInStatus = signInStatusOnGoing;

        //Call the service to send out the http request
        this.signInService.signIn(this.signInCredential)
            .then(() => {
                //Set status
                this.signInStatus = signInStatusNormal;

                //Routing to the right location
                let nextRoute = ["/harbor", "dashboard"];
                this.router.navigate(nextRoute);
            })
            .catch(error => {
                //Set error status
                this.signInStatus = signInStatusError;

                let message = error.status ? error.status + ":" + error.statusText : error;
                console.error("An error occurred when signing in:", message);
            });
    }

    //Help user navigate to the sign up
    signUp(): void {
        let nextRoute = ["/harbor", "signup"];
        this.router.navigate(nextRoute);
    }
}