import { Component, Input, Output, EventEmitter, ViewChild, Inject } from '@angular/core';
import { NgForm } from '@angular/forms';
import { Configuration } from '../config';
import { SERVICE_CONFIG, IServiceConfig, downloadUrl } from '../../service.config';
@Component({
    selector: 'system-settings',
    templateUrl: './system-settings.component.html',
    styleUrls: ['./system-settings.component.scss', '../registry-config.component.scss']
})
export class SystemSettingsComponent {
    config: Configuration;
    downloadLink: string = downloadUrl;
    @Output() configChange: EventEmitter<Configuration> = new EventEmitter<Configuration>();

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

    get editable(): boolean {
        return this.systemSettings &&
            this.systemSettings.token_expiration &&
            this.systemSettings.token_expiration.editable;
    }

    get isValid(): boolean {
        return this.systemSettingsForm && this.systemSettingsForm.valid;
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

    constructor( @Inject(SERVICE_CONFIG) private configInfo: IServiceConfig) {
        if (this.configInfo && this.configInfo.systemInfoEndpoint) {
            this.downloadLink = this.configInfo.systemInfoEndpoint + "/getcert";
        }
    }
}
