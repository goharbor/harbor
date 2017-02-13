import { NgModule } from '@angular/core';

import { SharedModule } from '../../shared.module';

import { ProjectDetailComponent } from './project-detail.component';
import { RepositoryComponent } from '../../repository/repository.component';
import { ReplicationComponent } from '../../replication/replication.component';
import { MemberComponent } from '../../member/member.component';
import { LogComponent } from '../../log/log.component';

import { RouterModule } from '@angular/router';

@NgModule({
  imports: [ 
    SharedModule,
    RouterModule
  ],
  declarations: [ 
    ProjectDetailComponent,
    RepositoryComponent,
    ReplicationComponent,
    MemberComponent,
    LogComponent
  ],
  exports: [ ProjectDetailComponent ]
})
export class ProjectDetailModule {}