import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpModule, Http } from '@angular/http';
import { ClarityModule } from 'clarity-angular';
import { FormsModule } from '@angular/forms';
import { TranslateModule, TranslateLoader, TranslateService, MissingTranslationHandler } from "@ngx-translate/core";
import { MyMissingTranslationHandler } from '../i18n/missing-trans.handler';
import { TranslateHttpLoader } from '@ngx-translate/http-loader';
import { TranslatorJsonLoader } from '../i18n/local-json.loader';


/*export function HttpLoaderFactory(http: Http) {
    return new TranslateHttpLoader(http, 'i18n/lang/', '-lang.json');
}*/

export function LocalJsonLoaderFactory() {
    return new TranslatorJsonLoader();
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
        ClarityModule.forRoot(),
        TranslateModule.forRoot({
            loader: {
                provide: TranslateLoader,
                useFactory: (LocalJsonLoaderFactory)
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
        ClarityModule,
        TranslateModule
    ]
})

export class SharedModule { }