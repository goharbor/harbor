import { NgModule } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CoreModule } from '../core/core.module';

import { SignInComponent } from './sign-in/sign-in.component';
import { PasswordSettingComponent } from './password/password-setting.component';
import { AccountSettingsModalComponent } from './account-settings/account-settings-modal.component';
import { SharedModule } from '../shared/shared.module';
import { SignUpComponent } from './sign-up/sign-up.component';
import { ForgotPasswordComponent } from './password/forgot-password.component';
import { ResetPasswordComponent } from './password/reset-password.component';
import { SignUpPageComponent } from './sign-up/sign-up-page.component';

import { PasswordSettingService } from './password/password-setting.service';
import { RepositoryModule } from '../repository/repository.module';

@NgModule({
  imports: [
    CoreModule,
    RouterModule,
    SharedModule,
    RepositoryModule
  ],
  declarations: [
    SignInComponent,
    PasswordSettingComponent,
    AccountSettingsModalComponent,
    SignUpComponent,
    ForgotPasswordComponent,
    ResetPasswordComponent,
    SignUpPageComponent],
  exports: [
    SignInComponent,
    PasswordSettingComponent,
    AccountSettingsModalComponent,
    ResetPasswordComponent,
    SignUpPageComponent],

  providers: [PasswordSettingService]
})
export class AccountModule { }