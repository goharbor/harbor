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
import { Component, OnInit } from '@angular/core';
import { AppConfigService } from '../../services/app-config.service';
import { SkinableConfig } from '../../services/skinable-config.service';
import { TranslateService } from '@ngx-translate/core';
import { getDatetimeRendering } from 'src/app/shared/units/shared.utils';
import { registerLocaleData } from '@angular/common';
import { forkJoin, Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { ClrCommonStrings } from '@clr/angular/utils/i18n/common-strings.interface';
import { ClrCommonStringsService } from '@clr/angular';
import {
    DATETIME_RENDERINGS,
    DatetimeRendering,
    DEFAULT_DATETIME_RENDERING_LOCALSTORAGE_KEY,
    DEFAULT_LANG_LOCALSTORAGE_KEY,
    DefaultDatetimeRendering,
    DeFaultLang,
    LANGUAGES,
    stringsForClarity,
    SupportedLanguage,
} from '../../shared/entities/shared.const';

@Component({
    selector: 'preference-settings',
    templateUrl: 'preference-settings.component.html',
    styleUrls: ['preference-settings.component.scss'],
})
export class PreferenceSettingsComponent implements OnInit {
    readonly guiLanguages = Object.entries(LANGUAGES);
    readonly guiDatetimeRenderings = Object.entries(DATETIME_RENDERINGS);
    selectedLang: SupportedLanguage = DeFaultLang;
    selectedDatetimeRendering: DatetimeRendering = DefaultDatetimeRendering;
    opened: boolean = false;
    build: string = '4276418';
    customIntroduction: string;
    customName: string;
    customLogo: string;

    constructor(
        private appConfigService: AppConfigService,
        private skinableConfig: SkinableConfig,
        private translate: TranslateService,
        private commonStrings: ClrCommonStringsService
    ) {}

    ngOnInit(): void {
        // custom skin
        let customSkinObj = this.skinableConfig.getSkinConfig();
        if (customSkinObj) {
            if (customSkinObj.product) {
                this.customLogo = customSkinObj.product.logo;
                this.customName = customSkinObj.product.name;
                this.customIntroduction = customSkinObj.product.introduction;
            }
        }
        this.selectedLang = this.translate.currentLang as SupportedLanguage;
        if (this.selectedLang) {
            registerLocaleData(
                LANGUAGES[this.selectedLang][1],
                this.selectedLang
            );
            this.translateClarityComponents();
        }
        this.selectedDatetimeRendering = getDatetimeRendering();
    }

    //Internationalization for Clarity components, refer to https://clarity.design/documentation/internationalization
    translateClarityComponents() {
        const translatedObservables: Observable<string | any>[] = [];
        const translatedStringsForClarity: Partial<ClrCommonStrings> = {};
        for (let key in stringsForClarity) {
            translatedObservables.push(
                this.translate.get(stringsForClarity[key]).pipe(
                    map(res => {
                        return [key, res];
                    })
                )
            );
        }
        forkJoin(translatedObservables).subscribe(res => {
            if (res?.length) {
                res.forEach(item => {
                    translatedStringsForClarity[item[0]] = item[1];
                });
                this.commonStrings.localize(translatedStringsForClarity);
            }
        });
    }

    public get version(): string {
        let appConfig = this.appConfigService.getConfig();
        return appConfig.harbor_version;
    }

    public get currentLang(): string {
        if (this.selectedLang) {
            return LANGUAGES[this.selectedLang][0] as string;
        }
        return null;
    }

    public get currentDatetimeRendering(): string {
        return DATETIME_RENDERINGS[this.selectedDatetimeRendering];
    }

    matchLang(lang: SupportedLanguage): boolean {
        return lang === this.selectedLang;
    }

    matchDatetimeRendering(datetime: DatetimeRendering): boolean {
        return datetime === this.selectedDatetimeRendering;
    }

    // Switch languages
    switchLanguage(lang: SupportedLanguage): void {
        this.selectedLang = lang;
        localStorage.setItem(DEFAULT_LANG_LOCALSTORAGE_KEY, lang);
        // due to the bug(https://github.com/ngx-translate/core/issues/1258) of translate module
        // have to reload
        this.translate.use(lang).subscribe(() => window.location.reload());
    }

    switchDatetimeRendering(datetime: DatetimeRendering): void {
        this.selectedDatetimeRendering = datetime;
        localStorage.setItem(
            DEFAULT_DATETIME_RENDERING_LOCALSTORAGE_KEY,
            datetime
        );
        // have to reload,as HarborDatetimePipe is pure pipe
        window.location.reload();
    }

    public open(): void {
        this.opened = true;
    }

    public close(): void {
        this.opened = false;
    }
}
