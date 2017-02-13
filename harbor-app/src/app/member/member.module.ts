import { NgModule } from '@angular/core';
import { MemberComponent } from './member.component';
import { SharedModule } from '../shared.module';

@NgModule({
  imports: [ SharedModule ],
  declarations: [ MemberComponent ],
  exports: [ MemberComponent ]
})
export class MemberModule {}