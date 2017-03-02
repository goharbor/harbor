import { Component } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { CookieService } from 'angular2-cookie/core';

import { supportedLangs, enLang } from './shared/shared.const';
import { SessionService } from './shared/session.service';

@Component({
    selector: 'harbor-app',
    templateUrl: 'app.component.html',
    styleUrls: []
})
export class AppComponent {
    constructor(
        private translate: TranslateService,
        private cookie: CookieService,
        private session: SessionService) {
        translate.addLangs(supportedLangs);
        translate.setDefaultLang(enLang);

        //If user has selected lang, then directly use it
        let langSetting = this.cookie.get("harbor-lang");
        if (!langSetting || langSetting.trim() === "") {
            //Use browser lang
            langSetting = translate.getBrowserLang();
        }
        translate.use(this.isLangMatch(langSetting, supportedLangs) ? langSetting : enLang);
    }

    private isLangMatch(browserLang: string, supportedLangs: string[]) {
        if (supportedLangs && supportedLangs.length > 0) {
            return supportedLangs.find(lang => lang === browserLang);
        }
    }
}
