import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpModule, Http } from '@angular/http';
import { ClarityModule } from 'clarity-angular';
import { FormsModule } from '@angular/forms';
import { TranslateModule, TranslateLoader, TranslateService, MissingTranslationHandler } from "@ngx-translate/core";
import { MyMissingTranslationHandler } from '../i18n/missing-trans.handler';
import { TranslateHttpLoader } from '@ngx-translate/http-loader';
import { TranslatorJsonLoader } from '../i18n/local-json.loader';
import { IServiceConfig, SERVICE_CONFIG } from '../service.config';
import { CookieService, CookieModule } from 'ngx-cookie';

/*export function HttpLoaderFactory(http: Http) {
    return new TranslateHttpLoader(http, 'i18n/lang/', '-lang.json');
}

export function LocalJsonLoaderFactory() {
    return new TranslatorJsonLoader();
}*/

export function GeneralTranslatorLoader(http: Http, config: IServiceConfig) {
    if (config && config.langMessageLoader === 'http') {
        let prefix: string = config.langMessagePathForHttpLoader ? config.langMessagePathForHttpLoader : "i18n/lang/";
        let suffix: string = config.langMessageFileSuffixForHttpLoader ? config.langMessageFileSuffixForHttpLoader : "-lang.json";
        return new TranslateHttpLoader(http, prefix, suffix);
    } else {
        return new TranslatorJsonLoader(config);
    }
}

/**
 * 
 * Module for sharing common modules
 * 
 * @export
 * @class SharedModule
 */
@NgModule({
    imports: [
        CommonModule,
        HttpModule,
        FormsModule,
        CookieModule.forRoot(),
        ClarityModule.forRoot(),
        TranslateModule.forRoot({
            loader: {
                provide: TranslateLoader,
                useFactory: (GeneralTranslatorLoader),
                deps: [Http, SERVICE_CONFIG]
            },
            missingTranslationHandler: {
                provide: MissingTranslationHandler,
                useClass: MyMissingTranslationHandler
            }
        })
    ],
    exports: [
        CommonModule,
        HttpModule,
        FormsModule,
        CookieModule,
        ClarityModule,
        TranslateModule
    ],
    providers: [CookieService]
})

export class SharedModule { }