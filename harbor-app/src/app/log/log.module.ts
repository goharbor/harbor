import { NgModule } from '@angular/core';
import { LogComponent } from './log.component';
import { SharedModule } from '../shared.module';

@NgModule({
  imports: [ SharedModule ],
  declarations: [ LogComponent ],
  exports: [ LogComponent ]
})
export class LogModule {}