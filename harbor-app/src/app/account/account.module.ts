import { NgModule } from '@angular/core';
import { RouterModule } from '@angular/router';
import { CoreModule } from '../core/core.module';

import { SignInComponent } from './sign-in/sign-in.component';
import { PasswordSettingComponent } from './password/password-setting.component';

import { PasswordSettingService } from './password/password-setting.service';

@NgModule({
  imports: [
    CoreModule,
    RouterModule
  ],
  declarations: [SignInComponent, PasswordSettingComponent],
  exports: [SignInComponent, PasswordSettingComponent],
  
  providers: [PasswordSettingService]
})
export class AccountModule { }