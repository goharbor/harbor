import { NgModule } from '@angular/core';
import { CoreModule } from '../core/core.module';

@NgModule({
  imports: [
    CoreModule
  ],
  exports: [
    CoreModule
  ]
})
export class SharedModule {

}