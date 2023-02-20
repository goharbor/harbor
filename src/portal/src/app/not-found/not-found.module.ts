import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../shared/shared.module';
import { PageNotFoundComponent } from './not-found.component';
const routes: Routes = [
    {
        path: '',
        component: PageNotFoundComponent,
    },
];
@NgModule({
    imports: [SharedModule, RouterModule.forChild(routes)],
    declarations: [PageNotFoundComponent],
})
export class NotFoundModule {}
