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
