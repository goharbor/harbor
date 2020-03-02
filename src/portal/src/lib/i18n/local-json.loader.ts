import { TranslateLoader } from "@ngx-translate/core";
import { of ,  Observable} from "rxjs";
import { IServiceConfig } from "../entities/service.config";


/**
 * Declare a translation loader with local json object
 *
 **
 * class TranslatorJsonLoader
 * extends {TranslateLoader}
 */
export class TranslatorJsonLoader extends TranslateLoader {
  constructor(private config: IServiceConfig) {
    super();
  }

  getTranslation(lang: string): Observable<any> {
    let dict: any =
      this.config &&
      this.config.localI18nMessageVariableMap &&
      this.config.localI18nMessageVariableMap[lang]
        ? this.config.localI18nMessageVariableMap[lang]
        : {};
    return of(dict);
  }
}
