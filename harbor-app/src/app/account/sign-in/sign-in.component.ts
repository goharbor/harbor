import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { Input, ViewChild, AfterViewChecked } from '@angular/core';
import { NgForm } from '@angular/forms';

import { SessionService } from '../../shared/session.service';
import { SignInCredential } from '../../shared/sign-in-credential';

import { SignUpComponent } from '../sign-up/sign-up.component';
import { harborRootRoute } from '../../shared/shared.const';
import { ForgotPasswordComponent } from '../password/forgot-password.component';

import { AppConfigService } from '../../app-config.service';
import { AppConfig } from '../../app-config';

//Define status flags for signing in states
export const signInStatusNormal = 0;
export const signInStatusOnGoing = 1;
export const signInStatusError = -1;

@Component({
    selector: 'sign-in',
    templateUrl: "sign-in.component.html",
    styleUrls: ['sign-in.component.css']
})

export class SignInComponent implements AfterViewChecked, OnInit {
    private redirectUrl: string = "";
    private appConfig: AppConfig = new AppConfig();
    //Form reference
    signInForm: NgForm;
    @ViewChild('signInForm') currentForm: NgForm;
    @ViewChild('signupDialog') signUpDialog: SignUpComponent;
    @ViewChild('forgotPwdDialog') forgotPwdDialog: ForgotPasswordComponent;

    //Status flag
    signInStatus: number = signInStatusNormal;

    //Initialize sign in credential
    @Input() signInCredential: SignInCredential = {
        principal: "",
        password: ""
    };

    constructor(
        private router: Router,
        private session: SessionService,
        private route: ActivatedRoute,
        private appConfigService: AppConfigService
    ) { }

    ngOnInit(): void {
        this.appConfig = this.appConfigService.getConfig();
        this.route.queryParams
            .subscribe(params => {
                this.redirectUrl = params["redirect_url"] || "";
                let isSignUp = params["sign_up"] || "";
                if (isSignUp != "") {
                    this.signUp();//Open sign up
                }
            });
    }

    //For template accessing
    public get isError(): boolean {
        return this.signInStatus === signInStatusError;
    }

    public get isOnGoing(): boolean {
        return this.signInStatus === signInStatusOnGoing;
    }

    //Validate the related fields
    public get isValid(): boolean {
        return this.currentForm.form.valid;
    }

    //Whether show the 'sign up' link
    public get selfSignUp(): boolean {
        return this.appConfig.auth_mode === 'db_auth'
            && this.appConfig.self_registration;
    }

    //General error handler
    private handleError(error) {
        //Set error status
        this.signInStatus = signInStatusError;

        let message = error.status ? error.status + ":" + error.statusText : error;
        console.error("An error occurred when signing in:", message);
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
        if (!this.isValid || this.isOnGoing) {
            return;
        }

        //Start signing in progress
        this.signInStatus = signInStatusOnGoing;

        //Call the service to send out the http request
        this.session.signIn(this.signInCredential)
            .then(() => {
                //Set status
                this.signInStatus = signInStatusNormal;

                //Redirect to the right route
                if (this.redirectUrl === "") {
                    //Routing to the default location
                    this.router.navigateByUrl(harborRootRoute);
                } else {
                    this.router.navigateByUrl(this.redirectUrl);
                }
            })
            .catch(error => {
                this.handleError(error);
            });
    }

    //Open sign up dialog
    signUp(): void {
        this.signUpDialog.open();
    }

    //Open forgot password dialog
    forgotPassword(): void {
        this.forgotPwdDialog.open();
    }
}