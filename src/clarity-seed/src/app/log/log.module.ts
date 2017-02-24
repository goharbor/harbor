import { NgModule } from '@angular/core';
import { AuditLogComponent } from './audit-log.component';
import { SharedModule } from '../shared/shared.module';
import { AuditLogService } from './audit-log.service';
@NgModule({
  imports: [ SharedModule ],
  declarations: [ AuditLogComponent ],
  providers: [ AuditLogService ],
  exports: [ AuditLogComponent ]
})
export class LogModule {}