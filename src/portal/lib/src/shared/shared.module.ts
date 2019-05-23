import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClientModule, HttpClient} from '@angular/common/http';
import { ClarityModule } from '@clr/angular';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { TranslateModule, TranslateLoader, MissingTranslationHandler } from '@ngx-translate/core';
import { CookieService, CookieModule } from 'ngx-cookie';
import { MarkdownModule } from 'ngx-markdown';
import { TranslateHttpLoader } from '@ngx-translate/http-loader';

import { ClipboardModule } from '../third-party/ngx-clipboard/index';
import { MyMissingTranslationHandler } from '../i18n/missing-trans.handler';
import { TranslatorJsonLoader } from '../i18n/local-json.loader';
import { IServiceConfig, SERVICE_CONFIG } from '../service.config';

/*export function HttpLoaderFactory(http: Http) {
    return new TranslateHttpLoader(http, 'i18n/lang/', '-lang.json');
}

export function LocalJsonLoaderFactory() {
    return new TranslatorJsonLoader();
}*/

export function GeneralTranslatorLoader(http: HttpClient, config: IServiceConfig) {
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
 **
 * class SharedModule
 */
@NgModule({
    imports: [
        CommonModule,
        HttpClientModule,
        FormsModule,
        ReactiveFormsModule,
        ClipboardModule,
        ClarityModule,
        CookieModule.forRoot(),
        MarkdownModule.forRoot(),
        TranslateModule.forRoot({
            loader: {
                provide: TranslateLoader,
                useFactory: (GeneralTranslatorLoader),
                deps: [HttpClient, SERVICE_CONFIG]
            },
            missingTranslationHandler: {
                provide: MissingTranslationHandler,
                useClass: MyMissingTranslationHandler
            }
        }),
    ],
    exports: [
        CommonModule,
        HttpClientModule,
        FormsModule,
        ReactiveFormsModule,
        ClipboardModule,
        ClarityModule,
        CookieModule,
        MarkdownModule,
        TranslateModule,
    ],
    providers: [CookieService]
})
export class SharedModule { }
