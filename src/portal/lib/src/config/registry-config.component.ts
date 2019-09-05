import { Component, OnInit, ViewChild, Input } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

import { ConfirmationState, ConfirmationTargets } from '../shared/shared.const';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ConfirmationMessage } from '../confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../confirmation-dialog/confirmation-state-message';
import { ConfigurationService, SystemInfoService, SystemInfo } from '../service/index';
import {
    compareValue,
    isEmptyObject,
    clone
} from '../utils';
import { ErrorHandler } from '../error-handler/index';
import { Configuration } from './config';
import { VulnerabilityConfigComponent } from "./vulnerability/vulnerability-config.component";
import { GcComponent } from "./gc";
import { SystemSettingsComponent } from "./system/system-settings.component";

@Component({
    selector: 'hbr-registry-config',
    templateUrl: './registry-config.component.html'
})
export class RegistryConfigComponent implements OnInit {
    config: Configuration = new Configuration();
    configCopy: Configuration;
    onGoing: boolean = false;
    systemInfo: SystemInfo;

    @Input() hasAdminRole: boolean = false;

    @ViewChild("systemSettings", {static: false}) systemSettings: SystemSettingsComponent;
    @ViewChild("vulnerabilityConfig", {static: false}) vulnerabilityCfg: VulnerabilityConfigComponent;
    @ViewChild("gc", {static: false}) gc: GcComponent;
    @ViewChild("cfgConfirmationDialog", {static: false}) confirmationDlg: ConfirmationDialogComponent;

    constructor(
        private configService: ConfigurationService,
        private errorHandler: ErrorHandler,
        private translate: TranslateService,
        private systemInfoService: SystemInfoService
    ) { }

    get shouldDisable(): boolean {
        return !this.isValid() || !this.hasChanges() || this.onGoing;
    }

    get hasCAFile(): boolean {
        return this.systemInfo && this.systemInfo.has_ca_root;
    }

    get withClair(): boolean {
        return this.systemInfo && this.systemInfo.with_clair;
    }

    get withAdmiral(): boolean {
        return this.systemInfo && this.systemInfo.with_admiral;
    }

    ngOnInit(): void {
        this.loadSystemInfo();
        // Initialize
        this.load();
    }

    isValid(): boolean {
        return this.systemSettings &&
            this.systemSettings.isValid &&
            this.vulnerabilityCfg &&
            this.vulnerabilityCfg.isValid;
    }

    hasChanges(): boolean {
        return !isEmptyObject(this.getChanges());
    }

    // Get system info
    loadSystemInfo(): void {
        this.systemInfoService.getSystemInfo()
        .subscribe((info: SystemInfo) => {
            this.systemInfo = info;
        }, error => this.errorHandler.error(error));
    }

    // Load configurations
    load(): void {
        this.onGoing = true;
        this.configService.getConfigurations()
            .subscribe((config: Configuration) => {
                this.configCopy = clone(config);
                this.config = config;
                this.onGoing = false;
            }, error => {
                this.errorHandler.error(error);
                this.onGoing = false;
            });
    }

    // Save configuration changes
    save(): void {
        let changes: { [key: string]: any | any[] } = this.getChanges();

        if (isEmptyObject(changes)) {
            // Guard code, do nothing
            return;
        }

        this.onGoing = true;
        this.configService.saveConfigurations(changes)
            .subscribe(() => {
                this.onGoing = false;

                this.translate.get("CONFIG.SAVE_SUCCESS").subscribe((res: string) => {
                    this.errorHandler.info(res);
                });
                // Reload to fetch all the updates
                this.load();
                // Reload all system info
                // this.loadSystemInfo();
            }
            , error => {
                this.onGoing = false;
                this.errorHandler.error(error);
            });
    }

    // Cancel the changes if have
    cancel(): void {
        let msg = new ConfirmationMessage(
            "CONFIG.CONFIRM_TITLE",
            "CONFIG.CONFIRM_SUMMARY",
            "",
            {},
            ConfirmationTargets.CONFIG
        );
        this.confirmationDlg.open(msg);
    }

    // Confirm cancel
    confirmCancel(ack: ConfirmationAcknowledgement): void {
        if (ack && ack.source === ConfirmationTargets.CONFIG &&
            ack.state === ConfirmationState.CONFIRMED) {
            this.reset();
        }
    }

    reset(): void {
        // Reset to the values of copy
        let changes: { [key: string]: any | any[] } = this.getChanges();
        for (let prop of Object.keys(changes)) {
            this.config[prop] = clone(this.configCopy[prop]);
        }
    }

    getChanges(): { [key: string]: any | any[] } {
        let changes: { [key: string]: any | any[] } = {};
        if (!this.config || !this.configCopy) {
            return changes;
        }

        for (let prop of Object.keys(this.config)) {
            let field = this.configCopy[prop];
            if (field && field.editable) {
                if (!compareValue(field.value, this.config[prop].value)) {
                    changes[prop] = this.config[prop].value;
                    // Number
                    if (typeof field.value === "number") {
                        changes[prop] = +changes[prop];
                    }

                    // Trim string value
                    if (typeof field.value === "string") {
                        changes[prop] = ('' + changes[prop]).trim();
                    }
                }
            }
        }

        return changes;
    }
}
