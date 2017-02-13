import { NgModule } from '@angular/core';

import { SharedModule } from '../shared.module';
import { RepositoryModule } from '../repository/repository.module';
import { ReplicationModule } from '../replication/replication.module';
import { MemberModule} from '../member/member.module';
import { LogModule } from '../log/log.module';

import { ProjectComponent } from './project.component';
import { CreateProjectComponent } from './create-project/create-project.component';
import { SearchProjectComponent } from './search-project/search-project.component';
import { FilterProjectComponent } from './filter-project/filter-project.component';
import { ListProjectComponent } from './list-project/list-project.component';
import { ProjectDetailComponent } from './project-detail/project-detail.component';
import { ProjectRoutingModule } from './project-routing.module';


@NgModule({
  imports: [ 
    SharedModule,
    RepositoryModule,
    ReplicationModule,
    MemberModule,
    LogModule,
    ProjectRoutingModule
  ],
  declarations: [ 
    ProjectComponent,
    CreateProjectComponent,
    SearchProjectComponent,
    FilterProjectComponent,
    ListProjectComponent,
    ProjectDetailComponent
  ],
  exports: [ ListProjectComponent ]
})
export class ProjectModule {}