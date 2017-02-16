import { NgModule } from '@angular/core';
import { SharedModule } from '../shared/shared.module';
import { RouterModule } from '@angular/router';

import { SignInComponent } from './sign-in/sign-in.component';

@NgModule({
  imports: [
    SharedModule,
    RouterModule
  ],
  declarations: [SignInComponent],
  exports: [SignInComponent]
})
export class AccountModule { }