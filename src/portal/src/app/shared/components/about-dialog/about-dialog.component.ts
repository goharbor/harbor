// Copyright Project Harbor Authors
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
import { Component, OnInit } from '@angular/core';
import { AppConfigService } from '../../../services/app-config.service';
import { SkinableConfig } from '../../../services/skinable-config.service';

@Component({
    selector: 'about-dialog',
    templateUrl: 'about-dialog.component.html',
    styleUrls: ['about-dialog.component.scss'],
})
export class AboutDialogComponent implements OnInit {
    opened: boolean = false;
    build: string = '4276418';
    customIntroduction: string;
    customName: string;
    customLogo: string;

    constructor(
        private appConfigService: AppConfigService,
        private skinableConfig: SkinableConfig
    ) {}

    ngOnInit(): void {
        // custom skin
        let customSkinObj = this.skinableConfig.getSkinConfig();
        if (customSkinObj) {
            if (customSkinObj.product) {
                this.customLogo = customSkinObj.product.logo;
                this.customName = customSkinObj.product.name;
                this.customIntroduction = customSkinObj.product.introduction;
            }
        }
    }

    public get version(): string {
        let appConfig = this.appConfigService.getConfig();
        return appConfig.harbor_version;
    }

    public open(): void {
        this.opened = true;
    }

    public close(): void {
        this.opened = false;
    }
}
