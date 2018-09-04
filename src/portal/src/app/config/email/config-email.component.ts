// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { Component, Input, ViewChild } from '@angular/core';
import { NgForm } from '@angular/forms';

import { Configuration } from 'harbor-ui';

@Component({
    selector: 'config-email',
    templateUrl: "config-email.component.html",
    styleUrls: ['./config-email.component.scss', '../config.component.scss']
})
export class ConfigurationEmailComponent {
    // tslint:disable-next-line:no-input-rename
    @Input("mailConfig") currentConfig: Configuration = new Configuration();

    @ViewChild("mailConfigFrom") mailForm: NgForm;

    constructor() { }

    disabled(prop: any): boolean {
        return !(prop && prop.editable);
    }

    setInsecureValue($event: any) {
        this.currentConfig.email_insecure.value = !$event;
    }

    public isValid(): boolean {
        return this.mailForm && this.mailForm.valid;
    }
}
