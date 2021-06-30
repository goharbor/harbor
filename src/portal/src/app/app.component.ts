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
import { Component } from '@angular/core';
import { Title } from '@angular/platform-browser';
import { TranslateService } from '@ngx-translate/core';
import { AppConfigService } from './services/app-config.service';
import { ThemeService } from './services/theme.service';
import { CustomStyle, HAS_STYLE_MODE, THEME_ARRAY, ThemeInterface } from './services/theme';
import { clone } from './shared/units/utils';
import { DEFAULT_LANG_LOCALSTORAGE_KEY, DeFaultLang, supportedLangs } from "./shared/entities/shared.const";
import { forkJoin, Observable } from "rxjs";
import { SkinableConfig } from "./services/skinable-config.service";



@Component({
    selector: 'harbor-app',
    templateUrl: 'app.component.html'
})
export class AppComponent {
    themeArray: ThemeInterface[] = clone(THEME_ARRAY);
    styleMode: string = this.themeArray[0].showStyle;
    constructor(
        private translate: TranslateService,
        private appConfigService: AppConfigService,
        private titleService: Title,
        public theme: ThemeService,
        private skinableConfig: SkinableConfig

        ) {
         // init language
        this.initLanguage();
        // Override page title
        let key: string = "APP_TITLE.HARBOR";
        if (this.appConfigService.isIntegrationMode()) {
            key = "APP_TITLE.REG";
        }

        translate.get(key).subscribe((res: string) => {
            const customSkinData: CustomStyle = this.skinableConfig.getSkinConfig();
            if (customSkinData && customSkinData.product && customSkinData.product.name) {
                this.titleService.setTitle(customSkinData.product.name);
                this.skinableConfig.setTitleIcon();
            } else {
                this.titleService.setTitle(res);
            }
        });
        this.setTheme();
    }
    setTheme () {
        let styleMode = this.themeArray[0].showStyle;
        const localHasStyle = localStorage && localStorage.getItem(HAS_STYLE_MODE);
        if (localHasStyle) {
            styleMode = localStorage.getItem(HAS_STYLE_MODE);
        } else {
            localStorage.setItem(HAS_STYLE_MODE, styleMode);
        }
        this.themeArray.forEach((themeItem) => {
            if (themeItem.showStyle === styleMode) {
                this.theme.loadStyle(themeItem.currentFileName);
            }
        });
    }
    initLanguage() {
        this.translate.addLangs(supportedLangs);
        this.translate.setDefaultLang(DeFaultLang);
        let selectedLang: string = DeFaultLang;
        if (localStorage && localStorage.getItem(DEFAULT_LANG_LOCALSTORAGE_KEY)) {// If user has selected lang, then directly use it
            selectedLang = localStorage.getItem(DEFAULT_LANG_LOCALSTORAGE_KEY);
        } else {// If user has not selected lang, then use browser language(if contained in supportedLangs)
            const browserCultureLang: string = this.translate
                .getBrowserCultureLang()
                .toLowerCase();
            if (browserCultureLang && browserCultureLang.trim() !== "") {
                if (supportedLangs && supportedLangs.length > 0) {
                    if (supportedLangs.find(lang => lang === browserCultureLang)) {
                        selectedLang = browserCultureLang;
                    }
                }
            }
        }
        localStorage.setItem(DEFAULT_LANG_LOCALSTORAGE_KEY, selectedLang);
        // use method will load related language json from backend server
        this.translate.use(selectedLang);
    }
}
