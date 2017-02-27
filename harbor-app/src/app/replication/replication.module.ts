import { NgModule } from '@angular/core';

import { ReplicationComponent } from './replication.component';
import { CreateEditPolicyComponent } from './create-edit-policy/create-edit-policy.component';
import { ListPolicyComponent } from './list-policy/list-policy.component';
import { ListJobComponent } from './list-job/list-job.component';

import { CustomHighlightDirective } from './list-policy/custom-highlight.directive';

import { SharedModule } from '../shared/shared.module';
import { ReplicationService } from './replication.service';

@NgModule({
  imports: [ SharedModule ],
  declarations: [ 
    ReplicationComponent,
    CreateEditPolicyComponent,
    ListPolicyComponent,
    ListJobComponent,
    CustomHighlightDirective
  ],
  exports: [ ReplicationComponent ],
  providers: [ ReplicationService ]
})
export class ReplicationModule {}