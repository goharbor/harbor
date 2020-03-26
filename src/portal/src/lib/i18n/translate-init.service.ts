import { Injectable } from "@angular/core";
import { I18nConfig } from "./i18n-config";
import { TranslateService } from "@ngx-translate/core";
import { CookieService } from "ngx-cookie";
import { DEFAULT_LANG, DEFAULT_LANG_COOKIE_KEY, DEFAULT_SUPPORTING_LANGS } from "../utils/utils";

@Injectable()
export class TranslateServiceInitializer {
  constructor(
    private translateService: TranslateService,
    private cookie: CookieService
  ) {}

  public init(config: I18nConfig = {}): void {
    let selectedLang: string = config.defaultLang
      ? config.defaultLang
      : DEFAULT_LANG;
    let supportedLangs: string[] = config.supportedLangs
      ? config.supportedLangs
      : DEFAULT_SUPPORTING_LANGS;

    this.translateService.addLangs(supportedLangs);
    this.translateService.setDefaultLang(selectedLang);

    if (config.enablei18Support) {
      // If user has selected lang, then directly use it
      let langSetting: string = this.cookie.get(
        config.langCookieKey ? config.langCookieKey : DEFAULT_LANG_COOKIE_KEY
      );
      if (!langSetting || langSetting.trim() === "") {
        // Use browser lang
        langSetting = this.translateService
          .getBrowserCultureLang()
          .toLowerCase();
      }

      if (langSetting && langSetting.trim() !== "") {
        if (supportedLangs && supportedLangs.length > 0) {
          if (supportedLangs.find(lang => lang === langSetting)) {
            selectedLang = langSetting;
          }
        }
      }
    }

    this.translateService.use(selectedLang);
  }
}
