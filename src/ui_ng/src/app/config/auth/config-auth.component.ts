import { Component, Input, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';
import { Subscription } from 'rxjs/Subscription';

import { Configuration } from '../config';

@Component({
    selector: 'config-auth',
    templateUrl: "config-auth.component.html",
    styleUrls: ['../config.component.css']
})
export class ConfigurationAuthComponent {
    private changeSub: Subscription;
    @Input("ldapConfig") currentConfig: Configuration = new Configuration();

    @ViewChild("authConfigFrom") authForm: NgForm;

    constructor() { }

    public get showLdap(): boolean {
        return this.currentConfig &&
            this.currentConfig.auth_mode &&
            this.currentConfig.auth_mode.value === 'ldap_auth';
    }

    public get showSelfReg(): boolean {
        if (!this.currentConfig || !this.currentConfig.auth_mode) {
            return true;
        } else {
            return this.currentConfig.auth_mode.value != 'ldap_auth';
        }
    }

    private disabled(prop: any): boolean {
        return !(prop && prop.editable);
    }

    public isValid(): boolean {
        return this.authForm && this.authForm.valid;
    }

    private handleOnChange($event): void {
        if ($event && $event.target && $event.target["value"]) {
            let authMode = $event.target["value"];
            if (authMode === 'ldap_auth') {
                if (this.currentConfig.self_registration.value) {
                    this.currentConfig.self_registration.value = false;//uncheck
                }
            }
        }
    }
}