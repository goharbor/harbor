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
import { Configuration, StringValueItem, SystemSettingsComponent, VulnerabilityConfigComponent,
    isEmpty, clone, getChanges } from '@harbor/ui';

import { ConfirmationTargets, ConfirmationState } from '../shared/shared.const';
import { SessionService } from '../shared/session.service';
import { confirmUnsavedChanges} from './config.msg.utils';
import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../shared/confirmation-dialog/confirmation-message';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';

import { AppConfigService } from '../app-config.service';
import { ConfigurationAuthComponent } from './auth/config-auth.component';
import { ConfigurationEmailComponent } from './email/config-email.component';
import { GcComponent } from './gc/gc.component';
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
    originalCopy: Configuration = new Configuration();
    confirmSub: Subscription;

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

    hasUnsavedChangesOfCurrentTab(allChanges: any): boolean {
        if (isEmpty(allChanges)) {
            return false;
        }

        let properties = [];
        switch (this.currentTabId) {
            case 'config-auth':
                return this.authConfig.hasUnsavedChanges(allChanges);
            case 'config-email':
                return this.mailConfig.hasUnsavedChanges(allChanges);
            case 'config-replication':
                properties = ['verify_remote_cert'];
                break;
            case 'config-system':
                return this.systemSettingsConfig.hasUnsavedChanges(allChanges);
        }

        for (let prop in allChanges) {
            if (properties.indexOf(prop) !== -1) {
                return true;
            }
        }

        return false;
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
        return this.systemSettingsConfig.isValid;
    }

    public hasChanges(): boolean {
        return !isEmpty(this.getSystemChanges());
    }


    public tabLinkClick(tabLink: string) {
        let allChanges = getChanges(this.originalCopy, this.allConfig);
        // Whether has unsaved changes in current tab
        let hasChanges = this.hasUnsavedChangesOfCurrentTab(allChanges);
        if (!hasChanges) {
            this.currentTabId = tabLink;
            return;
        }

        this.confirmUnsavedTabChanges(allChanges, tabLink);
    }

    public getSystemChanges() {
        let allChanges = getChanges(this.originalCopy, this.allConfig);
        if (allChanges) {
            return this.systemSettingsConfig.getSystemChanges(allChanges);
        }
        return null;
    }

    /**
     *
     * Save the changed values
     *
     * @memberOf ConfigurationComponent
     */
    public save(): void {
        let changes = this.getSystemChanges();
        if (!isEmpty(changes)) {
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
        let changes = this.getSystemChanges();
        if (!isEmpty(changes)) {
            confirmUnsavedChanges(changes);
        } else {
            // Invalid situation, should not come here
            console.error('Nothing changed');
        }
    }

    public get hideBtn(): boolean {
        return this.currentTabId !== 'config-system';
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
                this.originalCopy = clone(configurations);
            })
            .catch(error => {
                this.onGoing = false;
                this.msgHandler.handleError(error);
            });
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
        if (!isEmpty(changes)) {
            for (let prop in changes) {
                if (this.originalCopy[prop]) {
                    this.allConfig[prop] = clone(this.originalCopy[prop]);
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
