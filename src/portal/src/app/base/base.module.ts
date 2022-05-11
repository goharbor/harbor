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
import { SharedModule } from '../shared/shared.module';
import { RouterModule, Routes } from '@angular/router';
import { HarborShellComponent } from './harbor-shell/harbor-shell.component';
import { SystemAdminGuard } from '../shared/router-guard/system-admin-activate.service';
import { MemberGuard } from '../shared/router-guard/member-guard-activate.service';
import { ProjectRoutingResolver } from '../services/routing-resolvers/project-routing-resolver.service';
import { PasswordSettingComponent } from './password-setting/password-setting.component';
import { AccountSettingsModalComponent } from './account-settings/account-settings-modal.component';
import { ForgotPasswordComponent } from './password-setting/forgot-password/forgot-password.component';
import { GlobalConfirmationDialogComponent } from './global-confirmation-dialog/global-confirmation-dialog.component';

const routes: Routes = [
    {
        path: '',
        component: HarborShellComponent,
        children: [
            { path: '', redirectTo: 'projects', pathMatch: 'full' },
            {
                path: 'projects',
                loadChildren: () =>
                    import('./left-side-nav/projects/projects.module').then(
                        m => m.ProjectsModule
                    ),
            },
            {
                path: 'logs',
                loadChildren: () =>
                    import('./left-side-nav/log/log.module').then(
                        m => m.LogModule
                    ),
            },
            {
                path: 'users',
                canActivate: [SystemAdminGuard],
                loadChildren: () =>
                    import('./left-side-nav/user/user.module').then(
                        m => m.UserModule
                    ),
            },
            {
                path: 'robot-accounts',
                canActivate: [SystemAdminGuard],
                loadChildren: () =>
                    import(
                        './left-side-nav/system-robot-accounts/system-robot-accounts.module'
                    ).then(m => m.SystemRobotAccountsModule),
            },
            {
                path: 'groups',
                canActivate: [SystemAdminGuard],
                loadChildren: () =>
                    import('./left-side-nav/group/group.module').then(
                        m => m.GroupModule
                    ),
            },
            {
                path: 'registries',
                canActivate: [SystemAdminGuard],
                loadChildren: () =>
                    import('./left-side-nav/registries/endpoint.module').then(
                        m => m.EndpointModule
                    ),
            },
            {
                path: 'replications',
                canActivate: [SystemAdminGuard],
                loadChildren: () =>
                    import(
                        './left-side-nav/replication/replication.module'
                    ).then(m => m.ReplicationModule),
            },
            {
                path: 'distribution',
                canActivate: [SystemAdminGuard],
                loadChildren: () =>
                    import(
                        './left-side-nav/distribution/distribution.module'
                    ).then(m => m.DistributionModule),
            },
            {
                path: 'interrogation-services',
                canActivate: [SystemAdminGuard],
                loadChildren: () =>
                    import(
                        './left-side-nav/interrogation-services/interrogation-services.module'
                    ).then(m => m.InterrogationServicesModule),
            },
            {
                path: 'labels',
                canActivate: [SystemAdminGuard],
                loadChildren: () =>
                    import('./left-side-nav/labels/labels.module').then(
                        m => m.LabelsModule
                    ),
            },
            {
                path: 'project-quotas',
                canActivate: [SystemAdminGuard],
                loadChildren: () =>
                    import(
                        './left-side-nav/project-quotas/project-quotas.module'
                    ).then(m => m.ProjectQuotasModule),
            },
            {
                path: 'gc',
                canActivate: [SystemAdminGuard],
                loadChildren: () =>
                    import('./left-side-nav/gc-page/gc.module').then(
                        m => m.GcModule
                    ),
            },
            {
                path: 'configs',
                canActivate: [SystemAdminGuard],
                loadChildren: () =>
                    import('./left-side-nav/config/config.module').then(
                        m => m.ConfigurationModule
                    ),
            },
            {
                path: 'projects/:id',
                loadChildren: () =>
                    import('./project/project.module').then(
                        m => m.ProjectModule
                    ),
                canActivate: [MemberGuard],
                resolve: {
                    projectResolver: ProjectRoutingResolver,
                },
            },
            {
                path: 'projects/:id/repositories',
                loadChildren: () =>
                    import(
                        './project/repository/artifact/artifact.module'
                    ).then(m => m.ArtifactModule),
                canActivate: [MemberGuard],
                resolve: {
                    projectResolver: ProjectRoutingResolver,
                },
            },
            {
                path: 'projects/:id/helm-charts',
                canActivate: [MemberGuard],
                resolve: {
                    projectResolver: ProjectRoutingResolver,
                },
                loadChildren: () =>
                    import(
                        './project/helm-chart/helm-chart-detail/helm-chart-detail.module'
                    ).then(m => m.HelmChartListModule),
            },
        ],
    },
];
@NgModule({
    imports: [SharedModule, RouterModule.forChild(routes)],
    declarations: [
        HarborShellComponent,
        PasswordSettingComponent,
        AccountSettingsModalComponent,
        ForgotPasswordComponent,
        GlobalConfirmationDialogComponent,
    ],
})
export class BaseModule {}
