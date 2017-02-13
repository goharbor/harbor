import { NgModule } from '@angular/core';
import { SharedModule } from '../shared.module';

import { DashboardModule } from '../dashboard/dashboard.module';
import { ProjectModule } from '../project/project.module';
import { ProjectDetailModule } from '../project/project-detail/project-detail.module';

import { NavigatorComponent } from './navigator/navigator.component';
import { GlobalSearchComponent } from './global-search/global-search.component';
import { FooterComponent } from './footer/footer.component';
import { HarborShellComponent } from './harbor-shell/harbor-shell.component';
import { BaseComponent } from './base.component';

import { BaseRoutingModule } from './base-routing.module';

@NgModule({
  imports: [
    SharedModule,
    DashboardModule,
    ProjectModule,
    ProjectDetailModule,
    BaseRoutingModule
  ],
  declarations: [
    BaseComponent,
    NavigatorComponent,
    GlobalSearchComponent,
    FooterComponent,
    HarborShellComponent
  ],
  exports: [ BaseComponent ]
})
export class BaseModule {

}