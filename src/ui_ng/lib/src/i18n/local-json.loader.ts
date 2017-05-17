import { TranslateLoader } from '@ngx-translate/core';
import 'rxjs/add/observable/of';

import { Observable } from 'rxjs/Observable';
import { EN_US_LANG } from './lang/en-us-lang';
import { ES_ES_LANG } from './lang/es-es-lang';
import { ZH_CN_LANG } from './lang/zh-cn-lang';


/**
 * Define language mapping
 */
export const langs: { [key: string]: any } = {
    "en-us": EN_US_LANG,
    "es-es": ES_ES_LANG,
    "zh-cn": ZH_CN_LANG
};

/**
 * Declare a translation loader with local json object
 * 
 * @export
 * @class TranslatorJsonLoader
 * @extends {TranslateLoader}
 */
export class TranslatorJsonLoader extends TranslateLoader {
    getTranslation(lang: string): Observable<any> {
        let dict: any = langs[lang] ? langs[lang] : {};
        return Observable.of(dict);
    }
}