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
import { AppConfigService } from '../../app-config.service';
import { SessionService } from '../../shared/session.service';
import { TagClickEvent } from 'harbor-ui';
import { Project } from '../../project/project';

@Component({
  selector: 'tag-repository',
  templateUrl: 'tag-repository.component.html',
  styleUrls: ['./tag-repository.component.css']
})
export class TagRepositoryComponent implements OnInit {

  projectId: number;
  repoName: string;
  hasProjectAdminRole: boolean = false;
  registryUrl: string;
  withNotary: boolean;
  hasSignedIn: boolean;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private appConfigService: AppConfigService,
    private session: SessionService) {
  }

  ngOnInit() {
    this.hasSignedIn = (this.session.getCurrentUser() !== null);
    let resolverData = this.route.snapshot.data;
    if (resolverData) {
      this.hasProjectAdminRole = (<Project>resolverData['projectResolver']).has_project_admin_role;
    }
    this.projectId = this.route.snapshot.params['id'];
    this.repoName = this.route.snapshot.params['repo'];

    this.registryUrl = this.appConfigService.getConfig().registry_url;
    this.withNotary = this.appConfigService.getConfig().with_notary;
  }

  watchTagClickEvt(tagEvt: TagClickEvent): void {
    let linkUrl = ['harbor', 'projects', tagEvt.project_id, 'repositories', tagEvt.repository_name, 'tags', tagEvt.tag_name];
    this.router.navigate(linkUrl);
  }
}