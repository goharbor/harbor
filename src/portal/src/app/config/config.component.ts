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
import {
    Configuration, StringValueItem, SystemSettingsComponent,
    isEmpty, clone } from '@harbor/ui';
import { ConfirmationTargets, ConfirmationState } from '../shared/shared.const';
import { SessionService } from '../shared/session.service';
import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';

import { AppConfigService } from '../app-config.service';
import { ConfigurationAuthComponent } from './auth/config-auth.component';
import { ConfigurationEmailComponent } from './email/config-email.component';
import { ConfigurationService } from './config.service';


const fakePass = 'aWpLOSYkIzJTTU4wMDkx';
const TabLinkContentMap = {
    'config-auth': 'authentication',
    'config-replication': 'replication',
    'config-email': 'email',
    'config-system': 'system_settings',
    'config-label': 'system_label',
};

@Component({
    selector: 'config',
    templateUrl: 'config.component.html',
    styleUrls: ['config.component.scss']
})
export class ConfigurationComponent implements OnInit, OnDestroy {
    allConfig: Configuration = new Configuration();
    onGoing = false;
    currentTabId = 'config-auth'; // default tab
    originalCopy: Configuration = new Configuration();
    confirmSub: Subscription;

    @ViewChild(SystemSettingsComponent, { static: false }) systemSettingsConfig: SystemSettingsComponent;
    @ViewChild(ConfigurationEmailComponent, { static: false }) mailConfig: ConfigurationEmailComponent;
    @ViewChild(ConfigurationAuthComponent, { static: false }) authConfig: ConfigurationAuthComponent;

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

    public get withAdmiral(): boolean {
        return this.appConfigService.getConfig().with_admiral;
    }

    isCurrentTabLink(tabId: string): boolean {
        return this.currentTabId === tabId;
    }

    isCurrentTabContent(contentId: string): boolean {
        return TabLinkContentMap[this.currentTabId] === contentId;
    }
    refreshAllconfig() {
        this.retrieveConfig();
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

    handleReadyOnlyChange(event) {
        this.msgHandler.handleReadOnly();
        if (!event) {
            this.msgHandler.clear();
        }
    }

    handleAppConfig(event) {
        // Reload bootstrap option
        this.appConfigService.load().subscribe(() => {}
        , error => console.error('Failed to reload bootstrap option with error: ', error));
    }

    public tabLinkClick(tabLink: string) {
        this.currentTabId = tabLink;
    }

    retrieveConfig(): void {
        this.onGoing = true;
        this.configService.getConfiguration()
            .subscribe((configurations: Configuration) => {
                this.onGoing = false;

                // Add two password fields
                configurations.email_password = new StringValueItem(fakePass, true);
                configurations.ldap_search_password = new StringValueItem(fakePass, true);
                configurations.uaa_client_secret = new StringValueItem(fakePass, true);
                configurations.oidc_client_secret = new StringValueItem(fakePass, true);
                this.allConfig = configurations;
                // Keep the original copy of the data
                this.originalCopy = clone(configurations);
            }, error => {
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
