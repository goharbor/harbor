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
import { CookieService } from 'ngx-cookie';

import { SessionService } from './shared/session.service';
import { AppConfigService } from './services/app-config.service';
import { ThemeService } from './services/theme.service';
import { THEME_ARRAY, ThemeInterface } from './services/theme';
import { clone } from '../lib/utils/utils';

const HAS_STYLE_MODE: string = 'styleModeLocal';

@Component({
    selector: 'harbor-app',
    templateUrl: 'app.component.html'
})
export class AppComponent {
    themeArray: ThemeInterface[] = clone(THEME_ARRAY);

    styleMode: string = this.themeArray[0].showStyle;
    constructor(
        private translate: TranslateService,
        private cookie: CookieService,
        private session: SessionService,
        private appConfigService: AppConfigService,
        private titleService: Title,
        public theme: ThemeService

        ) {
        // Override page title
        let key: string = "APP_TITLE.HARBOR";
        if (this.appConfigService.isIntegrationMode()) {
            key = "APP_TITLE.REG";
        }

        translate.get(key).subscribe((res: string) => {
            this.titleService.setTitle(res);
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
}
