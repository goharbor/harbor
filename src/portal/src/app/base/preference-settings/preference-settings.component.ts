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
import { Component, OnInit, ViewChild } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import {
    getContainerRuntime,
    getCustomContainerRuntime,
    getDatetimeRendering,
} from 'src/app/shared/units/shared.utils';
import { registerLocaleData } from '@angular/common';
import { forkJoin, Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { ClrCommonStrings } from '@clr/angular/utils/i18n/common-strings.interface';
import { ClrCommonStringsService } from '@clr/angular';
import {
    CUSTOM_RUNTIME_LOCALSTORAGE_KEY,
    DATETIME_RENDERINGS,
    DatetimeRendering,
    DEFAULT_DATETIME_RENDERING_LOCALSTORAGE_KEY,
    DEFAULT_LANG_LOCALSTORAGE_KEY,
    DEFAULT_RUNTIME_LOCALSTORAGE_KEY,
    DefaultDatetimeRendering,
    DeFaultLang,
    DeFaultRuntime,
    LANGUAGES,
    RUNTIMES,
    stringsForClarity,
    SupportedLanguage,
    SupportedRuntime,
} from '../../shared/entities/shared.const';
import { NgForm } from '@angular/forms';
import { InlineAlertComponent } from 'src/app/shared/components/inline-alert/inline-alert.component';

@Component({
    selector: 'preference-settings',
    templateUrl: 'preference-settings.component.html',
    styleUrls: ['preference-settings.component.scss'],
})
export class PreferenceSettingsComponent implements OnInit {
    readonly guiLanguages = Object.entries(LANGUAGES);
    readonly guiRuntimes = Object.entries(RUNTIMES).filter(
        ([_, value]) => value !== RUNTIMES.custom
    );
    readonly guiDatetimeRenderings = Object.entries(DATETIME_RENDERINGS);
    selectedLang: SupportedLanguage = DeFaultLang;
    selectedRuntime: SupportedRuntime = DeFaultRuntime;
    selectedDatetimeRendering: DatetimeRendering = DefaultDatetimeRendering;
    opened: boolean = false;
    error: any = null;
    customRuntime: string = '';

    @ViewChild('customruntimeForm', { static: false })
    customRuntimeForm: NgForm;
    @ViewChild(InlineAlertComponent, { static: true })
    inlineAlert: InlineAlertComponent;

    constructor(
        private translate: TranslateService,
        private commonStrings: ClrCommonStringsService
    ) {}

    ngOnInit(): void {
        this.selectedLang = this.translate.currentLang as SupportedLanguage;
        if (this.selectedLang) {
            registerLocaleData(
                LANGUAGES[this.selectedLang][1],
                this.selectedLang
            );
            this.translateClarityComponents();
        }
        this.selectedDatetimeRendering = getDatetimeRendering();
        this.selectedRuntime = getContainerRuntime();
        this.customRuntime = getCustomContainerRuntime();
    }

    // Check If form is valid
    public get isValid(): boolean {
        const customPrefixControl =
            this.customRuntimeForm?.form.get('customPrefix');

        return (
            customPrefixControl?.valid &&
            customPrefixControl?.value?.trim() !== '' &&
            this.error === null
        );
    }

    addCustomRuntime() {
        if (this.customRuntime.trim()) {
            const customRuntimeValue = this.customRuntime.trim();
            this.switchRuntime('custom');
            this.switchCustomRuntime(customRuntimeValue);
        }
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

    public get currentRuntime(): string {
        if (this.selectedRuntime) {
            return RUNTIMES[this.selectedRuntime] as string;
        }
        return null;
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

    matchRuntime(runtime: SupportedRuntime): boolean {
        return runtime === this.selectedRuntime;
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

    switchRuntime(runtime: SupportedRuntime): void {
        this.selectedRuntime = runtime;
        localStorage.setItem(DEFAULT_RUNTIME_LOCALSTORAGE_KEY, runtime);
    }

    switchCustomRuntime(runtime: SupportedRuntime): void {
        localStorage.setItem(CUSTOM_RUNTIME_LOCALSTORAGE_KEY, runtime);
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
