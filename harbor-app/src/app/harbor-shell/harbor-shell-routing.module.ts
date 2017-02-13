import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { HarborShellComponent } from './harbor-shell.component';

import { DashboardComponent } from '../dashboard/dashboard.component';
import { ProjectComponent } from '../project/project.component';
import { ProjectListComponent } from '../project/project-list.component';

import { ProjectDetailComponent } from '../project-detail/project-detail.component';
import { RepositoryComponent } from '../repository/repository.component';
import { ReplicationComponent } from '../replication/replication.component';
import { MemberComponent } from '../member/member.component';
import { LogComponent } from '../log/log.component';

const harborShellRoutes: Routes = [
  { 
    path: 'harbor',
    component: HarborShellComponent, 
    children: [
      { path: 'dashboard', component: DashboardComponent },
      { path: 'project', component: ProjectComponent }
    ]
  }
];

@NgModule({
  imports: [
    RouterModule.forChild(harborShellRoutes)
  ],
  exports: [ RouterModule ]
})
export class HarborShellRoutingModule {

}