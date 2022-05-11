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
import { SharedModule } from '../../shared/shared.module';
import { ProjectDetailComponent } from './project-detail/project-detail.component';
import { MemberPermissionGuard } from '../../shared/router-guard/member-permission-guard-activate.service';
import { USERSTATICPERMISSION } from '../../shared/services';

const routes: Routes = [
    {
        path: '',
        component: ProjectDetailComponent,
        children: [
            {
                path: 'summary',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.PROJECT.KEY,
                        action: USERSTATICPERMISSION.PROJECT.VALUE.READ,
                    },
                },
                loadChildren: () =>
                    import('./summary/summary.module').then(
                        m => m.SummaryModule
                    ),
            },
            {
                path: 'repositories',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.REPOSITORY.KEY,
                        action: USERSTATICPERMISSION.REPOSITORY.VALUE.LIST,
                    },
                },
                loadChildren: () =>
                    import('./repository/repository.module').then(
                        m => m.RepositoryModule
                    ),
            },
            {
                path: 'helm-charts',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.HELM_CHART.KEY,
                        action: USERSTATICPERMISSION.HELM_CHART.VALUE.LIST,
                    },
                },
                loadChildren: () =>
                    import(
                        './helm-chart/helm-chart-list/helm-chart-list.module'
                    ).then(m => m.HelmChartListModule),
            },
            {
                path: 'members',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.MEMBER.KEY,
                        action: USERSTATICPERMISSION.MEMBER.VALUE.LIST,
                    },
                },
                loadChildren: () =>
                    import('./member/member.module').then(m => m.MemberModule),
            },
            {
                path: 'logs',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.LOG.KEY,
                        action: USERSTATICPERMISSION.LOG.VALUE.LIST,
                    },
                },
                loadChildren: () =>
                    import('./project-log/audit-log.module').then(
                        m => m.AuditLogModule
                    ),
            },
            {
                path: 'labels',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.LABEL.KEY,
                        action: USERSTATICPERMISSION.LABEL.VALUE.CREATE,
                    },
                },
                loadChildren: () =>
                    import('./project-label/project-label.module').then(
                        m => m.ProjectLabelModule
                    ),
            },
            {
                path: 'configs',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.CONFIGURATION.KEY,
                        action: USERSTATICPERMISSION.CONFIGURATION.VALUE.READ,
                    },
                },
                loadChildren: () =>
                    import('./project-config/project-config.module').then(
                        m => m.ProjectConfigModule
                    ),
            },
            {
                path: 'robot-account',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.ROBOT.KEY,
                        action: USERSTATICPERMISSION.ROBOT.VALUE.LIST,
                    },
                },
                loadChildren: () =>
                    import('./robot-account/project-robot-account.module').then(
                        m => m.ProjectRobotAccountModule
                    ),
            },
            {
                path: 'tag-strategy',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.TAG_RETENTION.KEY,
                        action: USERSTATICPERMISSION.TAG_RETENTION.VALUE.READ,
                    },
                },
                loadChildren: () =>
                    import(
                        './tag-feature-integration/tag-feature-integration.module'
                    ).then(m => m.TagFeatureIntegrationModule),
            },
            {
                path: 'webhook',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.WEBHOOK.KEY,
                        action: USERSTATICPERMISSION.WEBHOOK.VALUE.LIST,
                    },
                },
                loadChildren: () =>
                    import('./webhook/webhook.module').then(
                        m => m.WebhookModule
                    ),
            },
            {
                path: 'scanner',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.SCANNER.KEY,
                        action: USERSTATICPERMISSION.SCANNER.VALUE.READ,
                    },
                },
                loadChildren: () =>
                    import('./scanner/project-scanner.module').then(
                        m => m.ProjectScannerModule
                    ),
            },
            {
                path: 'p2p-provider',
                canActivate: [MemberPermissionGuard],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.P2P_PROVIDER.KEY,
                        action: USERSTATICPERMISSION.P2P_PROVIDER.VALUE.READ,
                    },
                },
                loadChildren: () =>
                    import('./p2p-provider/p2p-provider.module').then(
                        m => m.P2pProviderModule
                    ),
            },
            {
                path: '',
                redirectTo: 'repositories',
                pathMatch: 'full',
            },
        ],
    },
];
@NgModule({
    imports: [RouterModule.forChild(routes), SharedModule],
    declarations: [ProjectDetailComponent],
})
export class ProjectModule {}
