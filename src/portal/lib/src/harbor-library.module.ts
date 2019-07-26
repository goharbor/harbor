import { NgModule, ModuleWithProviders, Provider, APP_INITIALIZER } from '@angular/core';

import { LOG_DIRECTIVES } from './log/index';
import { FILTER_DIRECTIVES } from './filter/index';
import { ENDPOINT_DIRECTIVES } from './endpoint/index';
import { REPOSITORY_DIRECTIVES } from './repository/index';
import { TAG_DIRECTIVES } from './tag/index';

import { REPLICATION_DIRECTIVES } from './replication/index';
import { CREATE_EDIT_RULE_DIRECTIVES } from './create-edit-rule/index';
import { LIST_REPLICATION_RULE_DIRECTIVES } from './list-replication-rule/index';

import { CREATE_EDIT_ENDPOINT_DIRECTIVES } from './create-edit-endpoint/index';

import { SERVICE_CONFIG, IServiceConfig } from './service.config';

import { CONFIRMATION_DIALOG_DIRECTIVES } from './confirmation-dialog/index';
import { INLINE_ALERT_DIRECTIVES } from './inline-alert/index';
import { DATETIME_PICKER_DIRECTIVES } from './datetime-picker/index';
import { VULNERABILITY_DIRECTIVES } from './vulnerability-scanning/index';
import { PUSH_IMAGE_BUTTON_DIRECTIVES } from './push-image/index';
import { CONFIGURATION_DIRECTIVES } from './config/index';
import { PROJECT_POLICY_CONFIG_DIRECTIVES } from './project-policy-config/index';
import { HBR_GRIDVIEW_DIRECTIVES } from './gridview/index';
import { REPOSITORY_GRIDVIEW_DIRECTIVES } from './repository-gridview/index';
import { OPERATION_DIRECTIVES } from './operation/index';
import { LABEL_DIRECTIVES } from "./label/index";
import { CREATE_EDIT_LABEL_DIRECTIVES } from "./create-edit-label/index";
import { LABEL_PIECE_DIRECTIVES } from "./label-piece/index";
import { IMAGE_NAME_INPUT_DIRECTIVES } from "./image-name-input/index";
import { CRON_SCHEDULE_DIRECTIVES } from "./cron-schedule/index";
import {
  SystemInfoService,
  SystemInfoDefaultService,
  AccessLogService,
  AccessLogDefaultService,
  EndpointService,
  EndpointDefaultService,
  ReplicationService,
  ReplicationDefaultService,
  QuotaService,
  QuotaDefaultService,
  RepositoryService,
  RepositoryDefaultService,
  TagService,
  TagDefaultService,
  ScanningResultService,
  ScanningResultDefaultService,
  ConfigurationService,
  ConfigurationDefaultService,
  JobLogService,
  JobLogDefaultService,
  ProjectService,
  ProjectDefaultService,
  LabelService,
  LabelDefaultService,
  RetagService,
  RetagDefaultService,
  UserPermissionService,
  UserPermissionDefaultService
} from './service/index';
import { GcRepoService } from './config/gc/gc.service';
import { ScanAllRepoService } from './config/vulnerability/scanAll.service';
import {GcViewModelFactory} from './config/gc/gc.viewmodel.factory';
import {GcApiRepository, GcApiDefaultRepository} from './config/gc/gc.api.repository';
import {ScanApiRepository, ScanApiDefaultRepository} from './config/vulnerability/scanAll.api.repository';
import {
  ErrorHandler,
  DefaultErrorHandler
} from './error-handler/index';
import { SharedModule } from './shared/shared.module';
import { TranslateModule } from '@ngx-translate/core';

import { TranslateServiceInitializer } from './i18n/index';
import { DEFAULT_LANG_COOKIE_KEY, DEFAULT_SUPPORTING_LANGS, DEFAULT_LANG } from './utils';
import { ChannelService } from './channel/index';
import { OperationService } from './operation/operation.service';

/**
 * Declare default service configuration; all the endpoints will be defined in
 * this default configuration.
 */
export const DefaultServiceConfig: IServiceConfig = {
  baseEndpoint: "/api",
  systemInfoEndpoint: "/api/systeminfo",
  repositoryBaseEndpoint: "/api/repositories",
  logBaseEndpoint: "/api/logs",
  targetBaseEndpoint: "/api/registries",
  replicationBaseEndpoint: "/api/replication",
  replicationRuleEndpoint: "/api/replication/policies",
  vulnerabilityScanningBaseEndpoint: "/api/repositories",
  projectPolicyEndpoint: "/api/projects/configs",
  projectBaseEndpoint: "/api/projects",
  enablei18Support: false,
  langCookieKey: DEFAULT_LANG_COOKIE_KEY,
  supportedLangs: DEFAULT_SUPPORTING_LANGS,
  defaultLang: DEFAULT_LANG,
  langMessageLoader: "local",
  langMessagePathForHttpLoader: "i18n/langs/",
  langMessageFileSuffixForHttpLoader: "-lang.json",
  localI18nMessageVariableMap: {},
  configurationEndpoint: "/api/configurations",
  scanJobEndpoint: "/api/jobs/scan",
  labelEndpoint: "/api/labels",
  helmChartEndpoint: "/api/chartrepo",
  downloadChartEndpoint: "/chartrepo",
  gcEndpoint: "/api/system/gc",
  ScanAllEndpoint: "/api/system/scanAll"
};

/**
 * Define the configuration for harbor shareable module
 *
 **
 * interface HarborModuleConfig
 */
export interface HarborModuleConfig {
  // Service endpoints
  config?: Provider;

  // Handling error messages
  errorHandler?: Provider;

  // Service implementation for system info
  systemInfoService?: Provider;

  // Service implementation for log
  logService?: Provider;

  // Service implementation for endpoint
  endpointService?: Provider;

  // Service implementation for replication
  replicationService?: Provider;

  // Service implementation for replication
  QuotaService?: Provider;

  // Service implementation for repository
  repositoryService?: Provider;

  // Service implementation for tag
  tagService?: Provider;

  // Service implementation for retag
  retagService?: Provider;

  // Service implementation for vulnerability scanning
  scanningService?: Provider;

  // Service implementation for configuration
  configService?: Provider;

  // Service implementation for job log
  jobLogService?: Provider;

  // Service implementation for project policy
  projectPolicyService?: Provider;

  // Service implementation for label
  labelService?: Provider;

  // Service implementation for helmchart
  helmChartService?: Provider;
  // Service implementation for userPermission
  userPermissionService?: Provider;

  // Service implementation for gc
  gcApiRepository?: Provider;

  // Service implementation for scanAll
  ScanApiRepository?: Provider;

}

/**
 **
 *  ** deprecated param {AppConfigService} configService
 * returns
 */
export function initConfig(translateInitializer: TranslateServiceInitializer, config: IServiceConfig) {
  return (init);
  function init() {
    translateInitializer.init({
      enablei18Support: config.enablei18Support,
      supportedLangs: config.supportedLangs,
      defaultLang: config.defaultLang,
      langCookieKey: config.langCookieKey
    });
  }
}

@NgModule({
  imports: [
    SharedModule
  ],
  declarations: [
    LOG_DIRECTIVES,
    FILTER_DIRECTIVES,
    ENDPOINT_DIRECTIVES,
    REPOSITORY_DIRECTIVES,
    TAG_DIRECTIVES,
    CREATE_EDIT_ENDPOINT_DIRECTIVES,
    CONFIRMATION_DIALOG_DIRECTIVES,
    INLINE_ALERT_DIRECTIVES,
    REPLICATION_DIRECTIVES,
    LIST_REPLICATION_RULE_DIRECTIVES,
    CREATE_EDIT_RULE_DIRECTIVES,
    DATETIME_PICKER_DIRECTIVES,
    VULNERABILITY_DIRECTIVES,
    PUSH_IMAGE_BUTTON_DIRECTIVES,
    CONFIGURATION_DIRECTIVES,
    PROJECT_POLICY_CONFIG_DIRECTIVES,
    LABEL_DIRECTIVES,
    CREATE_EDIT_LABEL_DIRECTIVES,
    LABEL_PIECE_DIRECTIVES,
    HBR_GRIDVIEW_DIRECTIVES,
    REPOSITORY_GRIDVIEW_DIRECTIVES,
    OPERATION_DIRECTIVES,
    IMAGE_NAME_INPUT_DIRECTIVES,
    CRON_SCHEDULE_DIRECTIVES
  ],
  exports: [
    LOG_DIRECTIVES,
    FILTER_DIRECTIVES,
    ENDPOINT_DIRECTIVES,
    REPOSITORY_DIRECTIVES,
    TAG_DIRECTIVES,
    CREATE_EDIT_ENDPOINT_DIRECTIVES,
    CONFIRMATION_DIALOG_DIRECTIVES,
    INLINE_ALERT_DIRECTIVES,
    REPLICATION_DIRECTIVES,
    LIST_REPLICATION_RULE_DIRECTIVES,
    CREATE_EDIT_RULE_DIRECTIVES,
    DATETIME_PICKER_DIRECTIVES,
    VULNERABILITY_DIRECTIVES,
    PUSH_IMAGE_BUTTON_DIRECTIVES,
    CONFIGURATION_DIRECTIVES,
    TranslateModule,
    PROJECT_POLICY_CONFIG_DIRECTIVES,
    LABEL_DIRECTIVES,
    CREATE_EDIT_LABEL_DIRECTIVES,
    LABEL_PIECE_DIRECTIVES,
    HBR_GRIDVIEW_DIRECTIVES,
    REPOSITORY_GRIDVIEW_DIRECTIVES,
    OPERATION_DIRECTIVES,
    IMAGE_NAME_INPUT_DIRECTIVES,
    CRON_SCHEDULE_DIRECTIVES,
    SharedModule
  ],
  providers: []
})

export class HarborLibraryModule {
  static forRoot(config: HarborModuleConfig = {}): ModuleWithProviders {
    return {
      ngModule: HarborLibraryModule,
      providers: [
        config.config || { provide: SERVICE_CONFIG, useValue: DefaultServiceConfig },
        config.errorHandler || { provide: ErrorHandler, useClass: DefaultErrorHandler },
        config.systemInfoService || { provide: SystemInfoService, useClass: SystemInfoDefaultService },
        config.logService || { provide: AccessLogService, useClass: AccessLogDefaultService },
        config.endpointService || { provide: EndpointService, useClass: EndpointDefaultService },
        config.replicationService || { provide: ReplicationService, useClass: ReplicationDefaultService },
        config.QuotaService || { provide: QuotaService, useClass: QuotaDefaultService },
        config.repositoryService || { provide: RepositoryService, useClass: RepositoryDefaultService },
        config.tagService || { provide: TagService, useClass: TagDefaultService },
        config.retagService || { provide: RetagService, useClass: RetagDefaultService },
        config.scanningService || { provide: ScanningResultService, useClass: ScanningResultDefaultService },
        config.configService || { provide: ConfigurationService, useClass: ConfigurationDefaultService },
        config.jobLogService || { provide: JobLogService, useClass: JobLogDefaultService },
        config.projectPolicyService || { provide: ProjectService, useClass: ProjectDefaultService },
        config.labelService || { provide: LabelService, useClass: LabelDefaultService },
        config.userPermissionService || { provide: UserPermissionService, useClass: UserPermissionDefaultService },
        config.gcApiRepository || {provide: GcApiRepository, useClass: GcApiDefaultRepository},
        config.ScanApiRepository || {provide: ScanApiRepository, useClass: ScanApiDefaultRepository},
        // Do initializing
        TranslateServiceInitializer,
        {
          provide: APP_INITIALIZER,
          useFactory: initConfig,
          deps: [TranslateServiceInitializer, SERVICE_CONFIG],
          multi: true
        },
        ChannelService,
        OperationService,
        GcRepoService,
        ScanAllRepoService,
        GcViewModelFactory
      ]
    };
  }

  static forChild(config: HarborModuleConfig = {}): ModuleWithProviders {
    return {
      ngModule: HarborLibraryModule,
      providers: [
        config.config || { provide: SERVICE_CONFIG, useValue: DefaultServiceConfig },
        config.errorHandler || { provide: ErrorHandler, useClass: DefaultErrorHandler },
        config.systemInfoService || { provide: SystemInfoService, useClass: SystemInfoDefaultService },
        config.logService || { provide: AccessLogService, useClass: AccessLogDefaultService },
        config.endpointService || { provide: EndpointService, useClass: EndpointDefaultService },
        config.replicationService || { provide: ReplicationService, useClass: ReplicationDefaultService },
        config.QuotaService || { provide: QuotaService, useClass: QuotaDefaultService },
        config.repositoryService || { provide: RepositoryService, useClass: RepositoryDefaultService },
        config.tagService || { provide: TagService, useClass: TagDefaultService },
        config.retagService || { provide: RetagService, useClass: RetagDefaultService },
        config.scanningService || { provide: ScanningResultService, useClass: ScanningResultDefaultService },
        config.configService || { provide: ConfigurationService, useClass: ConfigurationDefaultService },
        config.jobLogService || { provide: JobLogService, useClass: JobLogDefaultService },
        config.projectPolicyService || { provide: ProjectService, useClass: ProjectDefaultService },
        config.labelService || { provide: LabelService, useClass: LabelDefaultService },
        config.userPermissionService || { provide: UserPermissionService, useClass: UserPermissionDefaultService },
        config.gcApiRepository || {provide: GcApiRepository, useClass: GcApiDefaultRepository},
        config.ScanApiRepository || {provide: ScanApiRepository, useClass: ScanApiDefaultRepository},
        ChannelService,
        OperationService,
        GcRepoService,
        ScanAllRepoService,
        GcViewModelFactory
      ]
    };
  }
}
