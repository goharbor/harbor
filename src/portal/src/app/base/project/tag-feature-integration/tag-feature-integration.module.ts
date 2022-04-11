import { NgModule } from '@angular/core';
import { TagFeatureIntegrationComponent } from './tag-feature-integration.component';
import { TagRetentionComponent } from "./tag-retention/tag-retention.component";
import { SharedModule } from '../../../shared/shared.module';
import { AddRuleComponent } from "./tag-retention/add-rule/add-rule.component";
import { RouterModule, Routes } from '@angular/router';
import { TagRetentionService } from './tag-retention/tag-retention.service';
import { ImmutableTagComponent } from "./immutable-tag/immutable-tag.component";
import { ImmutableTagService } from "./immutable-tag/immutable-tag.service";
import { AddImmutableRuleComponent } from "./immutable-tag/add-rule/add-immutable-rule.component";
import { TagRetentionTasksComponent } from './tag-retention/tag-retention-tasks/tag-retention-tasks/tag-retention-tasks.component';
import { TagAccelerationComponent } from './tag-acceleration/tag-acceleration.component';
import { AddAccelerationRuleComponent } from './tag-acceleration/add-rule/add-acceleration-rule.component';


const routes: Routes = [
  {
    path: '',
    component: TagFeatureIntegrationComponent,
    children: [
      {
        path: 'tag-retention',
        component: TagRetentionComponent
      },
      {
        path: 'immutable-tag',
        component: ImmutableTagComponent
      },
      {
        path: 'tag-acceleration',
        component: TagAccelerationComponent
      },
      { path: '', redirectTo: 'tag-retention', pathMatch: 'full' },
    ]
  }
];
@NgModule({
  imports: [
    RouterModule.forChild(routes),
    SharedModule,
  ],
  declarations: [
    TagFeatureIntegrationComponent,
    TagRetentionComponent,
    AddRuleComponent,
    ImmutableTagComponent,
    AddImmutableRuleComponent,
    TagRetentionTasksComponent,
    TagAccelerationComponent,
    AddAccelerationRuleComponent
  ],
  providers: [
    TagRetentionService,
    ImmutableTagService
  ]
})
export class TagFeatureIntegrationModule { }
