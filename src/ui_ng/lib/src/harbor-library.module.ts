import { NgModule, ModuleWithProviders, Provider } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SYSTEMINFO_DIRECTIVES } from './system/index';
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

/**
 * Declare default service configuration
 */
export const DefaultServiceConfig: IServiceConfig = {
  systemInfoEndpoint: "/api/system",
  repositoryBaseEndpoint: "",
  logBaseEndpoint: "",
  targetBaseEndpoint: "",
  replicationRuleEndpoint: "",
  replicationJobEndpoint: ""
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

@NgModule({
  imports: [
    CommonModule
  ],
  declarations: [SYSTEMINFO_DIRECTIVES],
  exports: [SYSTEMINFO_DIRECTIVES]
})

export class HarborLibraryModule {
  static forRoot(config: HarborModuleConfig = {}): ModuleWithProviders {
    return {
      ngModule: HarborLibraryModule,
      providers: [
        config.config || { provide: SERVICE_CONFIG, useValue: DefaultServiceConfig },
        config.logService || { provide: AccessLogService, useClass: AccessLogDefaultService },
        config.endpointService || { provide: EndpointService, useClass: EndpointDefaultService },
        config.replicationService || { provide: ReplicationService, useClass: ReplicationDefaultService },
        config.repositoryService || { provide: RepositoryService, useClass: RepositoryDefaultService },
        config.tagService || { provide: TagService, useClass: TagDefaultService }
      ]
    };
  }

  static forChild(config: HarborModuleConfig = {}): ModuleWithProviders {
    return {
      ngModule: HarborLibraryModule,
      providers: [
        config.config || { provide: SERVICE_CONFIG, useValue: DefaultServiceConfig },
        config.logService || { provide: AccessLogService, useClass: AccessLogDefaultService },
        config.endpointService || { provide: EndpointService, useClass: EndpointDefaultService },
        config.replicationService || { provide: ReplicationService, useClass: ReplicationDefaultService },
        config.repositoryService || { provide: RepositoryService, useClass: RepositoryDefaultService },
        config.tagService || { provide: TagService, useClass: TagDefaultService }
      ]
    };
  }
}
