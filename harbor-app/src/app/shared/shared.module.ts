import { NgModule } from '@angular/core';
import { CoreModule } from '../core/core.module';
//import { AccountModule } from '../account/account.module';

import { SessionService } from '../shared/session.service';
import { MessageComponent } from '../global-message/message.component';
import { MessageService } from '../global-message/message.service';
import { MaxLengthExtValidatorDirective } from './max-length-ext.directive';
import { FilterComponent } from './filter/filter.component';
import { DatagridActionOverflow } from './clg-dg-action-overflow/datagrid-action-overflow';

@NgModule({
  imports: [
    CoreModule
  ],
  declarations: [
    MessageComponent,
    MaxLengthExtValidatorDirective,
    FilterComponent,
    DatagridActionOverflow
  ],
  exports: [
    CoreModule,
    MessageComponent,
    MaxLengthExtValidatorDirective,
    FilterComponent,
    DatagridActionOverflow
  ],
  providers: [SessionService, MessageService]
})
export class SharedModule {

}