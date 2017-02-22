import { NgModule } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CoreModule } from '../core/core.module';

import { SignInComponent } from './sign-in/sign-in.component';
import { PasswordSettingComponent } from './password/password-setting.component';
import { AccountSettingsModalComponent } from './account-settings/account-settings-modal.component';
import { SharedModule } from '../shared/shared.module';

import { PasswordSettingService } from './password/password-setting.service';

@NgModule({
  imports: [
    CoreModule,
    RouterModule,
    SharedModule
  ],
  declarations: [SignInComponent, PasswordSettingComponent, AccountSettingsModalComponent],
  exports: [SignInComponent, PasswordSettingComponent, AccountSettingsModalComponent],
  
  providers: [PasswordSettingService]
})
export class AccountModule { }