// Copyright Project Harbor Authors
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
import { SharedModule } from '../shared/shared.module';
import { PasswordSettingComponent } from './password-setting/password-setting.component';
import { AccountSettingsModalComponent } from './account-settings/account-settings-modal.component';
import { SignUpComponent } from './sign-up/sign-up.component';
import { ForgotPasswordComponent } from './password-setting/forgot-password/forgot-password.component';
import { ResetPasswordComponent } from './password-setting/reset-password/reset-password.component';
import { SignUpPageComponent } from './sign-up/sign-up-page.component';
import { PasswordSettingService } from './password-setting/password-setting.service';
import { AccountSettingsModalService } from './account-settings/account-settings-modal-service.service';


@NgModule({
  imports: [
    CoreModule,
    RouterModule,
    SharedModule,
  ],
  declarations: [
    PasswordSettingComponent,
    AccountSettingsModalComponent,
    SignUpComponent,
    ForgotPasswordComponent,
    ResetPasswordComponent,
    SignUpPageComponent],
  exports: [
    PasswordSettingComponent,
    AccountSettingsModalComponent,
    ForgotPasswordComponent,
    ResetPasswordComponent,
    SignUpComponent,
    SignUpPageComponent],

  providers: [PasswordSettingService, AccountSettingsModalService]
})
export class AccountModule { }
