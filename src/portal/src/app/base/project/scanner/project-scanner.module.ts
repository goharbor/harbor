import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { SharedModule } from '../../../shared/shared.module';
import { ScannerComponent } from './scanner.component';

const routes: Routes = [
    {
        path: '',
        component: ScannerComponent,
    },
];
@NgModule({
    declarations: [ScannerComponent],
    imports: [RouterModule.forChild(routes), SharedModule],
})
export class ProjectScannerModule {}
