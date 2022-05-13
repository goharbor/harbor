import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';
import { ProjectConfigComponent } from './project-config.component';
import { ProjectPolicyConfigComponent } from './project-policy-config/project-policy-config.component';

const routes: Routes = [
    {
        path: '',
        component: ProjectConfigComponent,
    },
];
@NgModule({
    declarations: [ProjectConfigComponent, ProjectPolicyConfigComponent],
    imports: [RouterModule.forChild(routes), SharedModule],
})
export class ProjectConfigModule {}
