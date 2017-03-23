import { NgModule } from '@angular/core';

import { RouterModule } from '@angular/router';
import { SharedModule } from '../shared/shared.module';
import { RepositoryModule } from '../repository/repository.module';
import { ReplicationModule } from '../replication/replication.module';
import { LogModule } from '../log/log.module';

import { ProjectComponent } from './project.component';
import { CreateProjectComponent } from './create-project/create-project.component';
import { ListProjectComponent } from './list-project/list-project.component';

import { ProjectDetailComponent } from './project-detail/project-detail.component';
import { MemberComponent } from './member/member.component';
import { AddMemberComponent } from './member/add-member/add-member.component';

import { ProjectService } from './project.service';
import { MemberService } from './member/member.service';
import { ProjectRoutingResolver } from './project-routing-resolver.service';

import { TargetExistsValidatorDirective } from '../shared/target-exists-directive';

@NgModule({
  imports: [
    SharedModule,
    RepositoryModule,
    ReplicationModule,
    LogModule,
    RouterModule
  ],
  declarations: [
    ProjectComponent,
    CreateProjectComponent,
    ListProjectComponent,
    ProjectDetailComponent,
    MemberComponent,
    AddMemberComponent,
    TargetExistsValidatorDirective
  ],
  exports: [ProjectComponent, ListProjectComponent],
  providers: [ProjectRoutingResolver, ProjectService, MemberService]
})
export class ProjectModule {

}