import { Component, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';
import {
    BannerMessage,
    BannerMessageI18nMap,
    BannerMessageType,
    Configuration,
} from '../config';
import {
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

@Component({
    selector: 'system-settings',
    templateUrl: './system-settings.component.html',
    styleUrls: ['./system-settings.component.scss'],
})
export class SystemSettingsComponent implements OnInit, OnDestroy {
    bannerMessageTypes: string[] = Object.values(BannerMessageType);
    onGoing = false;
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
    @ViewChild('systemConfigFrom') systemSettingsForm: NgForm;

    constructor(
        private appConfigService: AppConfigService,
        private errorHandler: MessageHandlerService,
        private conf: ConfigService,
        private event: EventService
    ) {
        this.downloadLink = CURRENT_BASE_HREF + '/systeminfo/getcert';
    }

    ngOnInit() {
        this.conf.resetConfig();
        if (!this.bannerRefreshSub) {
            this.bannerRefreshSub = this.event.subscribe(
                HarborEvent.REFRESH_BANNER_MESSAGE,
                () => {
                    this.setValueForBannerMessage();
                }
            );
        }
        if (this.currentConfig.banner_message) {
            this.setValueForBannerMessage();
        }
    }

    ngOnDestroy() {
        if (this.bannerRefreshSub) {
            this.bannerRefreshSub.unsubscribe();
            this.bannerRefreshSub = null;
        }
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
                prop === 'audit_log_track_ip_address' ||
                prop === 'audit_log_track_user_agent' ||
                prop === 'session_timeout' ||
                prop === 'scanner_skip_update_pulltime' ||
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
