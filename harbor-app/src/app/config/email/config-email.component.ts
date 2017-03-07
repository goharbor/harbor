import { Component, Input } from '@angular/core';
import { NgForm } from '@angular/forms';

import { Configuration } from '../config';

@Component({
    selector: 'config-email',
    templateUrl: "config-email.component.html",
    styleUrls: ['../config.component.css']
})
export class ConfigurationEmailComponent {
    @Input("mailConfig") currentConfig: Configuration = new Configuration();
    
    constructor() { }
}