// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

import { SignInComponent } from './account/sign-in/sign-in.component';
import { HarborShellComponent } from './base/harbor-shell/harbor-shell.component';
import { ProjectComponent } from './project/project.component';
import { UserComponent } from './user/user.component';
import { ReplicationManagementComponent } from './replication/replication-management/replication-management.component';

import { TotalReplicationPageComponent } from './replication/total-replication/total-replication-page.component';
import { DestinationPageComponent } from './replication/destination/destination-page.component';

import { ProjectDetailComponent } from './project/project-detail/project-detail.component';

import { RepositoryPageComponent } from './repository/repository-page.component';
import { TagRepositoryComponent } from './repository/tag-repository/tag-repository.component';
import { ReplicationPageComponent } from './replication/replication-page.component';
import { MemberComponent } from './project/member/member.component';
import { AuditLogComponent } from './log/audit-log.component';
import { ProjectConfigComponent } from './project/project-config/project-config.component'

import { ProjectRoutingResolver } from './project/project-routing-resolver.service';
import { SystemAdminGuard } from './shared/route/system-admin-activate.service';
import { SignUpComponent } from './account/sign-up/sign-up.component';
import { ResetPasswordComponent } from './account/password/reset-password.component';
import { LogPageComponent } from './log/log-page.component';
import { ConfigurationComponent } from './config/config.component';
import { PageNotFoundComponent } from './shared/not-found/not-found.component'
import { StartPageComponent } from './base/start-page/start.component';
import { SignUpPageComponent } from './account/sign-up/sign-up-page.component';

import { AuthCheckGuard } from './shared/route/auth-user-activate.service';
import { SignInGuard } from './shared/route/sign-in-guard-activate.service';
import { LeavingConfigRouteDeactivate } from './shared/route/leaving-config-deactivate.service';

import { MemberGuard } from './shared/route/member-guard-activate.service';

import { TagDetailPageComponent } from './repository/tag-detail/tag-detail-page.component';
import { ReplicationRuleComponent} from "./replication/replication-rule/replication-rule.component";
import {LeavingNewRuleRouteDeactivate} from "./shared/route/leaving-new-rule-deactivate.service";

const harborRoutes: Routes = [
  { path: '', redirectTo: 'harbor', pathMatch: 'full' },
  { path: 'reset_password', component: ResetPasswordComponent },
  {
    path: 'harbor',
    component: HarborShellComponent,
    canActivateChild: [AuthCheckGuard],
    children: [
      { path: '', redirectTo: 'sign-in', pathMatch: 'full' },
      {
        path: 'sign-in',
        component: SignInComponent,
        canActivate: [SignInGuard]
      },
      {
        path: 'projects',
        component: ProjectComponent
      },
      {
        path: 'logs',
        component: LogPageComponent
      },
      {
        path: 'users',
        component: UserComponent,
        canActivate: [SystemAdminGuard]
      },
      {
        path: 'replications',
        component: TotalReplicationPageComponent,
        canActivate: [SystemAdminGuard],
        canActivateChild: [SystemAdminGuard],
      },
      {
          path: 'replications/:id/rule',
          component: ReplicationRuleComponent,
          canActivate: [SystemAdminGuard],
          canActivateChild: [SystemAdminGuard],
          canDeactivate: [LeavingNewRuleRouteDeactivate]
      },
      {
        path: 'replications/new-rule',
        component: ReplicationRuleComponent,
        canActivate: [SystemAdminGuard],
        canActivateChild: [SystemAdminGuard],
        canDeactivate: [LeavingNewRuleRouteDeactivate]
      },
      {
        path: 'tags/:id/:repo',
        component: TagRepositoryComponent,
        canActivate: [MemberGuard],
        resolve: {
          projectResolver: ProjectRoutingResolver
        }
      },
      {
        path: 'projects/:id/repositories/:repo',
        component: TagRepositoryComponent,
        canActivate: [MemberGuard],
        resolve: {
          projectResolver: ProjectRoutingResolver
        }
      },
      {
        path: 'projects/:id',
        component: ProjectDetailComponent,
        canActivate: [MemberGuard],
        resolve: {
          projectResolver: ProjectRoutingResolver
        },
        children: [
          {
            path: 'repositories',
            component: RepositoryPageComponent
          },
          {
            path: 'repositories/:repo/tags',
            component: TagRepositoryComponent,
          },
          {
            path: 'repositories/:repo/tags/:tag',
            component: TagDetailPageComponent
          },
          {
            path: 'replications',
            component: ReplicationPageComponent,
            canActivate: [SystemAdminGuard]
          },
          {
            path: 'members',
            component: MemberComponent
          },
          {
            path: 'logs',
            component: AuditLogComponent
          },
          {
            path: 'configs',
            component: ProjectConfigComponent
          }
        ]
      },
      {
        path: 'configs',
        component: ConfigurationComponent,
        canActivate: [SystemAdminGuard],
        canDeactivate: [LeavingConfigRouteDeactivate]
      },
      {
        path: 'registry',
        component: DestinationPageComponent,
        canActivate: [SystemAdminGuard],
        canActivateChild: [SystemAdminGuard],
      }
    ]
  },
  { path: "**", component: PageNotFoundComponent }
];

@NgModule({
  imports: [
    RouterModule.forRoot(harborRoutes)
  ],
  exports: [RouterModule]
})
export class HarborRoutingModule {

}