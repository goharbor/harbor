import { NgModule } from '@angular/core';
import { SharedModule } from '../shared/shared.module';
import { UserComponent } from './user.component';
@NgModule({
  imports: [
    SharedModule
  ],
  declarations: [
    UserComponent
  ],
  exports: [
    UserComponent
  ]
})
export class UserModule {

}