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
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';
import { ModelSyncPageComponent } from './model-sync-page.component';
import { ModelSyncPolicyListComponent } from './policy-list/policy-list.component';
import { CreateEditModelSyncPolicyComponent } from './create-edit-policy/create-edit-policy.component';
import { ModelSyncExecutionTasksComponent } from './execution-tasks/execution-tasks.component';
import { RouteConfigId } from '../../../route-reuse-strategy/harbor-route-reuse-strategy';

const routes: Routes = [
    {
        path: '',
        component: ModelSyncPageComponent,
        data: {
            reuse: true,
            routeConfigId: RouteConfigId.MODEL_SYNC_PAGE,
        },
    },
    {
        path: 'executions/:id/tasks',
        component: ModelSyncExecutionTasksComponent,
        data: {
            routeConfigId: RouteConfigId.MODEL_SYNC_TASKS_PAGE,
        },
    },
];

@NgModule({
    imports: [SharedModule, ReactiveFormsModule, RouterModule.forChild(routes)],
    declarations: [
        ModelSyncPageComponent,
        ModelSyncPolicyListComponent,
        CreateEditModelSyncPolicyComponent,
        ModelSyncExecutionTasksComponent,
    ],
})
export class ModelSyncModule {}
