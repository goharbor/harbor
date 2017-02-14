import { NgModule } from '@angular/core';
import { SharedModule } from '../shared/shared.module';

import { DashboardModule } from '../dashboard/dashboard.module';
import { ProjectModule } from '../project/project.module';
import { UserModule } from '../user/user.module';

import { NavigatorComponent } from './navigator/navigator.component';
import { GlobalSearchComponent } from './global-search/global-search.component';
import { FooterComponent } from './footer/footer.component';
import { HarborShellComponent } from './harbor-shell/harbor-shell.component';

import { BaseRoutingModule } from './base-routing.module';

@NgModule({
  imports: [
    SharedModule,
    DashboardModule,
    ProjectModule,
    UserModule,
    BaseRoutingModule
  ],
  declarations: [
    NavigatorComponent,
    GlobalSearchComponent,
    FooterComponent,
    HarborShellComponent
  ],
  exports: [ HarborShellComponent ]
})
export class BaseModule {

}