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
import { WebhookComponent } from './webhook.component';
import { AddWebhookFormComponent } from './add-webhook-form/add-webhook-form.component';
import { AddWebhookComponent } from './add-webhook/add-webhook.component';
import { ProjectWebhookService } from './webhook.service';
import { ExecutionsComponent } from './excutions/executions.component';
import { TasksComponent } from './tasks/tasks.component';
import { RouteConfigId } from '../../../route-reuse-strategy/harbor-route-reuse-strategy';

const routes: Routes = [
    {
        path: ':policyId/executions/:executionId/tasks',
        component: TasksComponent,
        data: {
            routeConfigId: RouteConfigId.WEBHOOK_TASKS_PAGE,
        },
    },
    {
        path: '',
        component: WebhookComponent,
        data: {
            reuse: true,
            routeConfigId: RouteConfigId.WEBHOOK_POLICIES_PAGE,
        },
    },
];
@NgModule({
    declarations: [
        WebhookComponent,
        AddWebhookFormComponent,
        AddWebhookComponent,
        ExecutionsComponent,
        TasksComponent,
    ],
    imports: [RouterModule.forChild(routes), SharedModule],
    providers: [ProjectWebhookService],
})
export class WebhookModule {}
