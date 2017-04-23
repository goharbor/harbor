import { NgModule, ModuleWithProviders, Provider } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SYSTEMINFO_DIRECTIVES } from './system/index';
import { SERVICE_CONFIG, IServiceConfig } from './service.config';

export const DefaultServiceConfig: IServiceConfig = {
  systemInfoEndpoint: "/api/system",
  repositoryBaseEndpoint: "",
  logBaseEndpoint: "",
  targetBaseEndpoint: "",
  replicationRuleEndpoint:"",
  replicationJobEndpoint: ""
};

export interface HarborModuleConfig {
  config?: Provider
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
        config.config || { provide: SERVICE_CONFIG, useValue: DefaultServiceConfig }
      ]
    };
  }

  static forChild(config: HarborModuleConfig = {}): ModuleWithProviders {
    return {
      ngModule: HarborLibraryModule,
      providers: [
        config.config || { provide: SERVICE_CONFIG, useValue: DefaultServiceConfig }
      ]
    };
  }
}
