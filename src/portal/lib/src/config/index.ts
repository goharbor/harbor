import { Type } from '@angular/core';

import { ReplicationConfigComponent } from './replication/replication-config.component';
import { SystemSettingsComponent } from './system/system-settings.component';
import { VulnerabilityConfigComponent } from './vulnerability/vulnerability-config.component';
import { RegistryConfigComponent } from './registry-config.component';
import { GcComponent } from './gc/gc.component';


export * from './config';
export * from './replication/replication-config.component';
export * from './system/system-settings.component';
export * from './vulnerability/vulnerability-config.component';
export * from './registry-config.component';
export * from './gc/index';

export const CONFIGURATION_DIRECTIVES: Type<any>[] = [
  ReplicationConfigComponent,
  GcComponent,
  SystemSettingsComponent,
  VulnerabilityConfigComponent,
  RegistryConfigComponent
];
