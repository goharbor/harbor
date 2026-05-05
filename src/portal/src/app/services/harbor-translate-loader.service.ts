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
import { Injectable } from '@angular/core';
import { TranslateLoader } from '@ngx-translate/core';
import { Observable } from 'rxjs';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

@Injectable({
    providedIn: 'root',
})
export class HarborTranslateLoaderService extends TranslateLoader {
    constructor(private http: HttpClient) {
        super();
    }
    getTranslation(lang: string): Observable<any> {
        const prefix: string = 'i18n/lang/';
        let suffix: string = '-lang.json';
        if (environment && environment.buildTimestamp) {
            suffix += `?buildTimeStamp=${environment.buildTimestamp}`;
        }
        return this.http.get(`${prefix}${lang}${suffix}`);
    }
}
