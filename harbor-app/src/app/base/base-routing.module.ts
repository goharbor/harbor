import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { HarborShellComponent } from './harbor-shell/harbor-shell.component';

import { DashboardComponent } from '../dashboard/dashboard.component';
import { ProjectComponent } from '../project/project.component';
import { UserComponent } from '../user/user.component';

import { BaseRoutingResolver } from './base-routing-resolver.service';
import { AuthGuard } from './auth-guard.service';

const baseRoutes: Routes = [
  {
    path: 'harbor',
    component: HarborShellComponent,
    resolve: {
      rootResolver: BaseRoutingResolver
    },
    children: [
      {
        path: 'dashboard',
        component: DashboardComponent
      },
      {
        path: 'projects',
        component: ProjectComponent
      },
      {
        path: 'users',
        component: UserComponent,
        canActivate: [AuthGuard]
      }
    ]
  }];

@NgModule({
  imports: [
    RouterModule.forChild(baseRoutes)
  ],
  exports: [RouterModule],

  providers: [BaseRoutingResolver, AuthGuard]
})
export class BaseRoutingModule {

}