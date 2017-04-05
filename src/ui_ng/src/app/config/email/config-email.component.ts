import { Component, Input, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';

import { Configuration } from '../config';

@Component({
    selector: 'config-email',
    templateUrl: "config-email.component.html",
    styleUrls: ['../config.component.css']
})
export class ConfigurationEmailComponent {
    @Input("mailConfig") currentConfig: Configuration = new Configuration();
    
    @ViewChild("mailConfigFrom") mailForm: NgForm;

    constructor() { }

    private disabled(prop: any): boolean {
        return !(prop && prop.editable);
    }

    public isValid(): boolean {
        return this.mailForm && this.mailForm.valid;
    }
}