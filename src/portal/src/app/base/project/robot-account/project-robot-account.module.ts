import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';
import { RobotAccountComponent } from './robot-account.component';
import { AddRobotComponent } from './add-robot/add-robot.component';

const routes: Routes = [
    {
        path: '',
        component: RobotAccountComponent,
    },
];
@NgModule({
    declarations: [AddRobotComponent, RobotAccountComponent],
    imports: [RouterModule.forChild(routes), SharedModule],
})
export class ProjectRobotAccountModule {}
