import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { GcComponent } from './gc-page/gc/gc.component';
import { GcHistoryComponent } from './gc-page/gc/gc-history/gc-history.component';
import { SharedModule } from '../../../shared/shared.module';
import { SetJobComponent } from './audit-log-purge/set-job/set-job.component';
import { ClearingJobComponent } from './clearing-job.component';
import { PurgeHistoryComponent } from './audit-log-purge/history/purge-history.component';

const routes: Routes = [
    {
        path: '',
        component: ClearingJobComponent,
        children: [
            {
                path: 'gc',
                component: GcComponent,
            },
            {
                path: 'audit-log-purge',
                component: SetJobComponent,
            },
            {
                path: '',
                redirectTo: 'gc',
                pathMatch: 'full',
            },
        ],
    },
];
@NgModule({
    imports: [SharedModule, RouterModule.forChild(routes)],
    declarations: [
        GcComponent,
        GcHistoryComponent,
        ClearingJobComponent,
        SetJobComponent,
        PurgeHistoryComponent,
    ],
})
export class ClearingJobModule {}
