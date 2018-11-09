// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { NgModule } from '@angular/core';
import { CoreModule } from '../core/core.module';
import { SharedModule } from '../shared/shared.module';

import { ConfigurationComponent } from './config.component';
import { ConfigurationService } from './config.service';
import { ConfirmMessageHandler } from './config.msg.utils';
import { ConfigurationAuthComponent } from './auth/config-auth.component';
import { ConfigurationEmailComponent } from './email/config-email.component';
import { GcComponent } from './gc/gc.component';
import { GcRepoService } from './gc/gc.service';
import { GcApiRepository } from './gc/gc.api.repository';
import { GcViewModelFactory } from './gc/gc.viewmodel.factory';
import { GcUtility } from './gc/gc.utility';


@NgModule({
  imports: [
    CoreModule,
    SharedModule
  ],
  declarations: [
    ConfigurationComponent,
    ConfigurationAuthComponent,
    ConfigurationEmailComponent,
    GcComponent
  ],
  exports: [ConfigurationComponent],
  providers: [ConfigurationService, GcRepoService, GcApiRepository, GcViewModelFactory, GcUtility, ConfirmMessageHandler]
})
export class ConfigurationModule { }
