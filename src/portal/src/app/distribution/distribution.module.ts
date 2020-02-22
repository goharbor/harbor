import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { DistributionHistoryComponent } from './distribution-history/distribution-history.component';
import { DistributionInstancesComponent } from './distribution-instances/distribution-instances.component';
// import { DistributionProvidersComponent } from './distribution-providers.bak/distribution-providers.component';
import { DistributionSetupModalComponent } from './distribution-setup-modal/distribution-setup-modal.component';
import { DistributionService } from './distribution.service';
import { SharedModule } from '../shared/shared.module';
import { MsgChannelService } from './msg-channel.service';

@NgModule({
  imports: [CommonModule, SharedModule],
  declarations: [
    DistributionHistoryComponent,
    DistributionSetupModalComponent,
    DistributionInstancesComponent
  ],
  providers: [DistributionService, MsgChannelService]
})
export class DistributionModule {}
