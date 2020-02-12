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
import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { AppConfigService } from '../../app-config.service';
import { SessionService } from '../../shared/session.service';
import { Project } from '../../project/project';
import { ArtifactListComponent } from "../../../lib/components/artifact-list/artifact-list.component";
import { ArtifactClickEvent, ArtifactService } from "../../../lib/services";
import { clone } from '../../../lib/utils/utils';

@Component({
  selector: 'artifact-list-page',
  templateUrl: 'artifact-list-page.component.html',
  styleUrls: ['./artifact-list-page.component.scss']
})
export class ArtifactListPageComponent implements OnInit {

  projectId: number;
  projectMemberRoleId: number;
  repoName: string;
  referArtifactNameArray: string[] = [];
  hasProjectAdminRole: boolean = false;
  isGuest: boolean;
  registryUrl: string;

  @ViewChild(ArtifactListComponent, {static: false})
  repositoryComponent: ArtifactListComponent;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private artifactService: ArtifactService,
    private appConfigService: AppConfigService,
    private session: SessionService) {
  }

  ngOnInit() {
    this.projectId = this.route.snapshot.params['id'];
    if (!this.projectId) {
      this.projectId = this.route.snapshot.parent.params['id'];
    }

    let resolverData = this.route.snapshot.data;

    if (resolverData) {
      this.hasProjectAdminRole = (<Project>resolverData['projectResolver']).has_project_admin_role;
      this.isGuest = (<Project>resolverData['projectResolver']).current_user_role_id === 3;
      this.projectMemberRoleId = (<Project>resolverData['projectResolver']).current_user_role_id;
    }
    this.repoName = this.route.snapshot.params['repo'];
    this.registryUrl = this.appConfigService.getConfig().registry_url;
    this.referArtifactNameArray = JSON.parse(sessionStorage.getItem('reference')) || [];

  }

  get withNotary(): boolean {
    return this.appConfigService.getConfig().with_notary;
  }
  get withAdmiral(): boolean {
    return this.appConfigService.getConfig().with_admiral;
  }

  get hasSignedIn(): boolean {
    return this.session.getCurrentUser() !== null;
  }

  hasChanges(): boolean {
    return this.repositoryComponent.hasChanges();
  }

  watchTagClickEvt(artifactEvt: ArtifactClickEvent): void {
    //  
    sessionStorage.setItem('referenceSummary', JSON.stringify(this.referArtifactNameArray));
    
    let linkUrl = ['harbor', 'projects', artifactEvt.project_id, 'repositories'
    , artifactEvt.repository_name, 'artifacts', artifactEvt.digest];
    this.router.navigate(linkUrl);
  }

  watchGoBackEvt(projectId: string| number): void {
    this.router.navigate(["harbor", "projects", projectId, "repositories"]);
  }
  goProBack(): void {
    this.router.navigate(["harbor", "projects"]);
  }
  backInitRepo() {
    this.referArtifactNameArray = [];
    sessionStorage.setItem('reference', JSON.stringify([]));

    this.updateArtifactList('repoName');
  }
  jumpDigest(referArtifactNameArray: string[], index: number) {
    this.referArtifactNameArray = referArtifactNameArray.slice(index);
    this.referArtifactNameArray.pop();
    this.referArtifactNameArray = referArtifactNameArray.slice(index);
    sessionStorage.setItem('reference', JSON.stringify(referArtifactNameArray.slice(index)));

    this.updateArtifactList(referArtifactNameArray.slice(index));
  }
  updateArtifactList(res): void {
      this.artifactService.triggerUploadArtifact.next(res);
  }
  putArtifactReferenceArr(digestArray) {
    this.referArtifactNameArray = digestArray;
  }
}
