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
import { ReactiveFormsModule } from '@angular/forms';
import { SharedModule } from '../../../shared/shared.module';
import { TotalReplicationPageComponent } from './total-replication-page.component';
import { ReplicationComponent } from './replication/replication.component';
import { ReplicationTasksComponent } from './replication/replication-tasks/replication-tasks.component';
import { ReplicationDefaultService } from '../../../shared/services';
import { ReplicationTasksRoutingResolverService } from '../../../services/routing-resolvers/replication-tasks-routing-resolver.service';
import { RouterModule, Routes } from '@angular/router';
import { ListReplicationRuleComponent } from './replication/list-replication-rule/list-replication-rule.component';
import { CreateEditRuleComponent } from './replication/create-edit-rule/create-edit-rule.component';
import { RouteConfigId } from '../../../route-reuse-strategy/harbor-route-reuse-strategy';
const routes: Routes = [
    {
        path: '',
        component: TotalReplicationPageComponent,
        data: {
            reuse: true,
            routeConfigId: RouteConfigId.REPLICATION_PAGE,
        },
    },
    {
        path: ':id/tasks',
        component: ReplicationTasksComponent,
        resolve: {
            replicationTasksRoutingResolver:
                ReplicationTasksRoutingResolverService,
        },
        data: {
            routeConfigId: RouteConfigId.REPLICATION_TASKS_PAGE,
        },
    },
];
@NgModule({
    imports: [SharedModule, ReactiveFormsModule, RouterModule.forChild(routes)],
    declarations: [
        TotalReplicationPageComponent,
        ReplicationComponent,
        ReplicationTasksComponent,
        ListReplicationRuleComponent,
        CreateEditRuleComponent,
    ],
    providers: [ReplicationDefaultService],
})
export class ReplicationModule {}
