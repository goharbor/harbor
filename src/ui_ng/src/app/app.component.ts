import { Component, ReflectiveInjector, LOCALE_ID } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { supportedLangs, enLang } from './shared/shared.const';
import { SessionService } from './shared/session.service';
import { NgXCookies } from 'ngx-cookies';


@Component({
    selector: 'harbor-app',
    templateUrl: 'app.component.html'
})
export class AppComponent {
    constructor(
        private translate: TranslateService,
        private session: SessionService) {

        translate.addLangs(supportedLangs);
        translate.setDefaultLang(enLang);

        //If user has selected lang, then directly use it
        let langSetting = NgXCookies.getCookie("harbor-lang");
        if (!langSetting || langSetting.trim() === "") {
            //Use browser lang
            langSetting = translate.getBrowserCultureLang().toLowerCase();
        }

        let selectedLang = this.isLangMatch(langSetting, supportedLangs) ? langSetting : enLang;
        translate.use(selectedLang);
        //this.session.switchLanguage(selectedLang).catch(error => console.error(error));
    }

    isLangMatch(browserLang: string, supportedLangs: string[]) {
        if (supportedLangs && supportedLangs.length > 0) {
            return supportedLangs.find(lang => lang === browserLang);
        }
    }
}
