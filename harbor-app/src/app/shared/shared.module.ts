import { NgModule } from '@angular/core';
import { CoreModule } from '../core/core.module';
import { AccountModule } from '../account/account.module';

import { SessionService } from '../shared/session.service';
import { MessageComponent } from '../global-message/message.component';
import { MessageService } from '../global-message/message.service';

@NgModule({
  imports: [
    CoreModule,
    AccountModule
  ],
  declarations: [
    MessageComponent
  ],
  exports: [
    CoreModule,
    AccountModule,
    MessageComponent
  ],
  providers: [SessionService, MessageService]
})
export class SharedModule {

}