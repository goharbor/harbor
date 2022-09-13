import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';
import { RepositoryGridviewComponent } from './repository-gridview.component';
import { GridViewComponent } from './gridview/grid-view.component';

const routes: Routes = [
    {
        path: '',
        component: RepositoryGridviewComponent,
    },
];
@NgModule({
    declarations: [RepositoryGridviewComponent, GridViewComponent],
    imports: [RouterModule.forChild(routes), SharedModule],
})
export class RepositoryModule {}
