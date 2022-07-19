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
