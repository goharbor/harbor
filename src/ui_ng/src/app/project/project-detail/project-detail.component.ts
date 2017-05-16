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
import { Component } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

import { Project } from '../project';

import { SessionService } from '../../shared/session.service';
import { ProjectService } from '../../project/project.service';

import { RoleMapping } from '../../shared/shared.const';

@Component({
    selector: 'project-detail',
    templateUrl: "project-detail.component.html",
    styleUrls: [ 'project-detail.component.css' ]
})
export class ProjectDetailComponent {

  hasSignedIn: boolean;
  currentProject: Project;
  
  isMember: boolean;
  roleName: string;

  constructor(
    private route: ActivatedRoute, 
    private router: Router,
    private sessionService: SessionService,
    private projectService: ProjectService) {

    this.hasSignedIn = this.sessionService.getCurrentUser() !== null;
    this.route.data.subscribe(data=>{
      this.currentProject = <Project>data['projectResolver'];
      this.isMember = this.currentProject.is_member;
      this.roleName = RoleMapping[this.currentProject.role_name];
    });
  }

  public get isSystemAdmin(): boolean {
    let account = this.sessionService.getCurrentUser();
    return account != null && account.has_admin_role > 0;
  }

  public get isSessionValid(): boolean {
    return this.sessionService.getCurrentUser() != null;
  }

}