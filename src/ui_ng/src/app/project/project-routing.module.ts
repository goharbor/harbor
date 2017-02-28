import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { HarborShellComponent } from '../base/harbor-shell/harbor-shell.component';
import { ProjectComponent } from './project.component';
import { ProjectDetailComponent } from './project-detail/project-detail.component';

import { RepositoryComponent } from '../repository/repository.component';
import { ReplicationComponent } from '../replication/replication.component';
import { MemberComponent } from './member/member.component';
import { AuditLogComponent } from '../log/audit-log.component';

import { BaseRoutingResolver } from '../base/base-routing-resolver.service';
import { ProjectRoutingResolver } from './project-routing-resolver.service';

const projectRoutes: Routes = [
  {
    path: 'harbor',
    component: HarborShellComponent,
    resolve: {
      harborResolver: BaseRoutingResolver
    },
    children: [
      {
        path: 'projects',
        component: ProjectComponent,
        resolve: {
          projectsResolver: BaseRoutingResolver
        }
      },
      {
        path: 'projects/:id',
        component: ProjectDetailComponent,
        resolve: {
          projectResolver: ProjectRoutingResolver
        },
        children: [
          { path: 'repository', component: RepositoryComponent },
          { 
            path: 'replication', component: ReplicationComponent,
            resolve: {
              replicationResolver: BaseRoutingResolver
            } 
          },
          { 
            path: 'member', component: MemberComponent, 
            resolve: {
              memberResolver: BaseRoutingResolver
            }
          },
          { 
            path: 'log', component: AuditLogComponent,
            resolve: {
              auditLogResolver: BaseRoutingResolver
            }   
          }
        ]
      }
    ]
  }
];

@NgModule({
  imports: [
    RouterModule.forChild(projectRoutes)
  ],
  exports: [RouterModule]
})
export class ProjectRoutingModule { }