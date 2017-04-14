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