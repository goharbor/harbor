import { NgModule } from '@angular/core';
import { AuditLogComponent } from './audit-log.component';
import { SharedModule } from '../shared/shared.module';

@NgModule({
  imports: [ SharedModule ],
  declarations: [ AuditLogComponent ],
  exports: [ AuditLogComponent ]
})
export class LogModule {}