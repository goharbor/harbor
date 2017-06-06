import { Injectable } from "@angular/core";
import { i18nConfig } from "./i18n-config";
import { TranslateService } from '@ngx-translate/core';
import { DEFAULT_LANG_COOKIE_KEY, DEFAULT_SUPPORTING_LANGS, DEFAULT_LANG } from '../utils';
import { CookieService } from 'ngx-cookie';

@Injectable()
export class TranslateServiceInitializer {
    constructor(
        private translateService: TranslateService,
        private cookie: CookieService
    ) { }

    public init(config: i18nConfig = {}): void {
        let selectedLang: string = config.defaultLang ? config.defaultLang : DEFAULT_LANG;

        this.translateService.addLangs(config.supportedLangs ? config.supportedLangs : DEFAULT_SUPPORTING_LANGS);
        this.translateService.setDefaultLang(selectedLang);

        if (config.enablei18Support) {
            //If user has selected lang, then directly use it
            let langSetting: string = this.cookie.get(config.langCookieKey ? config.langCookieKey : DEFAULT_LANG_COOKIE_KEY);
            if (!langSetting || langSetting.trim() === "") {
                //Use browser lang
                langSetting = this.translateService.getBrowserCultureLang().toLowerCase();
            }

            if (config.supportedLangs && config.supportedLangs.length > 0) {
                if (config.supportedLangs.find(lang => lang === langSetting)) {
                    selectedLang = langSetting;
                }
            }
        }

        this.translateService.use(selectedLang);
    }
}