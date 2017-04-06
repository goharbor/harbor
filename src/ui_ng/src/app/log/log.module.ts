import { NgModule } from '@angular/core';
import { AuditLogComponent } from './audit-log.component';
import { SharedModule } from '../shared/shared.module';
import { AuditLogService } from './audit-log.service';
import { RecentLogComponent } from './recent-log.component';

@NgModule({
  imports: [SharedModule],
  declarations: [
    AuditLogComponent,
    RecentLogComponent],
  providers: [AuditLogService],
  exports: [
    AuditLogComponent,
    RecentLogComponent]
})
export class LogModule { }