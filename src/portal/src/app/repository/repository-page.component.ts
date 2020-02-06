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
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { Project } from '../project/project';
import { SessionService } from '../shared/session.service';
import { RepositoryItem } from "../../lib/services";
@Component({
  selector: 'repository',
  templateUrl: 'repository-page.component.html'
})
export class RepositoryPageComponent implements OnInit {
  projectId: number;
  hasProjectAdminRole: boolean;
  hasSignedIn: boolean;
  projectName: string;
  mode = 'standalone';

  constructor(
    private route: ActivatedRoute,
    private session: SessionService,
    private router: Router,
  ) {
  }

  ngOnInit(): void {
    this.projectId = this.route.snapshot.parent.params['id'];
    let resolverData = this.route.snapshot.parent.data;
    if (resolverData) {
      let pro: Project = <Project>resolverData['projectResolver'];
      this.hasProjectAdminRole = pro.has_project_admin_role;
      this.projectName = pro.name;
    }
    this.hasSignedIn = this.session.getCurrentUser() !== null;
  }

  watchRepoClickEvent(repoEvt: RepositoryItem): void {
    let linkUrl = ['harbor', 'projects', repoEvt.project_id, 'repositories', repoEvt.name.split('/')[1]];
    this.router.navigate(linkUrl);
  }
}
