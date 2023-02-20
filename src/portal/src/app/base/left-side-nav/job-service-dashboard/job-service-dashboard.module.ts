import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from 'src/app/shared/shared.module';
import { JobServiceDashboardComponent } from './job-service-dashboard.component';
import { PendingCardComponent } from './pending-job-card/pending-job-card.component';
import { PendingListComponent } from './pending-job-list/pending-job-list.component';
import { ScheduleCardComponent } from './schedule-card/schedule-card.component';
import { ScheduleListComponent } from './schedule-list/schedule-list.component';
import { DonutChartComponent } from './worker-card/donut-chart/donut-chart.component';
import { WorkerCardComponent } from './worker-card/worker-card.component';
import { WorkerListComponent } from './worker-list/worker-list.component';
import { JobServiceDashboardSharedDataService } from './job-service-dashboard-shared-data.service';

const routes: Routes = [
    {
        path: '',
        component: JobServiceDashboardComponent,
        children: [
            {
                path: 'pending-jobs',
                component: PendingListComponent,
            },
            {
                path: 'schedules',
                component: ScheduleListComponent,
            },
            {
                path: 'workers',
                component: WorkerListComponent,
            },
            {
                path: '',
                redirectTo: 'pending-jobs',
                pathMatch: 'full',
            },
        ],
    },
];
@NgModule({
    imports: [SharedModule, RouterModule.forChild(routes)],
    declarations: [
        JobServiceDashboardComponent,
        DonutChartComponent,
        PendingCardComponent,
        ScheduleCardComponent,
        WorkerCardComponent,
        PendingListComponent,
        ScheduleListComponent,
        WorkerListComponent,
    ],
    providers: [JobServiceDashboardSharedDataService],
})
export class JobServiceDashboardModule {}
