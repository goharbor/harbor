import { NgModule } from '@angular/core';
import { DistributionInstancesComponent } from './distribution-instances/distribution-instances.component';
import { DistributionSetupModalComponent } from './distribution-setup-modal/distribution-setup-modal.component';
import { SharedModule } from '../../../shared/shared.module';
import { RouterModule, Routes } from '@angular/router';

const routes: Routes = [
    {
        path: 'instances',
        component: DistributionInstancesComponent,
    },
    { path: '', redirectTo: 'instances', pathMatch: 'full' },
];
@NgModule({
    imports: [SharedModule, RouterModule.forChild(routes)],
    declarations: [
        DistributionSetupModalComponent,
        DistributionInstancesComponent,
    ],
})
export class DistributionModule {}
