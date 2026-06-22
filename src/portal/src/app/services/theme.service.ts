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
import { Injectable, Inject, DOCUMENT } from '@angular/core';

import { environment } from '../../environments/environment';

@Injectable({
    providedIn: 'root',
})
export class ThemeService {
    constructor(@Inject(DOCUMENT) private document: Document) {}

    loadStyle(styleName: string) {
        // Clarity 17+ / CDS: set before stylesheet swap so tokens apply with first paint
        const cdsTheme = styleName.includes('dark-theme') ? 'dark' : 'light';
        this.document.body?.setAttribute('cds-theme', cdsTheme);

        const head = this.document.getElementsByTagName('head')[0];

        let themeLink = this.document.getElementById(
            'client-theme'
        ) as HTMLLinkElement;
        if (themeLink) {
            themeLink.href = `${styleName}?buildTimeStamp=${environment.buildTimestamp}`;
        } else {
            const style = this.document.createElement('link');
            style.id = 'client-theme';
            style.rel = 'stylesheet';
            style.href = `${styleName}?buildTimeStamp=${environment.buildTimestamp}`;

            head.appendChild(style);
        }
    }
}
