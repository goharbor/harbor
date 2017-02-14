import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { HarborShellComponent } from './harbor-shell/harbor-shell.component';

import { DashboardComponent } from '../dashboard/dashboard.component';
import { ProjectComponent } from '../project/project.component';

const baseRoutes: Routes = [
  { 
    path: 'harbor', component: HarborShellComponent,
    children: [
      { path: 'dashboard', component: DashboardComponent },
      { path: 'projects', component: ProjectComponent }
    ]
  }];

@NgModule({
  imports: [
    RouterModule.forChild(baseRoutes)
  ],
  exports: [ RouterModule ]
})
export class BaseRoutingModule {

}