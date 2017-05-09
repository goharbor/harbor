import { NgModule, ModuleWithProviders, Provider, APP_INITIALIZER, Inject } from '@angular/core';

import { LOG_DIRECTIVES } from './log/index';
import { FILTER_DIRECTIVES } from './filter/index';
import { SERVICE_CONFIG, IServiceConfig } from './service.config';
import {
  AccessLogService,
  AccessLogDefaultService,
  EndpointService,
  EndpointDefaultService,
  ReplicationService,
  ReplicationDefaultService,
  RepositoryService,
  RepositoryDefaultService,
  TagService,
  TagDefaultService
} from './service/index';
import {
  ErrorHandler,
  DefaultErrorHandler
} from './error-handler/index';
import { SharedModule } from './shared/shared.module';
import { DEFAULT_LANG_COOKIE_KEY, DEFAULT_SUPPORTING_LANGS, DEFAULT_LANG } from './utils';
import { TranslateService } from '@ngx-translate/core';
import { CookieService } from 'ngx-cookie';

/**
 * Declare default service configuration; all the endpoints will be defined in
 * this default configuration.
 */
export const DefaultServiceConfig: IServiceConfig = {
  systemInfoEndpoint: "/api/system",
  repositoryBaseEndpoint: "",
  logBaseEndpoint: "/api/logs",
  targetBaseEndpoint: "",
  replicationRuleEndpoint: "",
  replicationJobEndpoint: "",
  langCookieKey: DEFAULT_LANG_COOKIE_KEY,
  supportedLangs: DEFAULT_SUPPORTING_LANGS,
  enablei18Support: false
};

/**
 * Define the configuration for harbor shareable module
 * 
 * @export
 * @interface HarborModuleConfig
 */
export interface HarborModuleConfig {
  //Service endpoints
  config?: Provider,

  //Handling error messages
  errorHandler?: Provider,

  //Service implementation for log
  logService?: Provider,

  //Service implementation for endpoint
  endpointService?: Provider,

  //Service implementation for replication
  replicationService?: Provider,

  //Service implementation for repository
  repositoryService?: Provider,

  //Service implementation for tag
  tagService?: Provider
}

/**
 * 
 * 
 * @export
 * @param {AppConfigService} configService
 * @returns
 */
export function initConfig(translateService: TranslateService, config: IServiceConfig, cookie: CookieService) {
  return (init);
  function init() {
    let selectedLang: string = DEFAULT_LANG;

    translateService.addLangs(config.supportedLangs ? config.supportedLangs : [DEFAULT_LANG]);
    translateService.setDefaultLang(DEFAULT_LANG);

    if (config.enablei18Support) {
      //If user has selected lang, then directly use it
      let langSetting: string = cookie.get(config.langCookieKey ? config.langCookieKey : DEFAULT_LANG_COOKIE_KEY);
      if (!langSetting || langSetting.trim() === "") {
        //Use browser lang
        langSetting = translateService.getBrowserCultureLang().toLowerCase();
      }

      if (config.supportedLangs && config.supportedLangs.length > 0) {
        if (config.supportedLangs.find(lang => lang === langSetting)) {
          selectedLang = langSetting;
        }
      }
    }

    translateService.use(selectedLang);
     console.log('initConfig => ', translateService.currentLang);
  };
}

@NgModule({
  imports: [
    SharedModule
  ],
  declarations: [
    LOG_DIRECTIVES,
    FILTER_DIRECTIVES
  ],
  exports: [
    LOG_DIRECTIVES,
    FILTER_DIRECTIVES
  ]
})

export class HarborLibraryModule {
  static forRoot(config: HarborModuleConfig = {}): ModuleWithProviders {
    return {
      ngModule: HarborLibraryModule,
      providers: [
        config.config || { provide: SERVICE_CONFIG, useValue: DefaultServiceConfig },
        config.errorHandler || { provide: ErrorHandler, useClass: DefaultErrorHandler },
        config.logService || { provide: AccessLogService, useClass: AccessLogDefaultService },
        config.endpointService || { provide: EndpointService, useClass: EndpointDefaultService },
        config.replicationService || { provide: ReplicationService, useClass: ReplicationDefaultService },
        config.repositoryService || { provide: RepositoryService, useClass: RepositoryDefaultService },
        config.tagService || { provide: TagService, useClass: TagDefaultService },
        //Do initializing
        TranslateService,
        {
          provide: APP_INITIALIZER,
          useFactory: initConfig,
          deps: [TranslateService, SERVICE_CONFIG],
          multi: true
        },
      ]
    };
  }

  static forChild(config: HarborModuleConfig = {}): ModuleWithProviders {
    return {
      ngModule: HarborLibraryModule,
      providers: [
        config.config || { provide: SERVICE_CONFIG, useValue: DefaultServiceConfig },
        config.errorHandler || { provide: ErrorHandler, useClass: DefaultErrorHandler },
        config.logService || { provide: AccessLogService, useClass: AccessLogDefaultService },
        config.endpointService || { provide: EndpointService, useClass: EndpointDefaultService },
        config.replicationService || { provide: ReplicationService, useClass: ReplicationDefaultService },
        config.repositoryService || { provide: RepositoryService, useClass: RepositoryDefaultService },
        config.tagService || { provide: TagService, useClass: TagDefaultService }
      ]
    };
  }
}
