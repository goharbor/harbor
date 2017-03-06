import { NgModule } from '@angular/core';
import { CoreModule } from '../core/core.module';
import { SharedModule } from '../shared/shared.module';

import { ConfigurationComponent } from './config.component';
import { ConfigurationService } from './config.service';
import { ConfigurationAuthComponent } from './auth/config-auth.component';
import { ConfigurationEmailComponent } from './email/config-email.component';

@NgModule({
  imports: [
    CoreModule,
    SharedModule
  ],
  declarations: [
    ConfigurationComponent,
    ConfigurationAuthComponent,
    ConfigurationEmailComponent],
  exports: [ConfigurationComponent],
  providers: [ConfigurationService]
})
export class ConfigurationModule { }