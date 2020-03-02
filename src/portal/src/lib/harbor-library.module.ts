import { NgModule, ModuleWithProviders, Provider, APP_INITIALIZER } from '@angular/core';
import { SERVICE_CONFIG, IServiceConfig } from './entities/service.config';

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
  UserPermissionDefaultService,
} from './services';
import { GcRepoService } from './components/config/gc/gc.service';
import { ScanAllRepoService } from './components/config/vulnerability/scanAll.service';
import {GcViewModelFactory} from './components/config/gc/gc.viewmodel.factory';
import {GcApiRepository, GcApiDefaultRepository} from './components/config/gc/gc.api.repository';
import {ScanApiRepository, ScanApiDefaultRepository} from './components/config/vulnerability/scanAll.api.repository';
import {
  ErrorHandler,
  DefaultErrorHandler
} from './utils/error-handler';
import { DEFAULT_LANG_COOKIE_KEY, DEFAULT_SUPPORTING_LANGS, DEFAULT_LANG, CURRENT_BASE_HREF } from './utils/utils';
import { OperationService } from './components/operation/operation.service';
import { GcHistoryComponent } from "./components/config/gc/gc-history/gc-history.component";
import { GcComponent } from "./components/config/gc/gc.component";
import { EditProjectQuotasComponent } from "./components/config/project-quotas/edit-project-quotas/edit-project-quotas.component";
import { ProjectQuotasComponent } from "./components/config/project-quotas/project-quotas.component";
import { ReplicationConfigComponent } from "./components/config/replication/replication-config.component";
import { SystemSettingsComponent } from "./components/config/system/system-settings.component";
import { VulnerabilityConfigComponent } from "./components/config/vulnerability/vulnerability-config.component";
import { RegistryConfigComponent } from "./components/config/registry-config.component";
import { ConfirmationDialogComponent } from "./components/confirmation-dialog";
import { CreateEditEndpointComponent } from "./components/create-edit-endpoint/create-edit-endpoint.component";
import { CreateEditLabelComponent } from "./components/create-edit-label/create-edit-label.component";
import { CreateEditRuleComponent } from "./components/create-edit-rule/create-edit-rule.component";
import { FilterLabelComponent } from "./components/create-edit-rule/filter-label.component";
import { CronScheduleComponent, CronTooltipComponent } from "./components/cron-schedule";
import { DateValidatorDirective } from "./components/datetime-picker/date-validator.directive";
import { DatePickerComponent } from "./components/datetime-picker/datetime-picker.component";
import { EndpointComponent } from "./components/endpoint/endpoint.component";
import { ImageNameInputComponent } from "./components/image-name-input/image-name-input.component";
import { InlineAlertComponent } from "./components/inline-alert/inline-alert.component";
import { LabelSignPostComponent } from "./components/label/label-signpost/label-signpost.component";
import { LabelComponent } from "./components/label/label.component";
import { LabelPieceComponent } from "./components/label-piece/label-piece.component";
import { RecentLogComponent } from "./components/log/recent-log.component";
import { OperationComponent } from "./components/operation/operation.component";
import { ProjectPolicyConfigComponent } from "./components/project-policy-config/project-policy-config.component";
import { CopyInputComponent } from "./components/push-image/copy-input.component";
import { PushImageButtonComponent } from "./components/push-image/push-image.component";
import { ReplicationTasksComponent } from "./components/replication/replication-tasks/replication-tasks.component";
import { ReplicationComponent } from "./components/replication/replication.component";
import { FilterComponent } from "./components/filter/filter.component";
import { ListReplicationRuleComponent } from "./components/list-replication-rule/list-replication-rule.component";
import { ChannelService } from "./services/channel.service";
import { SharedModule } from "./utils/shared/shared.module";
import { TranslateServiceInitializer } from "./i18n";

/**
 * Declare default service configuration; all the endpoints will be defined in
 * this default configuration.
 */
export const DefaultServiceConfig: IServiceConfig = {
  baseEndpoint: CURRENT_BASE_HREF,
  systemInfoEndpoint: CURRENT_BASE_HREF + "/systeminfo",
  repositoryBaseEndpoint: CURRENT_BASE_HREF + "/repositories",
  logBaseEndpoint: CURRENT_BASE_HREF + "/logs",
  targetBaseEndpoint: CURRENT_BASE_HREF + "/registries",
  replicationBaseEndpoint: CURRENT_BASE_HREF + "/replication",
  replicationRuleEndpoint: CURRENT_BASE_HREF + "/replication/policies",
  vulnerabilityScanningBaseEndpoint: CURRENT_BASE_HREF + "/repositories",
  projectPolicyEndpoint: CURRENT_BASE_HREF + "/projects/configs",
  projectBaseEndpoint: CURRENT_BASE_HREF + "/projects",
  enablei18Support: false,
  langCookieKey: DEFAULT_LANG_COOKIE_KEY,
  supportedLangs: DEFAULT_SUPPORTING_LANGS,
  defaultLang: DEFAULT_LANG,
  langMessageLoader: "local",
  langMessagePathForHttpLoader: "i18n/langs/",
  langMessageFileSuffixForHttpLoader: "-lang.json",
  localI18nMessageVariableMap: {},
  configurationEndpoint: CURRENT_BASE_HREF + "/configurations",
  scanJobEndpoint: CURRENT_BASE_HREF + "/jobs/scan",
  labelEndpoint: CURRENT_BASE_HREF + "/labels",
  helmChartEndpoint: CURRENT_BASE_HREF + "/chartrepo",
  downloadChartEndpoint: "/chartrepo",
  gcEndpoint: CURRENT_BASE_HREF + "/system/gc",
  ScanAllEndpoint: CURRENT_BASE_HREF + "/system/scanAll"
};

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
  artifactService?: Provider;

  // Service implementation for gc
  gcApiRepository?: Provider;

  // Service implementation for scanAll
  ScanApiRepository?: Provider;

}


@NgModule({
    imports: [
        SharedModule,
    ],
    declarations: [
      GcHistoryComponent,
      GcComponent,
      EditProjectQuotasComponent,
      ProjectQuotasComponent,
      ReplicationConfigComponent,
      SystemSettingsComponent,
      VulnerabilityConfigComponent,
      RegistryConfigComponent,
      ConfirmationDialogComponent,
      CreateEditEndpointComponent,
      CreateEditLabelComponent,
      CreateEditRuleComponent,
      FilterLabelComponent,
      CronScheduleComponent,
      CronTooltipComponent,
      DateValidatorDirective,
      DatePickerComponent,
      EndpointComponent,
      FilterComponent,
      ImageNameInputComponent,
      InlineAlertComponent,
      LabelSignPostComponent,
      LabelComponent,
      LabelPieceComponent,
      ListReplicationRuleComponent,
      RecentLogComponent,
      OperationComponent,
      ProjectPolicyConfigComponent,
      CopyInputComponent,
      PushImageButtonComponent,
      ReplicationTasksComponent,
      ReplicationComponent,
  ],
  exports: [
      SharedModule,
      GcHistoryComponent,
      GcComponent,
      EditProjectQuotasComponent,
      ProjectQuotasComponent,
      ReplicationConfigComponent,
      SystemSettingsComponent,
      VulnerabilityConfigComponent,
      RegistryConfigComponent,
      ConfirmationDialogComponent,
      CreateEditEndpointComponent,
      CreateEditLabelComponent,
      CreateEditRuleComponent,
      FilterLabelComponent,
      CronScheduleComponent,
      CronTooltipComponent,
      DateValidatorDirective,
      DatePickerComponent,
      EndpointComponent,
      FilterComponent,
      ImageNameInputComponent,
      InlineAlertComponent,
      LabelSignPostComponent,
      LabelComponent,
      LabelPieceComponent,
      ListReplicationRuleComponent,
      RecentLogComponent,
      OperationComponent,
      ProjectPolicyConfigComponent,
      CopyInputComponent,
      PushImageButtonComponent,
      ReplicationTasksComponent,
      ReplicationComponent,
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
