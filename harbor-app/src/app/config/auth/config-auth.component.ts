import { Component, Input, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';
import { Subscription }   from 'rxjs/Subscription';

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
            this.currentConfig.auth_mode.value === 'ldap';
    }

    private disabled(prop: any): boolean {
        return !(prop && prop.editable);
    }

    public isValid(): boolean {
        return this.authForm && this.authForm.valid;
    }
}