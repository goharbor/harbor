import { NgModule } from '@angular/core';

import { SharedModule } from '../shared/shared.module';
import { RepositoryModule } from '../repository/repository.module';
import { ReplicationModule } from '../replication/replication.module';
import { LogModule } from '../log/log.module';

import { ProjectComponent } from './project.component';
import { CreateProjectComponent } from './create-project/create-project.component';
import { ActionProjectComponent } from './action-project/action-project.component';
import { ListProjectComponent } from './list-project/list-project.component';

import { ProjectDetailComponent } from './project-detail/project-detail.component';

import { MemberComponent } from './member/member.component';
import { AddMemberComponent } from './member/add-member/add-member.component';

import { ProjectRoutingModule } from './project-routing.module';

import { ProjectService } from './project.service';
import { MemberService } from './member/member.service';
import { ProjectRoutingResolver } from './project-routing-resolver.service';

@NgModule({
  imports: [ 
    SharedModule,
    RepositoryModule,
    ReplicationModule,
    LogModule,
    ProjectRoutingModule
  ],
  declarations: [ 
    ProjectComponent,
    CreateProjectComponent,
    ActionProjectComponent,
    ListProjectComponent,
    ProjectDetailComponent,
    MemberComponent,
    AddMemberComponent
  ],
  exports: [ ProjectComponent ],
  providers: [ ProjectRoutingResolver, ProjectService, MemberService ]
})
export class ProjectModule {
  
}