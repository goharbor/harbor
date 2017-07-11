import { Component, Input, Output, EventEmitter, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';

import { SYSTEM_SETTINGS_HTML } from './system-settings.component.html';
import { Configuration } from '../config';

@Component({
    selector: 'system-settings',
    template: SYSTEM_SETTINGS_HTML
})
export class SystemSettingsComponent {
    config: Configuration;
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

    @ViewChild("systemConfigFrom") systemSettingsForm: NgForm;

    get editable(): boolean {
        return this.systemSettings &&
            this.systemSettings.token_expiration &&
            this.systemSettings.token_expiration.editable;
    }

    get isValid(): boolean {
        return this.systemSettingsForm && this.systemSettingsForm.valid;
    }

    get canDownloadCert(): boolean {
        return this.hasAdminRole && this.hasCAFile;
    }
}