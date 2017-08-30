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
import { Router, NavigationExtras } from '@angular/router';
import { Project } from '../project';
import { ProjectService } from '../project.service';

import { SessionService } from '../../shared/session.service';
import { SearchTriggerService } from '../../base/global-search/search-trigger.service';
import { ProjectTypes, RoleInfo } from '../../shared/shared.const';

import { State } from 'clarity-angular';

@Component({
  selector: 'list-project',
  templateUrl: 'list-project.component.html',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListProjectComponent {
  _filterType: string = ProjectTypes[0];

  @Input() loading: boolean = true;
  @Input() projects: Project[];
  @Input()
  get filteredType(): string {
    return this._filterType;
  }
  set filteredType(value: string) {
    if (value && value.trim() !== "") {
      this._filterType = value;
    }
  }

  @Output() paginate = new EventEmitter<State>();
  @Output() toggle = new EventEmitter<Project>();
  @Output() delete = new EventEmitter<Project>();

  roleInfo = RoleInfo;

  constructor(
    private session: SessionService,
    private router: Router,
    private searchTrigger: SearchTriggerService,
    private ref: ChangeDetectorRef) {
    let hnd = setInterval(() => ref.markForCheck(), 100);
    setTimeout(() => clearInterval(hnd), 1000);
  }

  get showRoleInfo(): boolean {
    return this.filteredType !== ProjectTypes[2];
  }

  public get isSystemAdmin(): boolean {
    let account = this.session.getCurrentUser();
    return account != null && account.has_admin_role > 0;
  }

  goToLink(proId: number): void {
    this.searchTrigger.closeSearch(true);

    let linkUrl = ['harbor', 'projects', proId, 'repositories'];
    this.router.navigate(linkUrl);
  }

  refresh(state: State) {
    this.paginate.emit(state);
  }

  newReplicationRule(p: Project) {
    if (p) {
      this.router.navigateByUrl(`/harbor/projects/${p.project_id}/replications?is_create=true`);
    }
  }

  toggleProject(p: Project) {
    this.toggle.emit(p);
  }

  deleteProject(p: Project) {
    this.delete.emit(p);
  }

}