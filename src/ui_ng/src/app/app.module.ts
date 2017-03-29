import { BrowserModule } from '@angular/platform-browser';
import { NgModule, APP_INITIALIZER, LOCALE_ID } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpModule } from '@angular/http';
import { ClarityModule } from 'clarity-angular';
import { AppComponent } from './app.component';

import { BaseModule } from './base/base.module';
import { HarborRoutingModule } from './harbor-routing.module';
import { SharedModule } from './shared/shared.module';
import { AccountModule } from './account/account.module';
import { ConfigurationModule } from './config/config.module';

import { TranslateModule, TranslateLoader, TranslateService, MissingTranslationHandler } from "@ngx-translate/core";
import { MyMissingTranslationHandler } from './i18n/missing-trans.handler';
import { TranslateHttpLoader } from '@ngx-translate/http-loader';
import { Http } from '@angular/http';

import { AppConfigService } from './app-config.service';

export function HttpLoaderFactory(http: Http) {
    return new TranslateHttpLoader(http, 'i18n/lang/', '-lang.json');
}

export function initConfig(configService: AppConfigService) {
    return () => configService.load();
}

function getCurrentLanguage(translateService: TranslateService) {
    return translateService.currentLang;
}

@NgModule({
    declarations: [
        AppComponent,
    ],
    imports: [
        SharedModule,
        BaseModule,
        AccountModule,
        HarborRoutingModule,
        ConfigurationModule,
        TranslateModule.forRoot({
            loader: {
                provide: TranslateLoader,
                useFactory: (HttpLoaderFactory),
                deps: [Http]
            },
            missingTranslationHandler: {
                provide: MissingTranslationHandler,
                useClass: MyMissingTranslationHandler
            }
        })
    ],
    providers: [
      AppConfigService,
      { 
        provide: APP_INITIALIZER, 
        useFactory: initConfig, 
        deps: [ AppConfigService ],
        multi: true
      },
      {
        provide: LOCALE_ID,
        useFactory: getCurrentLanguage,
        deps:[ TranslateService ]
      }
    ],
    bootstrap: [AppComponent]
})
export class AppModule {}
