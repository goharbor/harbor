// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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