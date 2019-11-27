import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SharedModule } from '../../shared/shared.module';
import { ImmutableTagComponent } from './immutable-tag.component';
import { ImmutableTagService } from './immutable-tag.service';

import { TranslateModule } from '@ngx-translate/core';
import { AddRuleComponent } from './add-rule/add-rule.component';


@NgModule({
  declarations: [ImmutableTagComponent, AddRuleComponent],
  imports: [
    CommonModule,
    SharedModule,
    TranslateModule
  ],
  exports: [

  ],
  providers: [
    ImmutableTagService
  ]
})
export class ImmutableTagModule { }
