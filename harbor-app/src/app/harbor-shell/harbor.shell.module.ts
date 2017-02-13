import { NgModule } from '@angular/core';
import { NavigatorModule } from '../navigator/navigator.module';
import { GlobalSearchModule } from '../global-search/global-search.module';
import { FooterModule } from '../footer/footer.module';

import { DashboardModule } from '../dashboard/dashboard.module';
import { ProjectModule } from '../project/project.module';
import { ProjectDetailModule } from '../project-detail/project-detail.module';

import { HarborShellRoutingModule } from './harbor-shell-routing.module';

import { HarborShellComponent } from './harbor-shell.component';

import { SharedModule } from '../shared.module';

@NgModule({
  imports: [
    SharedModule,
    GlobalSearchModule,
    NavigatorModule,
    FooterModule,
    DashboardModule,
    ProjectModule,
    ProjectDetailModule,
    HarborShellRoutingModule
  ],
  declarations: [ HarborShellComponent ],
  exports: [ HarborShellComponent ]
})
export class HarborShellModule {}