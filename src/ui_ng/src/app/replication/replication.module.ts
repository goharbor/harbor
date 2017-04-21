// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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