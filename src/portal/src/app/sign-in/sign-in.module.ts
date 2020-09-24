import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SignInComponent } from './sign-in.component';
import { AccountModule } from '../account/account.module';
import { BaseModule } from '../base/base.module';
import { SharedModule } from '../shared/shared.module';
import { TopRepoComponent } from "./top-repo/top-repo.component";
import { TopRepoService } from "./top-repo/top-repository.service";

@NgModule({
  declarations: [
    SignInComponent,
    TopRepoComponent
  ],
  imports: [
    CommonModule,
    AccountModule,
    SharedModule,
    BaseModule
  ],
  providers: [
    TopRepoService
  ]
})
export class SignInModule { }
