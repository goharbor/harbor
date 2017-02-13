import { NgModule } from '@angular/core';
import { SignInComponent } from './sign-in.component';
import { SharedModule } from '../shared.module';
import { RouterModule } from '@angular/router'; 

@NgModule({
  imports: [ 
    SharedModule,
    RouterModule
  ],
  declarations: [ SignInComponent ],
  exports: [SignInComponent]
})
export class AccountModule {}