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
import { RouterModule } from '@angular/router';

import { SharedModule } from '../shared/shared.module';

import { RepositoryPageComponent } from './repository-page.component';
import { ArtifactListPageComponent } from './artifact-list-page/artifact-list-page.component';
import { TopRepoComponent } from './top-repo/top-repo.component';
import { ArtifactSummaryPageComponent } from './artifact-summary-page/artifact-summary-page.component';

@NgModule({
  imports: [
    SharedModule,
    RouterModule
  ],
  declarations: [
    RepositoryPageComponent,
    ArtifactListPageComponent,
    TopRepoComponent,
    ArtifactSummaryPageComponent
  ],
  exports: [
    RepositoryPageComponent,
    TopRepoComponent,
    ArtifactSummaryPageComponent
  ],
  providers: []
})
export class RepositoryModule { }
