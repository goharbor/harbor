// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { SharedModule } from '../shared/shared.module';
import { RouterModule } from '@angular/router';

import { ProjectModule } from '../project/project.module';
import { UserModule } from '../user/user.module';
import { AccountModule } from '../account/account.module';
import { RepositoryModule } from '../repository/repository.module';

import { NavigatorComponent } from './navigator/navigator.component';
import { GlobalSearchComponent } from './global-search/global-search.component';
import { FooterComponent } from './footer/footer.component';
import { HarborShellComponent } from './harbor-shell/harbor-shell.component';
import { SearchResultComponent } from './global-search/search-result.component';
import { StartPageComponent } from './start-page/start.component';

import { SearchTriggerService } from './global-search/search-trigger.service';

@NgModule({
  imports: [
    SharedModule,
    ProjectModule,
    UserModule,
    AccountModule,
    RouterModule,
    RepositoryModule
  ],
  declarations: [
    NavigatorComponent,
    GlobalSearchComponent,
    FooterComponent,
    HarborShellComponent,
    SearchResultComponent,
    StartPageComponent
  ],
  exports: [ HarborShellComponent ],
  providers: [SearchTriggerService]
})
export class BaseModule {

}