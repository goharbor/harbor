import { Component, Input, Output, EventEmitter, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';
import { Configuration } from '../config';

@Component({
    selector: 'replication-config',
    templateUrl: './replication-config.component.html',
    styles: ['./replication-config.component.scss', '../registry-config.component.scss']
})
export class ReplicationConfigComponent {
    config: Configuration;
    @Output() configChange: EventEmitter<Configuration> = new EventEmitter<Configuration>();

    @Input()
    get replicationConfig(): Configuration {
        return this.config;
    }
    set replicationConfig(cfg: Configuration) {
        this.config = cfg;
        this.configChange.emit(this.config);
    }

    @Input() showSubTitle: boolean = false;

    @ViewChild("replicationConfigFrom", { static: false }) replicationConfigForm: NgForm;

    get editable(): boolean {
        return this.replicationConfig &&
            this.replicationConfig.verify_remote_cert &&
            this.replicationConfig.verify_remote_cert.editable;
    }

    get isValid(): boolean {
        return this.replicationConfigForm && this.replicationConfigForm.valid;
    }
}
