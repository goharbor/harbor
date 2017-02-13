import { NgModule } from '@angular/core';
import { GlobalSearchComponent } from './global-search.component';
import { SharedModule } from '../shared.module';

@NgModule({
  imports: [ SharedModule ],
  declarations: [ GlobalSearchComponent ],
  exports: [ GlobalSearchComponent ]
})
export class GlobalSearchModule {}