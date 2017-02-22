import { NgModule } from '@angular/core';
import { CoreModule } from '../core/core.module';
import { AccountModule } from '../account/account.module';

import { SessionService } from '../shared/session.service';
import { MessageComponent } from '../global-message/message.component';
import { MessageService } from '../global-message/message.service';
import { MaxLengthExtValidatorDirective } from './max-length-ext.directive';

@NgModule({
  imports: [
    CoreModule,
    AccountModule
  ],
  declarations: [
    MessageComponent,
    MaxLengthExtValidatorDirective
  ],
  exports: [
    CoreModule,
    AccountModule,
    MessageComponent,
    MaxLengthExtValidatorDirective
  ],
  providers: [SessionService, MessageService]
})
export class SharedModule {

}