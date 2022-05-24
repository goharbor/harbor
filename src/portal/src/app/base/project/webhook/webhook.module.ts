import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';
import { WebhookComponent } from './webhook.component';
import { LastTriggerComponent } from './last-trigger/last-trigger.component';
import { AddWebhookFormComponent } from './add-webhook-form/add-webhook-form.component';
import { AddWebhookComponent } from './add-webhook/add-webhook.component';
import { ProjectWebhookService } from './webhook.service';

const routes: Routes = [
    {
        path: '',
        component: WebhookComponent,
    },
];
@NgModule({
    declarations: [
        WebhookComponent,
        LastTriggerComponent,
        AddWebhookFormComponent,
        AddWebhookComponent,
    ],
    imports: [RouterModule.forChild(routes), SharedModule],
    providers: [ProjectWebhookService],
})
export class WebhookModule {}
