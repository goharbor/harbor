export interface I18nConfig {
  /**
   * The cookie key used to store the current used language preference.
   *
   * * {string}
   * @memberOf IServiceConfig
   */
  langCookieKey?: string;

  /**
   * Declare what languages are supported.
   *
   * * {string[]}
   * @memberOf IServiceConfig
   */
  supportedLangs?: string[];

  /**
   * Define the default language the translate service uses.
   *
   * * {string}
   * @memberOf I18nConfig
   */
  defaultLang?: string;

  /**
   * To determine whether or not to enable the i18 multiple languages supporting.
   *
   * * {boolean}
   * @memberOf IServiceConfig
   */
  enablei18Support?: boolean;
}
