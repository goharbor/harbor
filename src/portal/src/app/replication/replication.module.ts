// Copyright Project Harbor Authors
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
import { ReplicationPageComponent } from './replication-page.component';
import { TotalReplicationPageComponent } from './total-replication/total-replication-page.component';
import { DestinationPageComponent } from './destination/destination-page.component';
import { ReplicationTasksPageComponent } from './replication-tasks-page/replication-tasks-page.component';

import { SharedModule } from '../shared/shared.module';
import {ReactiveFormsModule} from "@angular/forms";

@NgModule({
  imports: [
    SharedModule,
    RouterModule,
    ReactiveFormsModule
  ],
  declarations: [
    ReplicationPageComponent,
    ReplicationManagementComponent,
    TotalReplicationPageComponent,
    ReplicationTasksPageComponent,
    DestinationPageComponent,
  ],
  exports: [
    ReplicationPageComponent,
    DestinationPageComponent,
    ReplicationTasksPageComponent,
    TotalReplicationPageComponent,
  ]
})
export class ReplicationModule { }
