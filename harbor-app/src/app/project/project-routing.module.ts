import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { BaseComponent } from '../base/base.component';

import { ProjectComponent } from './project.component';
import { ProjectDetailComponent } from './project-detail/project-detail.component';

import { RepositoryComponent } from '../repository/repository.component';
import { ReplicationComponent } from '../replication/replication.component';
import { MemberComponent } from '../member/member.component';
import { LogComponent } from '../log/log.component';

const projectRoutes: Routes = [
  { path: 'harbor', 
    component: BaseComponent, 
    children: [
      { path: 'projects', component: ProjectComponent },
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
    ]
  }
];

@NgModule({
  imports: [
    RouterModule.forChild(projectRoutes)
  ],
  exports: [ RouterModule ]
})
export class ProjectRoutingModule {}