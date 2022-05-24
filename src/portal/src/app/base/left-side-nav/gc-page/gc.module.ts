import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { GcPageComponent } from './gc-page.component';
import { GcComponent } from './gc/gc.component';
import { GcHistoryComponent } from './gc/gc-history/gc-history.component';
import { SharedModule } from '../../../shared/shared.module';

const routes: Routes = [
    {
        path: '',
        component: GcPageComponent,
    },
];
@NgModule({
    imports: [SharedModule, RouterModule.forChild(routes)],
    declarations: [GcPageComponent, GcComponent, GcHistoryComponent],
})
export class GcModule {}
