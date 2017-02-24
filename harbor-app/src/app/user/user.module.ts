import { NgModule } from '@angular/core';
import { SharedModule } from '../shared/shared.module';
import { UserComponent } from './user.component';
import { NewUserFormComponent } from './new-user-form.component';
import { NewUserModalComponent } from './new-user-modal.component';

@NgModule({
  imports: [
    SharedModule
  ],
  declarations: [
    UserComponent,
    NewUserFormComponent,
    NewUserModalComponent
  ],
  exports: [
    UserComponent,
    NewUserFormComponent
  ]
})
export class UserModule {

}