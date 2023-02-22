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
