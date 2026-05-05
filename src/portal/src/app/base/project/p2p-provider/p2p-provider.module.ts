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
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';
import { TaskListComponent } from './task-list/task-list.component';
import { PolicyComponent } from './policy/policy.component';
import { AddP2pPolicyComponent } from './add-p2p-policy/add-p2p-policy.component';
import { P2pProviderService } from './p2p-provider.service';
import { RouteConfigId } from '../../../route-reuse-strategy/harbor-route-reuse-strategy';

const routes: Routes = [
    {
        path: 'policies',
        component: PolicyComponent,
        data: {
            reuse: true,
            routeConfigId: RouteConfigId.P2P_POLICIES_PAGE,
        },
    },
    {
        path: ':preheatPolicyName/executions/:executionId/tasks',
        component: TaskListComponent,
        data: {
            routeConfigId: RouteConfigId.P2P_TASKS_PAGE,
        },
    },
    { path: '', redirectTo: 'policies', pathMatch: 'full' },
];
@NgModule({
    declarations: [TaskListComponent, PolicyComponent, AddP2pPolicyComponent],
    imports: [RouterModule.forChild(routes), SharedModule],
    providers: [P2pProviderService],
})
export class P2pProviderModule {}
