import { NgModule } from '@angular/core';
import { NavigatorComponent } from './navigator.component';

import { SharedModule } from '../shared.module';
import { GlobalSearchComponent } from '../global-search/global-search.component';

import { GlobalSearchModule } from '../global-search/global-search.module';
import { RouterModule } from '@angular/router';

@NgModule({
  imports: [
    SharedModule,
    GlobalSearchModule,
    RouterModule
  ],
  declarations: [ NavigatorComponent ],
  exports: [ NavigatorComponent ]
})
export class NavigatorModule {}