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
import { Component, ViewChild, OnInit } from '@angular/core';
import { NgForm } from '@angular/forms';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { AppConfigService } from '../../../../services/app-config.service';
import { ConfigurationService } from '../../../../services/config.service';
import { SystemInfoService } from '../../../../shared/services';
import {
    isEmpty,
    getChanges as getChangesFunc,
} from '../../../../shared/units/utils';
import { CONFIG_AUTH_MODE } from '../../../../shared/entities/shared.const';
import { errorHandler } from '../../../../shared/units/shared.utils';
import { Configuration } from '../config';
import { finalize } from 'rxjs/operators';
import { ConfigService } from '../config.service';

@Component({
    selector: 'config-auth',
    templateUrl: 'config-auth.component.html',
    styleUrls: ['./config-auth.component.scss', '../config.component.scss'],
})
export class ConfigurationAuthComponent implements OnInit {
    testingOnGoing = false;
    onGoing = false;
    redirectUrl: string;
    @ViewChild('authConfigFrom') authForm: NgForm;

    get currentConfig(): Configuration {
        return this.conf.getConfig();
    }

    set currentConfig(c: Configuration) {
        this.conf.setConfig(c);
    }

    constructor(
        private msgHandler: MessageHandlerService,
        private configService: ConfigurationService,
        private appConfigService: AppConfigService,
        private conf: ConfigService,
        private systemInfo: SystemInfoService
    ) {}
    ngOnInit() {
        this.conf.resetConfig();
        this.getSystemInfo();
    }
    getSystemInfo(): void {
        this.systemInfo.getSystemInfo().subscribe(
            systemInfo => (this.redirectUrl = systemInfo.external_url),
            error => this.msgHandler.error(error)
        );
    }
    get checkable() {
        return (
            this.currentConfig &&
            this.currentConfig.self_registration &&
            this.currentConfig.self_registration.value === true
        );
    }
    public get showLdap(): boolean {
        return (
            this.currentConfig &&
            this.currentConfig.auth_mode &&
            this.currentConfig.auth_mode.value === CONFIG_AUTH_MODE.LDAP_AUTH
        );
    }

    public get showUAA(): boolean {
        return (
            this.currentConfig &&
            this.currentConfig.auth_mode &&
            this.currentConfig.auth_mode.value === CONFIG_AUTH_MODE.UAA_AUTH
        );
    }
    public get showOIDC(): boolean {
        return (
            this.currentConfig &&
            this.currentConfig.auth_mode &&
            this.currentConfig.auth_mode.value === CONFIG_AUTH_MODE.OIDC_AUTH
        );
    }
    public get showHttpAuth(): boolean {
        return (
            this.currentConfig &&
            this.currentConfig.auth_mode &&
            this.currentConfig.auth_mode.value === CONFIG_AUTH_MODE.HTTP_AUTH
        );
    }
    public get showSelfReg(): boolean {
        if (!this.currentConfig || !this.currentConfig.auth_mode) {
            return true;
        } else {
            return (
                this.currentConfig.auth_mode.value !==
                    CONFIG_AUTH_MODE.LDAP_AUTH &&
                this.currentConfig.auth_mode.value !==
                    CONFIG_AUTH_MODE.UAA_AUTH &&
                this.currentConfig.auth_mode.value !==
                    CONFIG_AUTH_MODE.HTTP_AUTH &&
                this.currentConfig.auth_mode.value !==
                    CONFIG_AUTH_MODE.OIDC_AUTH
            );
        }
    }

    isValid(): boolean {
        return this.authForm && this.authForm.valid;
    }

    inProcess(): boolean {
        return this.onGoing || this.conf.getLoadingConfigStatus();
    }

    hasChanges(): boolean {
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

            this.configService
                .testLDAPServer(settings)
                .pipe(finalize(() => (this.testingOnGoing = false)))
                .subscribe(
                    res => {
                        if (res && res.success) {
                            this.msgHandler.showSuccess(
                                'CONFIG.TEST_LDAP_SUCCESS'
                            );
                        } else if (res && res.message) {
                            this.msgHandler.showError(
                                'CONFIG.TEST_LDAP_FAILED',
                                { param: res.message }
                            );
                        }
                    },
                    error => {
                        let err = errorHandler(error);
                        if (!err || !err.trim()) {
                            err = 'UNKNOWN';
                        }
                        this.msgHandler.showError('CONFIG.TEST_LDAP_FAILED', {
                            param: err,
                        });
                    }
                );
        } else {
            for (let prop in this.currentConfig) {
                if (prop === 'oidc_endpoint') {
                    settings['url'] = this.currentConfig[prop].value;
                } else if (prop === 'oidc_verify_cert') {
                    settings['verify_cert'] = this.currentConfig[prop].value;
                }
            }
            this.testingOnGoing = true;
            this.configService.testOIDCServer(settings).subscribe(
                respone => {
                    this.testingOnGoing = false;
                    this.msgHandler.showSuccess('CONFIG.TEST_OIDC_SUCCESS');
                },
                error => {
                    this.testingOnGoing = false;
                    this.msgHandler.error(error);
                }
            );
        }
    }

    public get showTestingServerBtn(): boolean {
        return (
            this.currentConfig.auth_mode &&
            (this.currentConfig.auth_mode.value ===
                CONFIG_AUTH_MODE.LDAP_AUTH ||
                this.currentConfig.auth_mode.value ===
                    CONFIG_AUTH_MODE.OIDC_AUTH)
        );
    }

    public isConfigValidForTesting(): boolean {
        if (!this.authForm || !this.currentConfig) {
            return true;
        }
        return this.isValid() && !this.testingOnGoing && !this.inProcess();
    }

    public getChanges() {
        let allChanges = getChangesFunc(
            this.conf.getOriginalConfig(),
            this.currentConfig
        );
        let changes = {};
        for (let prop in allChanges) {
            if (
                prop.startsWith('ldap_') ||
                prop.startsWith('uaa_') ||
                prop.startsWith('oidc_') ||
                prop === 'auth_mode' ||
                prop === 'project_creattion_restriction' ||
                prop === 'self_registration' ||
                prop.startsWith('http_')
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
        if ($event && $event.target && $event.target['value']) {
            let authMode = $event.target['value'];
            if (
                authMode === CONFIG_AUTH_MODE.LDAP_AUTH ||
                authMode === CONFIG_AUTH_MODE.UAA_AUTH ||
                authMode === CONFIG_AUTH_MODE.HTTP_AUTH ||
                authMode === CONFIG_AUTH_MODE.OIDC_AUTH
            ) {
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
            this.configService.saveConfiguration(changes).subscribe(
                response => {
                    this.onGoing = false;
                    this.conf.updateConfig();
                    // Reload bootstrap option
                    this.appConfigService.load().subscribe(
                        () => {},
                        error =>
                            console.error(
                                'Failed to reload bootstrap option with error: ',
                                error
                            )
                    );
                    this.msgHandler.showSuccess('CONFIG.SAVE_SUCCESS');
                },
                error => {
                    this.onGoing = false;
                    this.msgHandler.handleError(error);
                }
            );
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
            this.conf.confirmUnsavedChanges(changes);
        } else {
            // Invalid situation, should not come here
            console.error('Nothing changed');
        }
    }
    changeAutoOnBoard() {
        if (!this.currentConfig.oidc_auto_onboard.value) {
            this.currentConfig.oidc_user_claim.value = '';
        }
    }
    trimSpace(e: any) {
        if (e && e.target) {
            if (e.target.value) {
                e.target.value = e.target.value.trim();
            } else {
                e.target.value = '';
            }
        }
    }
}
