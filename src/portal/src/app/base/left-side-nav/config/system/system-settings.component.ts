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
import {
    AfterViewChecked,
    ChangeDetectorRef,
    Component,
    OnDestroy,
    OnInit,
    ViewChild,
} from '@angular/core';
import { NgForm } from '@angular/forms';
import {
    BannerMessage,
    BannerMessageI18nMap,
    BannerMessageType,
    Configuration,
} from '../config';
import {
    clone,
    CURRENT_BASE_HREF,
    getChanges,
    isEmpty,
} from '../../../../shared/units/utils';
import { ConfigService } from '../config.service';
import { AppConfigService } from '../../../../services/app-config.service';
import { finalize } from 'rxjs/operators';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import {
    EventService,
    HarborEvent,
} from '../../../../services/event-service/event.service';
import { Subscription } from 'rxjs';
import { AuditlogService } from 'ng-swagger-gen/services';
import { AuditLogEventType } from 'ng-swagger-gen/models';

@Component({
    selector: 'system-settings',
    templateUrl: './system-settings.component.html',
    styleUrls: ['./system-settings.component.scss'],
})
export class SystemSettingsComponent
    implements OnInit, OnDestroy, AfterViewChecked
{
    bannerMessageTypes: string[] = Object.values(BannerMessageType);
    onGoing = false;
    loading = false;
    downloadLink: string;
    get currentConfig(): Configuration {
        return this.conf.getConfig();
    }

    set currentConfig(cfg: Configuration) {
        this.conf.setConfig(cfg);
    }

    messageText: string;
    messageType: string;
    messageClosable: boolean;
    messageFromDate: Date;
    messageToDate: Date;
    // the copy of bannerMessage
    messageTextCopy: string;
    messageTypeCopy: string;
    messageClosableCopy: boolean;
    messageFromDateCopy: Date;
    messageToDateCopy: Date;
    bannerRefreshSub: Subscription;
    currentDate: Date = new Date();
    logEventTypes: Record<string, string>[] = [];
    selectedLogEventTypes: string[] = clone([]);
    @ViewChild('systemConfigFrom') systemSettingsForm: NgForm;

    constructor(
        private appConfigService: AppConfigService,
        private conf: ConfigService,
        private logService: AuditlogService,
        private event: EventService,
        private errorHandler: MessageHandlerService,
        private changeDetectorRef: ChangeDetectorRef
    ) {
        this.downloadLink = CURRENT_BASE_HREF + '/systeminfo/getcert';
    }

    ngOnInit() {
        this.conf.resetConfig();
        if (!this.bannerRefreshSub) {
            this.bannerRefreshSub = this.event?.subscribe(
                HarborEvent.REFRESH_BANNER_MESSAGE,
                () => {
                    this.setValueForBannerMessage();
                    this.setValueForEnabledAuditLogEventTypes();
                }
            );
        }
        if (this.currentConfig.banner_message) {
            this.setValueForBannerMessage();
        }
        this.initLogEventTypes();
        this.setValueForEnabledAuditLogEventTypes();
    }

    ngAfterViewChecked() {
        this.changeDetectorRef.detectChanges();
    }

    ngOnDestroy() {
        if (this.bannerRefreshSub) {
            this.bannerRefreshSub.unsubscribe();
            this.bannerRefreshSub = null;
        }
    }

    initLogEventTypes() {
        this.loading = true;
        this.logService
            .listAuditLogEventTypesResponse()
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                response => {
                    const auditLogEventTypes =
                        response.body as AuditLogEventType[];
                    this.logEventTypes = auditLogEventTypes.map(event => ({
                        label:
                            event.event_type.charAt(0).toUpperCase() +
                            event.event_type.slice(1).replace(/_/g, ' '),
                        value: event.event_type,
                        id: event.event_type,
                    }));
                    this.setValueForEnabledAuditLogEventTypes();
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }

    setValueForEnabledAuditLogEventTypes() {
        const disabledEventTypes =
            this.currentConfig?.disabled_audit_log_event_types?.value;
        const disabledEvents =
            disabledEventTypes?.split(',')?.filter(evt => evt !== '') ?? [];

        const allEventTypes = this.logEventTypes.map(evt => evt.value);

        // Enabled = All - Disabled
        this.selectedLogEventTypes = allEventTypes.filter(
            evt => !disabledEvents.includes(evt)
        );
    }

    setValueForBannerMessage() {
        if (this.currentConfig.banner_message.value) {
            this.messageText = (
                JSON.parse(
                    this.currentConfig.banner_message.value
                ) as BannerMessage
            ).message;
            this.messageType = (
                JSON.parse(
                    this.currentConfig.banner_message.value
                ) as BannerMessage
            ).type;
            this.messageClosable = (
                JSON.parse(
                    this.currentConfig.banner_message.value
                ) as BannerMessage
            ).closable;
            this.messageFromDate = (
                JSON.parse(
                    this.currentConfig.banner_message.value
                ) as BannerMessage
            ).fromDate;
            this.messageToDate = (
                JSON.parse(
                    this.currentConfig.banner_message.value
                ) as BannerMessage
            ).toDate;
        } else {
            this.messageText = null;
            this.messageType = BannerMessageType.WARNING;
            this.messageClosable = false;
        }
        this.messageTextCopy = this.messageText;
        this.messageTypeCopy = this.messageType;
        this.messageClosableCopy = this.messageClosable;
        this.messageFromDateCopy = this.messageFromDate;
        this.messageToDateCopy = this.messageToDate;
    }

    get editable(): boolean {
        return (
            this.currentConfig &&
            this.currentConfig.token_expiration &&
            this.currentConfig.token_expiration.editable
        );
    }

    get robotExpirationEditable(): boolean {
        return (
            this.currentConfig &&
            this.currentConfig.robot_token_duration &&
            this.currentConfig.robot_token_duration.editable
        );
    }

    robotNamePrefixEditable(): boolean {
        return (
            this.currentConfig &&
            this.currentConfig.robot_name_prefix &&
            this.currentConfig.robot_name_prefix.editable
        );
    }

    public isValid(): boolean {
        return this.systemSettingsForm && this.systemSettingsForm.valid;
    }

    public hasChanges(): boolean {
        return !isEmpty(this.getChanges()) || this.hasBannerMessageChanged();
    }

    hasBannerMessageChanged() {
        return (
            this.messageTextCopy != this.messageText ||
            this.messageTypeCopy != this.messageType ||
            this.messageClosableCopy != this.messageClosable ||
            this.messageFromDateCopy != this.messageFromDate ||
            this.messageToDateCopy != this.messageToDate
        );
    }

    hasLogEventType(resourceType: string): boolean {
        return this.selectedLogEventTypes?.includes(resourceType);
    }

    setLogEventType(resourceType: string) {
        if (this.selectedLogEventTypes.includes(resourceType)) {
            this.selectedLogEventTypes = this.selectedLogEventTypes.filter(
                evt => evt !== resourceType
            );
        } else {
            this.selectedLogEventTypes.push(resourceType);
        }

        const allEventTypes = this.logEventTypes.map(evt => evt.value);
        // Disabled = All - Enabled
        const disabled = allEventTypes.filter(
            evt => !this.selectedLogEventTypes.includes(evt)
        );

        // Update backend config
        this.currentConfig.disabled_audit_log_event_types.value =
            disabled.join(',');
    }

    public getChanges() {
        let allChanges = getChanges(
            this.conf.getOriginalConfig(),
            this.currentConfig
        );
        if (allChanges) {
            return this.getSystemChanges(allChanges);
        }
        return null;
    }

    public getSystemChanges(allChanges: any) {
        let changes = {};
        for (let prop in allChanges) {
            if (
                prop === 'token_expiration' ||
                prop === 'read_only' ||
                prop === 'project_creation_restriction' ||
                prop === 'robot_token_duration' ||
                prop === 'notification_enable' ||
                prop === 'robot_name_prefix' ||
                prop === 'audit_log_forward_endpoint' ||
                prop === 'skip_audit_log_database' ||
                prop === 'session_timeout' ||
                prop === 'scanner_skip_update_pulltime' ||
                prop === 'disabled_audit_log_event_types' ||
                prop === 'banner_message'
            ) {
                changes[prop] = allChanges[prop];
            }
        }
        return changes;
    }

    setRepoReadOnlyValue($event: any) {
        this.currentConfig.read_only.value = $event;
    }

    setWebhookNotificationEnabledValue($event: any) {
        this.currentConfig.notification_enable.value = $event;
    }

    disabled(prop: any): boolean {
        return !(prop && prop.editable);
    }

    get canDownloadCert(): boolean {
        return this.appConfigService.getConfig().has_ca_root;
    }

    /**
     *
     * Save the changed values
     *
     * @memberOf ConfigurationComponent
     */
    public save(): void {
        let changes = this.getChanges();
        if (this.hasBannerMessageChanged()) {
            const bm = new BannerMessage();
            bm.message = this.messageText;
            bm.type = this.messageType;
            bm.closable = this.messageClosable;
            bm.fromDate = this.messageFromDate;
            bm.toDate = this.messageToDate;
            if (bm.message) {
                changes['banner_message'] = JSON.stringify(bm);
            } else {
                changes['banner_message'] = '';
            }
        }
        if (!isEmpty(changes)) {
            this.onGoing = true;
            this.conf
                .saveConfiguration(changes)
                .pipe(finalize(() => (this.onGoing = false)))
                .subscribe({
                    next: result => {
                        // API should return the updated configurations here
                        // Unfortunately API does not do that
                        // So we need to call update function again
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
                        this.errorHandler.info('CONFIG.SAVE_SUCCESS');
                    },
                    error: error => {
                        this.errorHandler.error(error);
                    },
                });
        } else {
            // Inprop situation, should not come here
            console.error('Save abort because nothing changed');
        }
    }
    public get inProgress(): boolean {
        return this.onGoing || this.conf.getLoadingConfigStatus();
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

    checkAuditLogForwardEndpoint(e: any) {
        if (!e?.target?.value) {
            this.currentConfig.skip_audit_log_database.value = false;
        }
    }

    translateMessageType(type: string): string {
        return BannerMessageI18nMap[type] || type;
    }

    minDateForEndDay(): Date {
        return this.messageFromDate ? this.messageFromDate : this.currentDate;
    }
}
