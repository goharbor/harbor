import { NgModule } from '@angular/core';
import { DevCenterComponent } from './dev-center.component';
import { RouterModule, Routes } from '@angular/router';

const routes: Routes = [
    {
        path: '',
        component: DevCenterComponent,
    },
];
@NgModule({
    imports: [RouterModule.forChild(routes)],
    declarations: [DevCenterComponent],
})
export class DeveloperCenterModule {}
