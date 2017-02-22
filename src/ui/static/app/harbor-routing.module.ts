import { NgModule } from '@angular/core';

import { RouterModule, Routes } from '@angular/router';

import { SignInComponent } from './account/sign-in/sign-in.component';
import { HarborShellComponent } from './base/harbor-shell/harbor-shell.component';

import { BaseRoutingResolver } from './base/base-routing-resolver.service';

const harborRoutes: Routes = [
  {
    path: 'harbor',
    component: HarborShellComponent
  },
  { path: '', redirectTo: '/harbor', pathMatch: 'full' },
  { path: 'sign-in', component: SignInComponent }
];

@NgModule({
  imports: [
    RouterModule.forRoot(harborRoutes)
  ],
  exports: [RouterModule]
})
export class HarborRoutingModule {

}