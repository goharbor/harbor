import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SystemRobotAccountsComponent } from './system-robot-accounts.component';
import { SharedModule } from '../shared/shared.module';
import { NewRobotComponent } from './new-robot/new-robot.component';
import { ProjectModule } from '../project/project.module';
import { ListAllProjectsComponent } from './list-all-projects/list-all-projects.component';
import { ProjectsModalComponent } from './projects-modal/projects-modal.component';



@NgModule({
  declarations: [
    SystemRobotAccountsComponent,
    NewRobotComponent,
    ListAllProjectsComponent,
    ProjectsModalComponent
  ],
  imports: [
    CommonModule,
    SharedModule,
    ProjectModule
  ]
})
export class SystemRobotAccountsModule { }
