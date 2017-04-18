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
import { Component, EventEmitter, Output, Input, OnInit } from '@angular/core';
import { Router, NavigationExtras } from '@angular/router';
import { Project } from '../project';
import { ProjectService } from '../project.service';

import { SessionService } from '../../shared/session.service';
import { SearchTriggerService } from '../../base/global-search/search-trigger.service';
import { ProjectTypes, RoleInfo } from '../../shared/shared.const';

import { State } from 'clarity-angular';

@Component({
  selector: 'list-project',
  templateUrl: 'list-project.component.html'
})
export class ListProjectComponent implements OnInit {

  @Input() projects: Project[];


  @Input() totalPage: number;
  @Input() totalRecordCount: number;
  pageOffset: number = 1;

  @Input() filteredType: string;

  @Output() paginate = new EventEmitter<State>();

  @Output() toggle = new EventEmitter<Project>();
  @Output() delete = new EventEmitter<Project>();

  roleInfo = RoleInfo;

  constructor(
    private session: SessionService,
    private router: Router,
    private searchTrigger: SearchTriggerService) { }

  ngOnInit(): void {
  }

  get showRoleInfo(): boolean {
    return this.filteredType === ProjectTypes[0];
  }

  public get isSystemAdmin(): boolean {
    let account = this.session.getCurrentUser();
    return account != null && account.has_admin_role > 0;
  }

  goToLink(proId: number): void {
    this.searchTrigger.closeSearch(true);

    let linkUrl = ['harbor', 'projects', proId, 'repository'];
    this.router.navigate(linkUrl);
  }

  refresh(state: State) {
    this.paginate.emit(state);
  }

  newReplicationRule(p: Project) {
    if(p) {
      this.router.navigateByUrl(`/harbor/projects/${p.project_id}/replication?is_create=true`);
    }
  }

  toggleProject(p: Project) {
    this.toggle.emit(p);
  }

  deleteProject(p: Project) {
    this.delete.emit(p);
  }

}