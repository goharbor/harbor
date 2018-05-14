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
import { BrowserModule } from '@angular/platform-browser';
import { NgModule, APP_INITIALIZER, LOCALE_ID } from '@angular/core';
import { AppComponent } from './app.component';

import { BaseModule } from './base/base.module';
import { HarborRoutingModule } from './harbor-routing.module';
import { SharedModule } from './shared/shared.module';
import { AccountModule } from './account/account.module';
import { ConfigurationModule } from './config/config.module';

import { TranslateService } from "@ngx-translate/core";
import { AppConfigService } from './app-config.service';
import {SkinableConfig} from "./skinable-config.service";
import { ProjectConfigComponent } from './project/project-config/project-config.component';

export function initConfig(configService: AppConfigService, skinableService: SkinableConfig) {
    return () => {
        skinableService.getCustomFile();
        configService.load();
    };
}

export function getCurrentLanguage(translateService: TranslateService) {
    return translateService.currentLang;
}

@NgModule({
    declarations: [
        AppComponent,
        ProjectConfigComponent,
    ],
    imports: [
        BrowserModule,
        SharedModule,
        BaseModule,
        AccountModule,
        HarborRoutingModule,
        ConfigurationModule,
    ],
    providers: [
      AppConfigService,
      SkinableConfig,
      {
        provide: APP_INITIALIZER,
        useFactory: initConfig,
        deps: [ AppConfigService, SkinableConfig],
        multi: true
      },
      {
        provide: LOCALE_ID,
        useFactory: getCurrentLanguage,
        deps: [ TranslateService ]
      }
    ],
    bootstrap: [AppComponent]
})
export class AppModule {}
