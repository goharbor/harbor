import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SignInComponent } from './sign-in.component';
import { AccountModule } from '../account/account.module';
import { BaseModule } from '../base/base.module';
import { SharedModule } from '../shared/shared.module';
import { RepositoryModule } from '../repository/repository.module';

@NgModule({
  declarations: [
    SignInComponent,

  ],
  imports: [
    CommonModule,
    AccountModule,
    SharedModule,
    BaseModule,
    RepositoryModule
  ]
})
export class SignInModule { }
