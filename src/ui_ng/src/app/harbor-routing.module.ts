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

import { TotalReplicationComponent } from './replication/total-replication/total-replication.component';
import { DestinationComponent } from './replication/destination/destination.component';

import { ProjectDetailComponent } from './project/project-detail/project-detail.component';

import { RepositoryComponent } from './repository/repository.component';
import { TagRepositoryComponent } from './repository/tag-repository/tag-repository.component';
import { ReplicationComponent } from './replication/replication.component';
import { MemberComponent } from './project/member/member.component';
import { AuditLogComponent } from './log/audit-log.component';

import { ProjectRoutingResolver } from './project/project-routing-resolver.service';
import { SystemAdminGuard } from './shared/route/system-admin-activate.service';
import { SignUpComponent } from './account/sign-up/sign-up.component';
import { ResetPasswordComponent } from './account/password/reset-password.component';
import { RecentLogComponent } from './log/recent-log.component';
import { ConfigurationComponent } from './config/config.component';
import { PageNotFoundComponent } from './shared/not-found/not-found.component'
import { StartPageComponent } from './base/start-page/start.component';
import { SignUpPageComponent } from './account/sign-up/sign-up-page.component';

import { AuthCheckGuard } from './shared/route/auth-user-activate.service';
import { SignInGuard } from './shared/route/sign-in-guard-activate.service';
import { LeavingConfigRouteDeactivate } from './shared/route/leaving-config-deactivate.service';

import { MemberGuard } from './shared/route/member-guard-activate.service';

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
        component: RecentLogComponent
      },
      {
        path: 'users',
        component: UserComponent,
        canActivate: [SystemAdminGuard]
      },
      {
        path: 'replications',
        component: ReplicationManagementComponent,
        canActivate: [SystemAdminGuard],
        canActivateChild: [SystemAdminGuard],
        children: [
          {
            path: 'rules',
            component: TotalReplicationComponent
          },
          {
            path: 'endpoints',
            component: DestinationComponent
          }
        ]
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
        path: 'projects/:id',
        component: ProjectDetailComponent,
        canActivate: [MemberGuard],
        resolve: {
          projectResolver: ProjectRoutingResolver
        },
        children: [
          {
            path: 'repository',
            component: RepositoryComponent
          },
          {
            path: 'replication',
            component: ReplicationComponent,
            canActivate: [SystemAdminGuard]
          },
          {
            path: 'member',
            component: MemberComponent
          },
          {
            path: 'log',
            component: AuditLogComponent
          }
        ]
      },
      {
        path: 'configs',
        component: ConfigurationComponent,
        canActivate: [SystemAdminGuard],
        canDeactivate: [LeavingConfigRouteDeactivate]
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