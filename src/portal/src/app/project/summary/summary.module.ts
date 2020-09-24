import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SummaryComponent } from './summary.component';
import { TranslateModule } from '@ngx-translate/core';

@NgModule({
  declarations: [SummaryComponent],
  imports: [
    CommonModule,
    TranslateModule
  ]
})
export class SummaryModule { }
