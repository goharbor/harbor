import { NgModule } from '@angular/core';

import { SharedModule } from '../shared/shared.module';
import { RepositoryModule } from '../repository/repository.module';
import { ReplicationModule } from '../replication/replication.module';
import { LogModule } from '../log/log.module';

import { ProjectComponent } from './project.component';
import { CreateProjectComponent } from './create-project/create-project.component';
import { SearchProjectComponent } from './search-project/search-project.component';
import { FilterProjectComponent } from './filter-project/filter-project.component';
import { ActionProjectComponent } from './action-project/action-project.component';
import { ListProjectComponent } from './list-project/list-project.component';
import { ProjectDetailComponent } from './project-detail/project-detail.component';

import { MemberComponent } from './member/member.component';
import { ProjectRoutingModule } from './project-routing.module';

import { ProjectService } from './project.service';
import { DATAGRID_DIRECTIVES } from 'clarity-angular';

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
    SearchProjectComponent,
    FilterProjectComponent,
    ActionProjectComponent,
    ListProjectComponent,
    ProjectDetailComponent,
    MemberComponent 
  ],
  exports: [ ListProjectComponent ],
  providers: [ ProjectService ]
})
export class ProjectModule {
  
}