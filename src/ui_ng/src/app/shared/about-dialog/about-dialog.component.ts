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
import { Component } from '@angular/core';

import { AppConfigService } from '../../app-config.service';

@Component({
    selector: 'about-dialog',
    templateUrl: "about-dialog.component.html",
    styleUrls: ["about-dialog.component.css"]
})
export class AboutDialogComponent {
    private opened: boolean = false;
    private build: string = "4276418";

    constructor(private appConfigService: AppConfigService) { }

    public get version(): string {
        let appConfig = this.appConfigService.getConfig();
        return appConfig?appConfig.harbor_version: "n/a";
    }

    public open(): void {
        this.opened = true;
    }

    public close(): void {
        this.opened = false;
    }
}