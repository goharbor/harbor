import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TagFeatureIntegrationComponent } from './tag-feature-integration.component';
import { TranslateModule } from '@ngx-translate/core';
import { TagRetentionComponent } from "./tag-retention/tag-retention.component";
import { ImmutableTagModule } from "./immutable-tag/immutable-tag.module";
import { ClarityModule } from '@clr/angular';
import { SharedModule } from '../../shared/shared.module';
import { AddRuleComponent } from "./tag-retention/add-rule/add-rule.component";
import { RouterModule } from '@angular/router';
import { TagRetentionService } from './tag-retention/tag-retention.service';



@NgModule({
    declarations: [TagFeatureIntegrationComponent, TagRetentionComponent, AddRuleComponent],
  imports: [
    CommonModule,
    TranslateModule,
    ImmutableTagModule,
    ClarityModule,
    SharedModule,
    RouterModule
  ],
  providers: [
    TagRetentionService
  ]
})
export class TagFeatureIntegrationModule { }
