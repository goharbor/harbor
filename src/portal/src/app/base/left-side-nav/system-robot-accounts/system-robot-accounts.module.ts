// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
