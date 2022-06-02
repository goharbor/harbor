import { NgModule } from '@angular/core';
import { SummaryComponent } from './summary.component';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';

const routes: Routes = [
    {
        path: '',
        component: SummaryComponent,
    },
];
@NgModule({
    declarations: [SummaryComponent],
    imports: [RouterModule.forChild(routes), SharedModule],
})
export class SummaryModule {}
