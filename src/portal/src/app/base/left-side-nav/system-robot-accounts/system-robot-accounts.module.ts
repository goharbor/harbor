import { NgModule } from '@angular/core';
import { SystemRobotAccountsComponent } from './system-robot-accounts.component';
import { SharedModule } from '../../../shared/shared.module';
import { NewRobotComponent } from './new-robot/new-robot.component';
import { ListAllProjectsComponent } from './list-all-projects/list-all-projects.component';
import { ProjectsModalComponent } from './projects-modal/projects-modal.component';
import { RouterModule, Routes } from '@angular/router';

const routes: Routes = [
    {
        path: '',
        component: SystemRobotAccountsComponent,
    },
];

@NgModule({
    declarations: [
        SystemRobotAccountsComponent,
        NewRobotComponent,
        ListAllProjectsComponent,
        ProjectsModalComponent,
    ],
    imports: [SharedModule, RouterModule.forChild(routes)],
})
export class SystemRobotAccountsModule {}
