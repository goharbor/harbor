// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { Component, EventEmitter, Output, Input, ChangeDetectionStrategy, ChangeDetectorRef } from '@angular/core';
import { Router } from '@angular/router';
import { State } from '../../../lib/services/interface';

import { SearchTriggerService } from '../../base/global-search/search-trigger.service';
import { Project } from '../../project/project';

@Component({
  selector: 'list-project-ro',
  templateUrl: 'list-project-ro.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListProjectROComponent {
  @Input() projects: Project[];
  @Output() paginate = new EventEmitter<State>();


  constructor(
    private searchTrigger: SearchTriggerService,
    private router: Router) {}

  goToLink(proId: number): void {
    this.searchTrigger.closeSearch(true);

    let linkUrl = ['harbor', 'projects', proId, 'repositories'];
    this.router.navigate(linkUrl);
  }

  refresh(state: State) {
    this.paginate.emit(state);
  }
}
