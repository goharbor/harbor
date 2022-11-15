import { Component, OnInit, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';
import { Configuration } from '../config';
import {
    CURRENT_BASE_HREF,
    getChanges,
    isEmpty,
} from '../../../../shared/units/utils';
import { ConfigService } from '../config.service';
import { AppConfigService } from '../../../../services/app-config.service';
import { finalize } from 'rxjs/operators';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';

@Component({
    selector: 'system-settings',
    templateUrl: './system-settings.component.html',
    styleUrls: ['./system-settings.component.scss'],
})
export class SystemSettingsComponent implements OnInit {
    onGoing = false;
    downloadLink: string;
    get currentConfig(): Configuration {
        return this.conf.getConfig();
    }

    set currentConfig(cfg: Configuration) {
        this.conf.setConfig(cfg);
    }
    @ViewChild('systemConfigFrom') systemSettingsForm: NgForm;

    constructor(
        private appConfigService: AppConfigService,
        private errorHandler: MessageHandlerService,
        private conf: ConfigService
    ) {
        this.downloadLink = CURRENT_BASE_HREF + '/systeminfo/getcert';
    }

    ngOnInit() {
        this.conf.resetConfig();
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

    get tokenExpirationValue() {
        return this.currentConfig.token_expiration.value;
    }

    set tokenExpirationValue(v) {
        // convert string to number
        this.currentConfig.token_expiration.value = +v;
    }

    get sessionTimeout() {
        return this.currentConfig.session_timeout.value;
    }

    set sessionTimeout(v) {
        // convert string to number
        this.currentConfig.session_timeout.value = +v;
    }

    get robotTokenExpirationValue() {
        return this.currentConfig.robot_token_duration.value;
    }

    set robotTokenExpirationValue(v) {
        // convert string to number
        this.currentConfig.robot_token_duration.value = +v;
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
        return !isEmpty(this.getChanges());
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
                prop === 'session_timeout'
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
                        // Handle read only
                        if (changes['read_only']) {
                            this.errorHandler.handleReadOnly();
                        } else {
                            this.errorHandler.clear();
                        }
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
}
