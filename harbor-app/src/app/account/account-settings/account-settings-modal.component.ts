import { Component, OnInit, ViewChild, AfterViewChecked } from '@angular/core';
import { NgForm } from '@angular/forms';

import { SessionUser } from '../../shared/session-user';
import { SessionService } from '../../shared/session.service';

@Component({
    selector: "account-settings-modal",
    templateUrl: "account-settings-modal.component.html"
})

export class AccountSettingsModalComponent implements OnInit, AfterViewChecked {
    opened: boolean = false;
    staticBackdrop: boolean = true;
    account: SessionUser;
    error: any;

    private isOnCalling: boolean = false;
    private formValueChanged: boolean = false;

    accountFormRef: NgForm;
    @ViewChild("accountSettingsFrom") accountForm: NgForm;

    constructor(private session: SessionService) { }

    ngOnInit(): void {
        //Value copy
        this.account = Object.assign({}, this.session.getCurrentUser());
    }

    public get isValid(): boolean {
        return this.accountForm && this.accountForm.valid;
    }

    public get showProgress(): boolean {
        return this.isOnCalling;
    }

    public get errorMessage(): string {
        return this.error ? (this.error.message ? this.error.message : this.error) : "";
    }

    ngAfterViewChecked(): void {
        if (this.accountFormRef != this.accountForm) {
            this.accountFormRef = this.accountForm;
            if (this.accountFormRef) {
                this.accountFormRef.valueChanges.subscribe(data => {
                    if (this.error) {
                        this.error = null;
                    }
                    this.formValueChanged = true;
                });
            }
        }
    }

    open() {
        this.account = Object.assign({}, this.session.getCurrentUser());
        this.formValueChanged = false;

        this.opened = true;
    }

    close() {
        this.opened = false;
    }

    submit() {
        if (!this.isValid || this.isOnCalling) {
            return;
        }

        //Double confirm session is valid
        let cUser = this.session.getCurrentUser();
        if (!cUser) {
            return;
        }

        this.isOnCalling = true;

        this.session.updateAccountSettings(this.account)
            .then(() => {
                this.isOnCalling = false;
                this.close();
            })
            .catch(error => {
                this.isOnCalling = false;
                this.error = error
            });
    }

}