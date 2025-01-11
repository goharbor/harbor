import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';
import { ProjectAuditLogComponent } from './audit-log.component';
import { ProjectLogsComponent } from './project-logs.component';
import { ProjectAuditLegacyLogComponent } from './audit-legacy-log.component';

const routes: Routes = [
    {
        path: '',
        component: ProjectLogsComponent,
        children: [
            {
                path: 'project-audit-log',
                component: ProjectAuditLogComponent,
            },
            {
                path: 'project-audit-legacy-log',
                component: ProjectAuditLegacyLogComponent,
            },
            {
                path: '',
                redirectTo: 'project-audit-log',
                pathMatch: 'full',
            },
        ],
    },
];
@NgModule({
    declarations: [
        ProjectLogsComponent,
        ProjectAuditLogComponent,
        ProjectAuditLegacyLogComponent,
    ],
    imports: [RouterModule.forChild(routes), SharedModule],
})
export class AuditLogModule {}
