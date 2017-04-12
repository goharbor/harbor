import { Component, ReflectiveInjector, LOCALE_ID } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { CookieService } from 'angular2-cookie/core';

import { supportedLangs, enLang } from './shared/shared.const';
import { SessionService } from './shared/session.service';
import { AppConfigService } from './app-config.service';
import { Title } from '@angular/platform-browser';

@Component({
    selector: 'harbor-app',
    templateUrl: 'app.component.html'
})
export class AppComponent {
    constructor(
        private translate: TranslateService,
        private cookie: CookieService,
        private session: SessionService,
        private appConfigService: AppConfigService,
        private titleService: Title) {

        translate.addLangs(supportedLangs);
        translate.setDefaultLang(enLang);

        //If user has selected lang, then directly use it
        let langSetting = this.cookie.get("harbor-lang");
        if (!langSetting || langSetting.trim() === "") {
            //Use browser lang
            langSetting = translate.getBrowserCultureLang().toLowerCase();
        }

        let selectedLang = this.isLangMatch(langSetting, supportedLangs) ? langSetting : enLang;
        translate.use(selectedLang);
        //this.session.switchLanguage(selectedLang).catch(error => console.error(error));

        //Override page title
        let key: string = "APP_TITLE.HARBOR";
        if (this.appConfigService.isIntegrationMode()) {
            key = "APP_TITLE.REG";
        }

        translate.get(key).subscribe((res: string) => {
            this.titleService.setTitle(res);
        });
    }

    private isLangMatch(browserLang: string, supportedLangs: string[]) {
        if (supportedLangs && supportedLangs.length > 0) {
            return supportedLangs.find(lang => lang === browserLang);
        }
    }
}
