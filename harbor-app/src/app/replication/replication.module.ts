import { NgModule } from '@angular/core';
import { ReplicationComponent } from './replication.component';
import { SharedModule } from '../shared.module';

@NgModule({
  imports: [ SharedModule ],
  declarations: [ ReplicationComponent ],
  exports: [ ReplicationComponent ]
})
export class ReplicationModule {}