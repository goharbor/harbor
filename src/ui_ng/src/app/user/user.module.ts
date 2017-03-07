import { NgModule } from '@angular/core';
import { SharedModule } from '../shared/shared.module';
import { UserComponent } from './user.component';
import { NewUserModalComponent } from './new-user-modal.component';
import { UserService } from './user.service';

@NgModule({
  imports: [
    SharedModule
  ],
  declarations: [
    UserComponent,
    NewUserModalComponent
  ],
  exports: [
    UserComponent
  ],
  providers:[UserService]
})
export class UserModule {

}