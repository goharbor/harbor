import { NgModule } from '@angular/core';
import { LicenseComponent } from './license.component';
import { RouterModule, Routes } from '@angular/router';
import { CommonModule } from '@angular/common';
const routes: Routes = [
    {
        path: '',
        component: LicenseComponent,
    },
];
@NgModule({
    imports: [CommonModule, RouterModule.forChild(routes)],
    declarations: [LicenseComponent],
})
export class LicenseModule {}
