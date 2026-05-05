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
