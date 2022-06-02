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
import { BrowserModule } from '@angular/platform-browser';
import {
    NgModule,
    APP_INITIALIZER,
    CUSTOM_ELEMENTS_SCHEMA,
} from '@angular/core';
import { AppComponent } from './app.component';
import { InterceptHttpService } from './services/intercept-http.service';
import { HarborRoutingModule } from './harbor-routing.module';
import { AppConfigService } from './services/app-config.service';
import { SkinableConfig } from './services/skinable-config.service';
import { HTTP_INTERCEPTORS, HttpClientModule } from '@angular/common/http';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { CookieModule } from 'ngx-cookie';
import {
    MissingTranslationHandler,
    MissingTranslationHandlerParams,
    TranslateLoader,
    TranslateModule,
} from '@ngx-translate/core';
import {
    ProjectDefaultService,
    ProjectService,
    UserPermissionDefaultService,
    UserPermissionService,
} from './shared/services';
import { ErrorHandler } from './shared/units/error-handler';
import { MessageHandlerService } from './shared/services/message-handler.service';
import { HarborTranslateLoaderService } from './services/harbor-translate-loader.service';

function initConfig(
    configService: AppConfigService,
    skinableService: SkinableConfig
) {
    return () => {
        skinableService.getCustomFile().subscribe();
        configService.load().subscribe();
    };
}

class MyMissingTranslationHandler implements MissingTranslationHandler {
    handle(params: MissingTranslationHandlerParams) {
        const missingText: string = '{Harbor}';
        return params.key || missingText;
    }
}

@NgModule({
    declarations: [AppComponent],
    imports: [
        TranslateModule.forRoot({
            loader: {
                provide: TranslateLoader,
                useClass: HarborTranslateLoaderService,
            },
            missingTranslationHandler: {
                provide: MissingTranslationHandler,
                useClass: MyMissingTranslationHandler,
            },
        }),
        BrowserModule,
        BrowserAnimationsModule,
        HttpClientModule,
        HarborRoutingModule,
        CookieModule.forRoot(),
    ],
    providers: [
        AppConfigService,
        SkinableConfig,
        {
            provide: APP_INITIALIZER,
            useFactory: initConfig,
            deps: [AppConfigService, SkinableConfig],
            multi: true,
        },
        {
            provide: HTTP_INTERCEPTORS,
            useClass: InterceptHttpService,
            multi: true,
        },
        { provide: ProjectService, useClass: ProjectDefaultService },
        { provide: ErrorHandler, useClass: MessageHandlerService },
        {
            provide: UserPermissionService,
            useClass: UserPermissionDefaultService,
        },
    ],
    schemas: [CUSTOM_ELEMENTS_SCHEMA],
    bootstrap: [AppComponent],
})
export class AppModule {}
