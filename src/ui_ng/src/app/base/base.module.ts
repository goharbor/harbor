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