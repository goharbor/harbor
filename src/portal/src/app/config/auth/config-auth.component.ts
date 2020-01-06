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
import { Component, Input, ViewChild, SimpleChanges, OnChanges, OnInit, Output, EventEmitter } from '@angular/core';
import { NgForm } from '@angular/forms';
import { Subscription } from "rxjs";
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { ConfirmMessageHandler } from '../config.msg.utils';
import { AppConfigService } from '../../app-config.service';
import { ConfigurationService } from '../config.service';
import { Configuration } from "../../../lib/components/config/config";
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { SystemInfoService } from "../../../lib/services";
import { clone, isEmpty, getChanges as getChangesFunc } from "../../../lib/utils/utils";
import { CONFIG_AUTH_MODE } from "../../../lib/entities/shared.const";
const fakePass = 'aWpLOSYkIzJTTU4wMDkx';

@Component({
    selector: 'config-auth',
    templateUrl: 'config-auth.component.html',
    styleUrls: ['./config-auth.component.scss', '../config.component.scss']
})
export class ConfigurationAuthComponent implements OnChanges, OnInit {
    changeSub: Subscription;
    testingOnGoing = false;
    onGoing = false;
    redirectUrl: string;
    // tslint:disable-next-line:no-input-rename
    @Input('allConfig') currentConfig: Configuration = new Configuration();
    private originalConfig: Configuration;
    @ViewChild('authConfigFrom', {static: false}) authForm: NgForm;
    @Output() refreshAllconfig = new EventEmitter<any>();

    constructor(
        private msgHandler: MessageHandlerService,
        private configService: ConfigurationService,
        private appConfigService: AppConfigService,
        private confirmMessageHandler: ConfirmMessageHandler,
        private systemInfo: SystemInfoService,
        private errorHandler: ErrorHandler,
    ) {
    }
    ngOnInit() {
        this.getSystemInfo();
    }
    getSystemInfo(): void {
        this.systemInfo.getSystemInfo()
            .subscribe(systemInfo => (this.redirectUrl = systemInfo.external_url)
                , error => this.errorHandler.error(error));
    }
    get checkable() {
        return this.currentConfig &&
            this.currentConfig.self_registration &&
            this.currentConfig.self_registration.value === true;
    }

    ngOnChanges(changes: SimpleChanges): void {
        if (changes && changes["currentConfig"]) {
            this.originalConfig = clone(this.currentConfig);

        }
    }

    public get showLdap(): boolean {
        return this.currentConfig &&
            this.currentConfig.auth_mode &&
            this.currentConfig.auth_mode.value === CONFIG_AUTH_MODE.LDAP_AUTH;
    }

    public get showUAA(): boolean {
        return this.currentConfig && this.currentConfig.auth_mode && this.currentConfig.auth_mode.value === CONFIG_AUTH_MODE.UAA_AUTH;
    }
    public get showOIDC(): boolean {
        return this.currentConfig && this.currentConfig.auth_mode && this.currentConfig.auth_mode.value === CONFIG_AUTH_MODE.OIDC_AUTH;
    }
    public get showHttpAuth(): boolean {
        return this.currentConfig && this.currentConfig.auth_mode && this.currentConfig.auth_mode.value === CONFIG_AUTH_MODE.HTTP_AUTH;
    }
    public get showSelfReg(): boolean {
        if (!this.currentConfig || !this.currentConfig.auth_mode) {
            return true;
        } else {
            return this.currentConfig.auth_mode.value !== CONFIG_AUTH_MODE.LDAP_AUTH
                && this.currentConfig.auth_mode.value !== CONFIG_AUTH_MODE.UAA_AUTH
                && this.currentConfig.auth_mode.value !== CONFIG_AUTH_MODE.HTTP_AUTH
                && this.currentConfig.auth_mode.value !== CONFIG_AUTH_MODE.OIDC_AUTH;
        }
    }

    public isValid(): boolean {
        return this.authForm && this.authForm.valid;
    }

    public hasChanges(): boolean {
        return !isEmpty(this.getChanges());
    }

    setVerifyCertValue($event: any) {
        this.currentConfig.ldap_verify_cert.value = $event;
    }

    public pingTestServer(): void {
        if (this.testingOnGoing) {
            return; // Should not come here
        }

        let settings = {};
        if (this.currentConfig.auth_mode.value === CONFIG_AUTH_MODE.LDAP_AUTH) {

            for (let prop in this.currentConfig) {
                if (prop.startsWith('ldap_')) {
                    settings[prop] = this.currentConfig[prop].value;
                }
            }

            let allChanges = this.getChanges();
            this.testingOnGoing = true;
            // set password for ldap
            let ldapSearchPwd = allChanges['ldap_search_password'];
            if (ldapSearchPwd) {
                settings['ldap_search_password'] = ldapSearchPwd;
            } else {
                delete settings['ldap_search_password'];
            }

            // Fix: Confirm ldap scope is number
            settings['ldap_scope'] = +settings['ldap_scope'];

            this.configService.testLDAPServer(settings)
                .subscribe(respone => {
                    this.testingOnGoing = false;
                    this.msgHandler.showSuccess('CONFIG.TEST_LDAP_SUCCESS');
                }, error => {
                    this.testingOnGoing = false;
                    let err = error.error;
                    if (!err || !err.trim()) {
                        err = 'UNKNOWN';
                    }
                    this.msgHandler.showError('CONFIG.TEST_LDAP_FAILED', { 'param': err });
                });
        } else {
            for (let prop in this.currentConfig) {
                if (prop === 'oidc_endpoint') {
                    settings['url'] = this.currentConfig[prop].value;
                } else if (prop === 'oidc_verify_cert') {
                    settings['verify_cert'] = this.currentConfig[prop].value;
                }
            }
            this.testingOnGoing = true;
            this.configService.testOIDCServer(settings)
                .subscribe(respone => {
                    this.testingOnGoing = false;
                    this.msgHandler.showSuccess('CONFIG.TEST_OIDC_SUCCESS');
                }, error => {
                    this.testingOnGoing = false;
                    this.errorHandler.error(error);
                });
        }

    }

    public get showTestingServerBtn(): boolean {
        return this.currentConfig.auth_mode &&
            (this.currentConfig.auth_mode.value === CONFIG_AUTH_MODE.LDAP_AUTH
                || this.currentConfig.auth_mode.value === CONFIG_AUTH_MODE.OIDC_AUTH);
    }

    public isConfigValidForTesting(): boolean {
        return this.isValid() &&
            !this.testingOnGoing;
    }

    public getChanges() {
        let allChanges = getChangesFunc(this.originalConfig, this.currentConfig);
        let changes = {};
        for (let prop in allChanges) {
            if (prop.startsWith('ldap_')
                || prop.startsWith('uaa_')
                || prop.startsWith('oidc_')
                || prop === 'auth_mode'
                || prop === 'project_creattion_restriction'
                || prop === 'self_registration'
                || prop.startsWith('http_')
            ) {
                changes[prop] = allChanges[prop];
            }
        }
        return changes;
    }

    public get hideTestingSpinner(): boolean {
        return !this.testingOnGoing || !this.showTestingServerBtn;
    }

    disabled(prop: any): boolean {
        return !(prop && prop.editable);
    }

    handleOnChange($event: any): void {
        if ($event && $event.target && $event.target["value"]) {
            let authMode = $event.target["value"];
            if (authMode === CONFIG_AUTH_MODE.LDAP_AUTH || authMode === CONFIG_AUTH_MODE.UAA_AUTH || authMode === CONFIG_AUTH_MODE.HTTP_AUTH
                || authMode === CONFIG_AUTH_MODE.OIDC_AUTH) {
                if (this.currentConfig.self_registration.value) {
                    this.currentConfig.self_registration.value = false; // unselect
                }
            }
        }
    }

    /**
    *
    * Save the changed values
    *
    * @memberOf ConfigurationComponent
    */
    public save(): void {
        let changes = this.getChanges();
        if (!isEmpty(changes)) {
            this.onGoing = true;
            this.configService.saveConfiguration(changes)
                .subscribe(response => {
                    this.onGoing = false;
                    this.refreshAllconfig.emit();
                    // Reload bootstrap option
                    this.appConfigService.load().subscribe(() => { }
                        , error => console.error('Failed to reload bootstrap option with error: ', error));
                    this.msgHandler.showSuccess('CONFIG.SAVE_SUCCESS');
                }, error => {
                    this.onGoing = false;
                    this.msgHandler.handleError(error);
                });
        } else {
            // Inprop situation, should not come here
            console.error('Save abort because nothing changed');
        }
    }

    /**
     *
     * Discard current changes if have and reset
     *
     * @memberOf ConfigurationComponent
     */
    public cancel(): void {
        let changes = this.getChanges();
        if (!isEmpty(changes)) {
            this.confirmMessageHandler.confirmUnsavedChanges(changes);
        } else {
            // Invalid situation, should not come here
            console.error('Nothing changed');
        }
    }

}
