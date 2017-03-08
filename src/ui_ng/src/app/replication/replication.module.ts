import { NgModule } from '@angular/core';
import { RouterModule } from '@angular/router';
import { ReplicationManagementComponent } from './replication-management/replication-management.component';

import { ReplicationComponent } from './replication.component';
import { ListJobComponent } from './list-job/list-job.component';
import { TotalReplicationComponent } from './total-replication/total-replication.component';
import { DestinationComponent } from './destination/destination.component';
import { CreateEditDestinationComponent } from './create-edit-destination/create-edit-destination.component';

import { SharedModule } from '../shared/shared.module';
import { ReplicationService } from './replication.service';

@NgModule({
  imports: [ 
    SharedModule,
    RouterModule
  ],
  declarations: [ 
    ReplicationComponent,
    ReplicationManagementComponent,
    ListJobComponent,
    TotalReplicationComponent,
    DestinationComponent,
    CreateEditDestinationComponent
  ],
  exports: [ ReplicationComponent ],
  providers: [ ReplicationService ]
})
export class ReplicationModule {}