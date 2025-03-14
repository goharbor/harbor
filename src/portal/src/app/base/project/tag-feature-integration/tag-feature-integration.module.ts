// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { NgModule } from '@angular/core';
import { TagFeatureIntegrationComponent } from './tag-feature-integration.component';
import { TagRetentionComponent } from './tag-retention/tag-retention.component';
import { SharedModule } from '../../../shared/shared.module';
import { AddRuleComponent } from './tag-retention/add-rule/add-rule.component';
import { RouterModule, Routes } from '@angular/router';
import { TagRetentionService } from './tag-retention/tag-retention.service';
import { ImmutableTagComponent } from './immutable-tag/immutable-tag.component';
import { ImmutableTagService } from './immutable-tag/immutable-tag.service';
import { AddImmutableRuleComponent } from './immutable-tag/add-rule/add-immutable-rule.component';
import { TagRetentionTasksComponent } from './tag-retention/tag-retention-tasks/tag-retention-tasks/tag-retention-tasks.component';
import { USERSTATICPERMISSION } from '../../../shared/services';
import { TagFeatureGuardService } from './tag-feature-guard.service';

const routes: Routes = [
    {
        path: '',
        component: TagFeatureIntegrationComponent,
        children: [
            {
                path: 'tag-retention',
                canActivate: [TagFeatureGuardService],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.TAG_RETENTION.KEY,
                        action: USERSTATICPERMISSION.TAG_RETENTION.VALUE.READ,
                    },
                },
                component: TagRetentionComponent,
            },
            {
                path: 'immutable-tag',
                canActivate: [TagFeatureGuardService],
                data: {
                    permissionParam: {
                        resource: USERSTATICPERMISSION.IMMUTABLE_TAG.KEY,
                        action: USERSTATICPERMISSION.IMMUTABLE_TAG.VALUE.LIST,
                    },
                },
                component: ImmutableTagComponent,
            },
            { path: '', redirectTo: 'tag-retention', pathMatch: 'full' },
        ],
    },
];
@NgModule({
    imports: [RouterModule.forChild(routes), SharedModule],
    declarations: [
        TagFeatureIntegrationComponent,
        TagRetentionComponent,
        AddRuleComponent,
        ImmutableTagComponent,
        AddImmutableRuleComponent,
        TagRetentionTasksComponent,
    ],
    providers: [TagRetentionService, ImmutableTagService],
})
export class TagFeatureIntegrationModule {}
