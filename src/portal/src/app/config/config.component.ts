// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the 'License');
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an 'AS IS' BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { Component, OnInit, OnDestroy, ViewChild } from '@angular/core';
import { Subscription } from "rxjs";
import { Configuration, StringValueItem, SystemSettingsComponent, VulnerabilityConfigComponent} from '@harbor/ui';

import { ConfirmationTargets, ConfirmationState } from '../shared/shared.const';
import { SessionService } from '../shared/session.service';
import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../shared/confirmation-dialog/confirmation-message';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';

import { AppConfigService } from '../app-config.service';
import { ConfigurationAuthComponent } from './auth/config-auth.component';
import { ConfigurationEmailComponent } from './email/config-email.component';
import { GcComponent} from './gc/gc.component';
import { ConfigurationService } from './config.service';


const fakePass = 'aWpLOSYkIzJTTU4wMDkx';
const TabLinkContentMap = {
    'config-auth': 'authentication',
    'config-replication': 'replication',
    'config-email': 'email',
    'config-system': 'system_settings',
    'config-vulnerability': 'vulnerability',
    'config-gc': 'gc',
    'config-label': 'system_label',
};

@Component({
    selector: 'config',
    templateUrl: 'config.component.html',
    styleUrls: ['config.component.scss']
})
export class ConfigurationComponent implements OnInit, OnDestroy {
    onGoing = false;
    allConfig: Configuration = new Configuration();
    currentTabId = 'config-auth'; // default tab
    originalCopy: Configuration;
    confirmSub: Subscription;
    testingMailOnGoing = false;
    testingLDAPOnGoing = false;

    @ViewChild(SystemSettingsComponent) systemSettingsConfig: SystemSettingsComponent;
    @ViewChild(VulnerabilityConfigComponent) vulnerabilityConfig: VulnerabilityConfigComponent;
    @ViewChild(GcComponent) gcConfig: GcComponent;
    @ViewChild(ConfigurationEmailComponent) mailConfig: ConfigurationEmailComponent;
    @ViewChild(ConfigurationAuthComponent) authConfig: ConfigurationAuthComponent;

    constructor(
        private msgHandler: MessageHandlerService,
        private configService: ConfigurationService,
        private confirmService: ConfirmationDialogService,
        private appConfigService: AppConfigService,
        private session: SessionService) { }

    public get hasAdminRole(): boolean {
        return this.session.getCurrentUser() &&
            this.session.getCurrentUser().has_admin_role;
    }

    public get hasCAFile(): boolean {
        return this.appConfigService.getConfig().has_ca_root;
    }

    public get withClair(): boolean {
        return this.appConfigService.getConfig().with_clair;
    }

    public get withAdmiral(): boolean {
        return this.appConfigService.getConfig().with_admiral;
    }

    isCurrentTabLink(tabId: string): boolean {
        return this.currentTabId === tabId;
    }

    isCurrentTabContent(contentId: string): boolean {
        return TabLinkContentMap[this.currentTabId] === contentId;
    }

    hasUnsavedChangesOfCurrentTab(): any {
        let allChanges = this.getChanges();
        if (this.isEmpty(allChanges)) {
            return null;
        }

        let properties = [];
        switch (this.currentTabId) {
            case 'config-auth':
                for (let prop in allChanges) {
                    if (prop.startsWith('ldap_')) {
                        return allChanges;
                    }
                }
                properties = ['auth_mode', 'project_creation_restriction', 'self_registration'];
                break;
            case 'config-email':
                for (let prop in allChanges) {
                    if (prop.startsWith('email_')) {
                        return allChanges;
                    }
                }
                return null;
            case 'config-replication':
                properties = ['verify_remote_cert'];
                break;
            case 'config-system':
                properties = ['token_expiration'];
                break;
            case 'config-vulnerability':
                properties = ['scan_all_policy'];
                break;
            default:
                return null;
        }

        for (let prop in allChanges) {
            if (properties.indexOf(prop) !== -1) {
                return allChanges;
            }
        }

        return null;
    }

    ngOnInit(): void {
        // First load
        // Double confirm the current use has admin role
        let currentUser = this.session.getCurrentUser();
        if (currentUser && currentUser.has_admin_role) {
            this.retrieveConfig();
        }

        this.confirmSub = this.confirmService.confirmationConfirm$.subscribe(confirmation => {
            if (confirmation &&
                confirmation.state === ConfirmationState.CONFIRMED) {
                if (confirmation.source === ConfirmationTargets.CONFIG) {
                    this.reset(confirmation.data);
                } else if (confirmation.source === ConfirmationTargets.CONFIG_TAB) {
                    this.reset(confirmation.data['changes']);
                    this.currentTabId = confirmation.data['tabId'];
                }
            }
        });
    }

    ngOnDestroy(): void {
        if (this.confirmSub) {
            this.confirmSub.unsubscribe();
        }
    }

    public get inProgress(): boolean {
        return this.onGoing;
    }

    public isValid(): boolean {
        return this.systemSettingsConfig &&
            this.systemSettingsConfig.isValid &&
            this.mailConfig &&
            this.mailConfig.isValid() &&
            this.authConfig &&
            this.authConfig.isValid() &&
            this.isVulnerabiltyValid;
    }

    public get isVulnerabiltyValid(): boolean {
        return !this.appConfigService.getConfig().with_clair ||
            (this.vulnerabilityConfig &&
                this.vulnerabilityConfig.isValid);
    }

    public hasChanges(): boolean {
        return !this.isEmpty(this.getChanges());
    }

    public isMailConfigValid(): boolean {
        return this.mailConfig &&
            this.mailConfig.isValid() &&
            !this.testingMailOnGoing;
    }

    public get showTestServerBtn(): boolean {
        return this.currentTabId === 'config-email';
    }

    public get showLdapServerBtn(): boolean {
        return this.currentTabId === 'config-auth' &&
            this.allConfig.auth_mode &&
            this.allConfig.auth_mode.value === 'ldap_auth';
    }

    public get hideBtn(): boolean {
        return this.currentTabId === 'config-label' || this.currentTabId === 'config-gc' || this.currentTabId === 'config-vulnerability';
    }

    public get hideMailTestingSpinner(): boolean {
        return !this.testingMailOnGoing || !this.showTestServerBtn;
    }

    public get hideLDAPTestingSpinner(): boolean {
        return !this.testingLDAPOnGoing || !this.showLdapServerBtn;
    }

    public isLDAPConfigValid(): boolean {
        return this.authConfig &&
            this.authConfig.isValid() &&
            !this.testingLDAPOnGoing;
    }

    public tabLinkClick(tabLink: string) {
        // Whether has unsaved changes in current tab
        let changes = this.hasUnsavedChangesOfCurrentTab();
        if (!changes) {
            this.currentTabId = tabLink;
            return;
        }

        this.confirmUnsavedTabChanges(changes, tabLink);
    }

    /**
     *
     * Save the changed values
     *
     * @memberOf ConfigurationComponent
     */
    public save(): void {
        let changes = this.getChanges();
        if (!this.isEmpty(changes)) {
            // Fix policy parameters issue
            let scanningAllPolicy = changes['scan_all_policy'];
            if (scanningAllPolicy &&
                scanningAllPolicy.type !== 'daily' &&
                scanningAllPolicy.parameters) {
                delete (scanningAllPolicy.parameters);
            }

            this.onGoing = true;
            this.configService.saveConfiguration(changes)
                .then(response => {
                    this.onGoing = false;
                    // API should return the updated configurations here
                    // Unfortunately API does not do that
                    // To refresh the view, we can clone the original data copy
                    // or force refresh by calling service.
                    // HERE we choose force way
                    this.retrieveConfig();

                    if (changes['read_only']) {
                        this.msgHandler.handleReadOnly();
                    }

                    if (changes && changes['read_only'] === false) {
                        this.msgHandler.clear();
                    }

                    // Reload bootstrap option
                    this.appConfigService.load().catch(error => console.error('Failed to reload bootstrap option with error: ', error));

                    this.msgHandler.showSuccess('CONFIG.SAVE_SUCCESS');
                })
                .catch(error => {
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
        if (!this.isEmpty(changes)) {
            this.confirmUnsavedChanges(changes);
        } else {
            // Invalid situation, should not come here
            console.error('Nothing changed');
        }
    }

    /**
     *
     * Test the connection of specified mail server
     *
     *
     * @memberOf ConfigurationComponent
     */
    public testMailServer(): void {
        if (this.testingMailOnGoing) {
            return; // Should not come here
        }
        let mailSettings = {};
        for (let prop in this.allConfig) {
            if (prop.startsWith('email_')) {
                mailSettings[prop] = this.allConfig[prop].value;
            }
        }
        // Confirm port is number
        mailSettings['email_port'] = +mailSettings['email_port'];
        let allChanges = this.getChanges();
        let password = allChanges['email_password'];
        if (password) {
            mailSettings['email_password'] = password;
        } else {
            delete mailSettings['email_password'];
        }

        this.testingMailOnGoing = true;
        this.configService.testMailServer(mailSettings)
            .then(response => {
                this.testingMailOnGoing = false;
                this.msgHandler.showSuccess('CONFIG.TEST_MAIL_SUCCESS');
            })
            .catch(error => {
                this.testingMailOnGoing = false;
                let err = error._body;
                if (!err) {
                    err = 'UNKNOWN';
                }
                this.msgHandler.showError('CONFIG.TEST_MAIL_FAILED', { 'param': err });
            });
    }

    public testLDAPServer(): void {
        if (this.testingLDAPOnGoing) {
            return; // Should not come here
        }

        let ldapSettings = {};
        for (let prop in this.allConfig) {
            if (prop.startsWith('ldap_')) {
                ldapSettings[prop] = this.allConfig[prop].value;
            }
        }

        let allChanges = this.getChanges();
        let ldapSearchPwd = allChanges['ldap_search_password'];
        if (ldapSearchPwd) {
            ldapSettings['ldap_search_password'] = ldapSearchPwd;
        } else {
            delete ldapSettings['ldap_search_password'];
        }

        // Fix: Confirm ldap scope is number
        ldapSettings['ldap_scope'] = +ldapSettings['ldap_scope'];

        this.testingLDAPOnGoing = true;
        this.configService.testLDAPServer(ldapSettings)
            .then(respone => {
                this.testingLDAPOnGoing = false;
                this.msgHandler.showSuccess('CONFIG.TEST_LDAP_SUCCESS');
            })
            .catch(error => {
                this.testingLDAPOnGoing = false;
                let err = error._body;
                if (!err || !err.trim()) {
                    err = 'UNKNOWN';
                }
                this.msgHandler.showError('CONFIG.TEST_LDAP_FAILED', { 'param': err });
            });
    }

    confirmUnsavedChanges(changes: any) {
        let msg = new ConfirmationMessage(
            'CONFIG.CONFIRM_TITLE',
            'CONFIG.CONFIRM_SUMMARY',
            '',
            changes,
            ConfirmationTargets.CONFIG
        );

        this.confirmService.openComfirmDialog(msg);
    }

    confirmUnsavedTabChanges(changes: any, tabId: string) {
        let msg = new ConfirmationMessage(
            'CONFIG.CONFIRM_TITLE',
            'CONFIG.CONFIRM_SUMMARY',
            '',
            {
                'changes': changes,
                'tabId': tabId
            },
            ConfirmationTargets.CONFIG_TAB
        );

        this.confirmService.openComfirmDialog(msg);
    }

    retrieveConfig(): void {
        this.onGoing = true;
        this.configService.getConfiguration()
            .then((configurations: Configuration) => {
                this.onGoing = false;

                // Add two password fields
                configurations.email_password = new StringValueItem(fakePass, true);
                configurations.ldap_search_password = new StringValueItem(fakePass, true);
                configurations.uaa_client_secret = new StringValueItem(fakePass, true);
                this.allConfig = configurations;
                // Keep the original copy of the data
                this.originalCopy = this.clone(configurations);
            })
            .catch(error => {
                this.onGoing = false;
                this.msgHandler.handleError(error);
            });
    }

    /**
     *
     * Get the changed fields and return a map
     *
     * @private
     * returns {*}
     *
     * @memberOf ConfigurationComponent
     */
    getChanges(): { [key: string]: any | any[] } {
        let changes: { [key: string]: any | any[] } = {};
        if (!this.allConfig || !this.originalCopy) {
            return changes;
        }
        for (let prop of Object.keys(this.allConfig)) {
            let field = this.originalCopy[prop];
            if (field && field.editable) {
                if (!this.compareValue(field.value, this.allConfig[prop].value)) {
                    changes[prop] = this.allConfig[prop].value;
                    // Number
                    if (typeof field.value === 'number') {
                        changes[prop] = +changes[prop];
                    }

                    // Trim string value
                    if (typeof field.value === 'string') {
                        changes[prop] = ('' + changes[prop]).trim();
                    }
                }
            }
        }

        return changes;
    }

    // private
    compareValue(a: any, b: any): boolean {
        if ((a && !b) || (!a && b)) { return false; }
        if (!a && !b) { return true; }

        return JSON.stringify(a) === JSON.stringify(b);
    }

    // private
    isEmpty(obj: any): boolean {
        return !obj || JSON.stringify(obj) === '{}';
    }

    /**
     *
     * Deep clone the configuration object
     *
     * @private
     *  ** deprecated param {Configuration} src
     * returns {Configuration}
     *
     * @memberOf ConfigurationComponent
     */
    clone(src: Configuration): Configuration {
        if (!src) {
            return new Configuration(); // Empty
        }

        return JSON.parse(JSON.stringify(src));
    }

    /**
     *
     * Reset the configuration form
     *
     * @private
     *  ** deprecated param {*} changes
     *
     * @memberOf ConfigurationComponent
     */
    reset(changes: any): void {
        if (!this.isEmpty(changes)) {
            for (let prop in changes) {
                if (this.originalCopy[prop]) {
                    this.allConfig[prop] = this.clone(this.originalCopy[prop]);
                }
            }
        } else {
            // force reset
            this.retrieveConfig();
        }
    }

    disabled(prop: any): boolean {
        return !(prop && prop.editable);
    }
}
