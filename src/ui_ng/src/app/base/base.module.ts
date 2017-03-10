import { NgModule } from '@angular/core';
import { SharedModule } from '../shared/shared.module';
import { RouterModule } from '@angular/router';

import { ProjectModule } from '../project/project.module';
import { UserModule } from '../user/user.module';
import { AccountModule } from '../account/account.module';

import { NavigatorComponent } from './navigator/navigator.component';
import { GlobalSearchComponent } from './global-search/global-search.component';
import { FooterComponent } from './footer/footer.component';
import { HarborShellComponent } from './harbor-shell/harbor-shell.component';
import { SearchResultComponent } from './global-search/search-result.component';
import { SearchStartComponent } from './global-search/search-start.component';

import { SearchTriggerService } from './global-search/search-trigger.service';

@NgModule({
  imports: [
    SharedModule,
    ProjectModule,
    UserModule,
    AccountModule,
    RouterModule
  ],
  declarations: [
    NavigatorComponent,
    GlobalSearchComponent,
    FooterComponent,
    HarborShellComponent,
    SearchResultComponent,
    SearchStartComponent
  ],
  exports: [ HarborShellComponent ],
  providers: [SearchTriggerService]
})
export class BaseModule {

}