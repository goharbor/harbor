import { NgModule } from '@angular/core';
import { CoreModule } from '../core/core.module';

import { SessionService } from '../shared/session.service';

@NgModule({
  imports: [
    CoreModule
  ],
  exports: [
    CoreModule
  ],
  providers: [SessionService]
})
export class SharedModule {

}