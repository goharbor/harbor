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
import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import {AppConfigService} from "../../app-config.service";

@Component({
  selector: 'artifact-summary-page',
  templateUrl: 'artifact-summary-page.component.html',
  styleUrls: ["artifact-summary-page.component.scss"]
})
export class ArtifactSummaryPageComponent implements OnInit, OnDestroy {
  tagId: string;
  artifactDigest: string;
  repositoryName: string;
  projectId: string | number;
  referArtifactNameArray: string[];
  constructor(
    private route: ActivatedRoute,
    private appConfigService: AppConfigService,
    private router: Router
  ) {
  }

  ngOnInit(): void {
    this.repositoryName = this.route.snapshot.params["repo"];
    this.artifactDigest = this.route.snapshot.params["digest"];
    this.projectId = this.route.snapshot.params["id"];
    this.referArtifactNameArray = JSON.parse(sessionStorage.getItem('referenceSummary')) || [];
  }

  get withAdmiral(): boolean {
    return this.appConfigService.getConfig().with_admiral;
  }

  goBack(repositoryName: string): void {
    this.router.navigate(["harbor", "projects", this.projectId, "repositories", repositoryName]);
  }
  goBackRep(): void {
    this.router.navigate(["harbor", "projects", this.projectId, "repositories"]);
  }
  goBackPro(): void {
    this.router.navigate(["harbor", "projects"]);
  }
  ngOnDestroy(): void {
    sessionStorage.setItem('referenceSummary', JSON.stringify([]));
  }
  jumpDigest(referArtifactNameArray: string[], index: number) {
    sessionStorage.setItem('referenceSummary', JSON.stringify([]));
    sessionStorage.setItem('reference', JSON.stringify(referArtifactNameArray.slice(index)));

    this.router.navigate(["harbor", "projects", this.projectId, "repositories", this.repositoryName]);
  }
}
