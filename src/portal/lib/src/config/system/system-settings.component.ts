import { Component, Input, OnInit, Output, EventEmitter, ViewChild, Inject, OnChanges, SimpleChanges } from '@angular/core';
import { NgForm } from '@angular/forms';
import { Configuration, StringValueItem } from '../config';
import { SERVICE_CONFIG, IServiceConfig } from '../../service.config';
import { clone, isEmpty, getChanges } from '../../utils';
import { ErrorHandler } from '../../error-handler/index';
import { ConfirmationMessage } from '../../confirmation-dialog/confirmation-message';
import { ConfirmationDialogComponent } from '../../confirmation-dialog/confirmation-dialog.component';
import { ConfirmationState, ConfirmationTargets } from '../../shared/shared.const';
import { ConfirmationAcknowledgement } from '../../confirmation-dialog/confirmation-state-message';
import {
    ConfigurationService
} from '../../service/index';
import { from } from 'rxjs';
const fakePass = 'aWpLOSYkIzJTTU4wMDkx';
const ONE_HOUR_MINUTES: number = 60;
const ONE_DAY_MINUTES: number = 24 * ONE_HOUR_MINUTES;
@Component({
    selector: 'system-settings',
    templateUrl: './system-settings.component.html',
    styleUrls: ['./system-settings.component.scss', '../registry-config.component.scss']
})
export class SystemSettingsComponent implements OnChanges, OnInit {
    config: Configuration = new Configuration();
    onGoing = false;
    private originalConfig: Configuration;
    downloadLink: string;
    robotTokenExpiration: string;
    @Output() configChange: EventEmitter<Configuration> = new EventEmitter<Configuration>();
    @Output() readOnlyChange: EventEmitter<boolean> = new EventEmitter<boolean>();
    @Output() reloadSystemConfig: EventEmitter<any> = new EventEmitter<any>();

    @Input()
    get systemSettings(): Configuration {
        return this.config;
    }
    set systemSettings(cfg: Configuration) {
        this.config = cfg;
        this.configChange.emit(this.config);
    }

    @Input() showSubTitle: boolean = false;
    @Input() hasAdminRole: boolean = false;
    @Input() hasCAFile: boolean = false;
    @Input() withAdmiral = false;

    @ViewChild("systemConfigFrom") systemSettingsForm: NgForm;
    @ViewChild("cfgConfirmationDialog") confirmationDlg: ConfirmationDialogComponent;

    get editable(): boolean {
        return this.systemSettings &&
            this.systemSettings.token_expiration &&
            this.systemSettings.token_expiration.editable;
    }

    get robotExpirationEditable(): boolean {
        return this.systemSettings &&
            this.systemSettings.robot_token_duration &&
            this.systemSettings.robot_token_duration.editable;
    }

    public isValid(): boolean {
        return this.systemSettingsForm && this.systemSettingsForm.valid;
    }

    public hasChanges(): boolean {
        return !isEmpty(this.getChanges());
    }

    public getChanges() {
        let allChanges = getChanges(this.originalConfig, this.config);
        if (allChanges) {
            return this.getSystemChanges(allChanges);
        }
        return null;
    }

    ngOnChanges(changes: SimpleChanges): void {
        if (changes && changes["systemSettings"]) {
            this.originalConfig = clone(this.config);
        }
    }

    public getSystemChanges(allChanges: any) {
        let changes = {};
        for (let prop in allChanges) {
            if (prop === 'token_expiration' || prop === 'read_only' || prop === 'project_creation_restriction'
            || prop === 'robot_token_duration'  || prop === 'ip_white') {
                changes[prop] = allChanges[prop];
            }
        }
        return changes;
    }

    setRepoReadOnlyValue($event: any) {
        this.systemSettings.read_only.value = $event;
    }

    disabled(prop: any): boolean {
        return !(prop && prop.editable);
    }

    get canDownloadCert(): boolean {
        return this.hasAdminRole && this.hasCAFile;
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
            this.configService.saveConfigurations(changes)
                .subscribe(response => {
                    this.onGoing = false;
                    // API should return the updated configurations here
                    // Unfortunately API does not do that
                    // To refresh the view, we can clone the original data copy
                    // or force refresh by calling service.
                    // HERE we choose force way
                    this.retrieveConfig();
                    if ('read_only' in changes) {
                        this.readOnlyChange.emit(changes['read_only']);
                    }

                    this.reloadSystemConfig.emit();
                    this.errorHandler.info('CONFIG.SAVE_SUCCESS');
                }
                , error => {
                    this.onGoing = false;
                    this.errorHandler.error(error);
                });
        } else {
            // Inprop situation, should not come here
            console.error('Save abort because nothing changed');
        }
    }

    retrieveConfig(): void {
        this.onGoing = true;
        this.configService.getConfigurations()
            .subscribe((configurations: Configuration) => {
                this.onGoing = false;
                // Add two password fields
                configurations.email_password = new StringValueItem(fakePass, true);
                this.config = configurations;
                // Keep the original copy of the data
                this.originalConfig = clone(configurations);
            }, error => {
                this.onGoing = false;
                this.errorHandler.error(error);
            });
    }

    reset(changes: any): void {
        if (!isEmpty(changes)) {
            for (let prop in changes) {
                if (this.originalConfig[prop]) {
                    this.config[prop] = clone(this.originalConfig[prop]);
                }
            }
        } else {
            // force reset
            this.retrieveConfig();
        }
    }

    confirmCancel(ack: ConfirmationAcknowledgement): void {
        if (ack && ack.source === ConfirmationTargets.CONFIG &&
            ack.state === ConfirmationState.CONFIRMED) {
            let changes = this.getChanges();
            this.reset(changes);
            this.initRobotToken();
        }
    }


    public get inProgress(): boolean {
        return this.onGoing;
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
            let msg = new ConfirmationMessage(
                'CONFIG.CONFIRM_TITLE',
                'CONFIG.CONFIRM_SUMMARY',
                '',
                {},
                ConfirmationTargets.CONFIG
            );
            this.confirmationDlg.open(msg);
        } else {
            // Invalid situation, should not come here
            console.error('Nothing changed');
        }
    }

    constructor(@Inject(SERVICE_CONFIG) private configInfo: IServiceConfig,
        private configService: ConfigurationService,
        private errorHandler: ErrorHandler) {
        if (this.configInfo && this.configInfo.systemInfoEndpoint) {
            this.downloadLink = this.configInfo.systemInfoEndpoint + "/getcert";
        }
    }
    ngOnInit() {
        this.initRobotToken();
    }

    private initRobotToken (): void {
        if (this.config &&
            this.config.robot_token_duration ) {
            let robotExpiration = this.config.robot_token_duration.value;
            this.robotTokenExpiration = Math.floor(robotExpiration / ONE_DAY_MINUTES) + '';
        }
    }
    changeToken(v: string) {
        if (!v || v === "") {
            return;
        }
        if (!(this.config &&
            this.config.robot_token_duration)) {
            return;
        }
        this.config.robot_token_duration.value = +v * ONE_DAY_MINUTES;
    }


}
