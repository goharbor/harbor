import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';
import { AuditLogComponent } from './audit-log.component';

const routes: Routes = [
    {
        path: '',
        component: AuditLogComponent,
    },
];
@NgModule({
    declarations: [AuditLogComponent],
    imports: [RouterModule.forChild(routes), SharedModule],
})
export class AuditLogModule {}
