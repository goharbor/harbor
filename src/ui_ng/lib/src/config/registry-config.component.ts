import { Component, OnInit, EventEmitter, Output } from '@angular/core';

import { Configuration, ComplexValueItem } from './config';
import { REGISTRY_CONFIG_HTML } from './registry-config.component.html';
import { ConfigurationService } from '../service/index';
import { toPromise } from '../utils';
import { ErrorHandler } from '../error-handler';

@Component({
    selector: 'hbr-registry-config',
    template: REGISTRY_CONFIG_HTML
})
export class RegistryConfigComponent implements OnInit {
    config: Configuration = new Configuration();
    configCopy: Configuration;

    @Output() configChanged: EventEmitter<any> = new EventEmitter<any>();

    constructor(
        private configService: ConfigurationService,
        private errorHandler: ErrorHandler
    ) { }

    ngOnInit(): void {
        //Initialize
        this.load();
    }

    //Load configurations
    load(): void {
        toPromise<Configuration>(this.configService.getConfigurations())
            .then((config: Configuration) => {
                this.configCopy = Object.assign({}, config);
                this.config = config;
            })
            .catch(error => this.errorHandler.error(error));
    }

    //Save configuration changes
    save(): void {
        let changes: { [key: string]: any | any[] } = this.getChanges();

        if (this._isEmptyObject(changes)) {
            //Guard code, do nothing
            return;
        }

        //Fix policy parameters issue
        let scanningAllPolicy = changes["scan_all_policy"];
        if (scanningAllPolicy &&
            scanningAllPolicy.type !== "daily" &&
            scanningAllPolicy.parameters) {
            delete (scanningAllPolicy.parameters);
        }

        toPromise<any>(this.configService.saveConfigurations(changes))
            .then(() => {
                this.configChanged.emit(changes);
            })
            .catch(error => this.errorHandler.error(error));
    }

    reset(): void {
        //Reset to the values of copy
        let changes: { [key: string]: any | any[] } = this.getChanges();
        for (let prop in changes) {
            this.config[prop] = Object.assign({}, this.configCopy[prop]);
        }
    }

    getChanges(): { [key: string]: any | any[] } {
        let changes: { [key: string]: any | any[] } = {};
        if (!this.config || !this.configCopy) {
            return changes;
        }

        for (let prop in this.config) {
            let field = this.configCopy[prop];
            if (field && field.editable) {
                if (!this._compareValue(field.value, this.config[prop].value)) {
                    changes[prop] = this.config[prop].value;
                    //Number 
                    if (typeof field.value === "number") {
                        changes[prop] = +changes[prop];
                    }

                    //Trim string value
                    if (typeof field.value === "string") {
                        changes[prop] = ('' + changes[prop]).trim();
                    }
                }
            }
        }

        return changes;
    }

    //private
    _compareValue(a: any, b: any): boolean {
        if ((a && !b) || (!a && b)) return false;
        if (!a && !b) return true;

        return JSON.stringify(a) === JSON.stringify(b);
    }

    //private
    _isEmptyObject(obj: any): boolean {
        return !obj || JSON.stringify(obj) === "{}";
    }
}