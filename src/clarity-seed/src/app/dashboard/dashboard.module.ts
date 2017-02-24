import { NgModule } from '@angular/core';
import { DashboardComponent } from './dashboard.component';
import { SharedModule } from '../shared/shared.module';

@NgModule({
  imports: [ SharedModule ],
  declarations: [ DashboardComponent ],
  exports: [ DashboardComponent ]
})
export class DashboardModule {}