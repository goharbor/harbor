import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { BaseComponent } from './base.component';

import { DashboardComponent } from '../dashboard/dashboard.component';
import { ProjectComponent } from '../project/project.component';

import { ProjectDetailComponent } from '../project/project-detail/project-detail.component';
import { RepositoryComponent } from '../repository/repository.component';
import { ReplicationComponent } from '../replication/replication.component';
import { MemberComponent } from '../member/member.component';
import { LogComponent } from '../log/log.component';

const baseRoutes: Routes = [
  { 
    path: 'harbor', component: BaseComponent,
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