import { NgModule } from '@angular/core';
import { CoreModule } from '../core/core.module';
//import { AccountModule } from '../account/account.module';

import { SessionService } from '../shared/session.service';
import { MessageComponent } from '../global-message/message.component';
import { MessageService } from '../global-message/message.service';
import { MaxLengthExtValidatorDirective } from './max-length-ext.directive';
import { FilterComponent } from './filter/filter.component';
import { HarborActionOverflow } from './harbor-action-overflow/harbor-action-overflow';

@NgModule({
  imports: [
    CoreModule
  ],
  declarations: [
    MessageComponent,
    MaxLengthExtValidatorDirective,
    FilterComponent,
    HarborActionOverflow
  ],
  exports: [
    CoreModule,
    MessageComponent,
    MaxLengthExtValidatorDirective,
    FilterComponent,
    HarborActionOverflow
  ],
  providers: [SessionService, MessageService]
})
export class SharedModule {

}