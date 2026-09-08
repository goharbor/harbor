// Copyright Project Harbor Authors
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
import { Injectable } from '@angular/core';
import { ConfirmationDialogService } from '../../global-confirmation-dialog/confirmation-dialog.service';
import {
    ConfirmationState,
    ConfirmationTargets,
} from '../../../shared/entities/shared.const';
import { ConfirmationMessage } from '../../global-confirmation-dialog/confirmation-message';
import { Configuration, StringValueItem } from './config';
import { ConfigureService } from 'ng-swagger-gen/services/configure.service';
import { clone } from '../../../shared/units/utils';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { finalize } from 'rxjs/operators';
import { Observable, Subscription } from 'rxjs';
import {
    EventService,
    HarborEvent,
} from '../../../services/event-service/event.service';

const fakePass = 'aWpLOSYkIzJTTU4wMDkx';

@Injectable()
export class ConfigService {
    private _loadingConfig: boolean = false;
    private _confirmSub: Subscription;
    private _currentConfig: Configuration = new Configuration();
    private _originalConfig: Configuration;

    constructor(
        private confirmService: ConfirmationDialogService,
        private configureService: ConfigureService,
        private msgHandler: MessageHandlerService,
        private event: EventService
    ) {
        this._confirmSub = this.confirmService.confirmationConfirm$.subscribe(
            confirmation => {
                if (
                    confirmation &&
                    confirmation.state === ConfirmationState.CONFIRMED
                ) {
                    this.resetConfig();
                }
            }
        );
    }

    getConfig(): Configuration {
        return this._currentConfig;
    }

    setConfig(c: Configuration) {
        this._currentConfig = c;
    }

    getOriginalConfig(): Configuration {
        return this._originalConfig;
    }

    setOriginalConfig(c: Configuration) {
        this._originalConfig = c;
    }

    getLoadingConfigStatus(): boolean {
        return this._loadingConfig;
    }

    updateConfig() {
        this._loadingConfig = true;
        this.configureService
            .getConfigurations()
            .pipe(finalize(() => (this._loadingConfig = false)))
            .subscribe(
                res => {
                    this._currentConfig = res as Configuration;
                    this.event.publish(HarborEvent.REFRESH_BANNER_MESSAGE);
                    // Add password fields
                    this._currentConfig.email_password = new StringValueItem(
                        fakePass,
                        true
                    );
                    this._currentConfig.ldap_search_password =
                        new StringValueItem(fakePass, true);
                    this._currentConfig.uaa_client_secret = new StringValueItem(
                        fakePass,
                        true
                    );
                    this._currentConfig.oidc_client_secret =
                        new StringValueItem(fakePass, true);
                    if (!this._currentConfig.disabled_audit_log_event_types) {
                        this._currentConfig.disabled_audit_log_event_types =
                            new StringValueItem('', true);
                    }
                    // Keep the original copy of the data
                    this._originalConfig = clone(this._currentConfig);
                },
                error => {
                    this.msgHandler.handleError(error);
                }
            );
    }

    resetConfig() {
        if (this._originalConfig) {
            this._currentConfig = clone(this._originalConfig);
        }
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

    saveConfiguration(changes: any): Observable<any> {
        return this.configureService.updateConfigurations({
            configurations: changes,
        });
    }
}
