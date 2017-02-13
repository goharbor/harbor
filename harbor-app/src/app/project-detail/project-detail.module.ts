import { NgModule } from '@angular/core';

import { NavigatorModule } from '../navigator/navigator.module';
import { GlobalSearchModule } from '../global-search/global-search.module';
import { FooterModule } from '../footer/footer.module';


import { RepositoryModule } from '../repository/repository.module';
import { ReplicationModule } from '../replication/replication.module';
import { MemberModule } from '../member/member.module';
import { LogModule } from '../log/log.module';

import { ProjectDetailComponent } from './project-detail.component';

import { SharedModule } from '../shared.module';

import { ProjectDetailRoutingModule } from './project-detail-routing.module';

@NgModule({
  imports: [ 
    SharedModule,
    GlobalSearchModule,
    NavigatorModule,
    FooterModule,
    RepositoryModule,
    ReplicationModule,
    MemberModule,
    LogModule,
    ProjectDetailRoutingModule
  ],
  declarations: [ ProjectDetailComponent ],
  exports: [ ProjectDetailComponent ]
})
export class ProjectDetailModule {}