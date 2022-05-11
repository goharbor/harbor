import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';
import { ProjectLabelComponent } from './project-label.component';

const routes: Routes = [
    {
        path: '',
        component: ProjectLabelComponent,
    },
];
@NgModule({
    declarations: [ProjectLabelComponent],
    imports: [RouterModule.forChild(routes), SharedModule],
})
export class ProjectLabelModule {}
