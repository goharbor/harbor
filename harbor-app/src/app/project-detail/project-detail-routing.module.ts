import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { ProjectDetailComponent } from './project-detail.component';

import { RepositoryComponent } from '../repository/repository.component';
import { ReplicationComponent } from '../replication/replication.component';
import { MemberComponent } from '../member/member.component';
import { LogComponent } from '../log/log.component';

const projectDetailRoutes: Routes = [
  { 
    path: 'projects/:id', 
    component: ProjectDetailComponent, 
    children: [
      { path: 'repository', component: RepositoryComponent },
      { path: 'replication', component: ReplicationComponent },
      { path: 'member', component: MemberComponent },
      { path: 'log', component: LogComponent }
    ]
  }
];

@NgModule({
  imports: [
    RouterModule.forChild(projectDetailRoutes)
  ],
  exports: [ RouterModule ]
})
export class ProjectDetailRoutingModule {}