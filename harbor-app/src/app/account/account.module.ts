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

import { PasswordSettingService } from './password/password-setting.service';

@NgModule({
  imports: [
    CoreModule,
    RouterModule,
    SharedModule
  ],
  declarations: [
    SignInComponent,
    PasswordSettingComponent,
    AccountSettingsModalComponent,
    SignUpComponent,
    ForgotPasswordComponent,
    ResetPasswordComponent],
  exports: [
    SignInComponent,
    PasswordSettingComponent,
    AccountSettingsModalComponent,
    ResetPasswordComponent],

  providers: [PasswordSettingService]
})
export class AccountModule { }