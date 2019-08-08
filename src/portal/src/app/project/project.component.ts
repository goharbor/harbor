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
import { Component, OnInit, ViewChild } from '@angular/core';
import { CreateProjectComponent } from './create-project/create-project.component';
import { ListProjectComponent } from './list-project/list-project.component';
import { ProjectTypes } from '../shared/shared.const';
import { ConfigurationService } from '../config/config.service';
import { Configuration, QuotaHardInterface } from '@harbor/ui';
import { SessionService } from "../shared/session.service";

@Component({
  selector: 'project',
  templateUrl: 'project.component.html',
  styleUrls: ['./project.component.scss']
})
export class ProjectComponent implements OnInit {
  projectTypes = ProjectTypes;
  quotaObj: QuotaHardInterface;
  @ViewChild(CreateProjectComponent, {static: false})
  creationProject: CreateProjectComponent;

  @ViewChild(ListProjectComponent, {static: false})
  listProject: ListProjectComponent;

  currentFilteredType: number = 0; // all projects
  projectName: string = "";

  loading: boolean = true;

  get selecteType(): number {
    return this.currentFilteredType;
  }
  set selecteType(_project: number) {
    this.currentFilteredType = _project;
    if (window.sessionStorage) {
      window.sessionStorage['projectTypeValue'] = _project;
    }
  }

  constructor(
    public configService: ConfigurationService,
    private session: SessionService
  ) { }

  ngOnInit(): void {
    if (window.sessionStorage && window.sessionStorage['projectTypeValue'] && window.sessionStorage['fromDetails']) {
      this.currentFilteredType = +window.sessionStorage['projectTypeValue'];
      window.sessionStorage.removeItem('fromDetails');
    }
    if (this.isSystemAdmin) {
      this.getConfigration();
    }
  }
  getConfigration() {
    this.configService.getConfiguration()
      .subscribe((configurations: Configuration) => {
        this.quotaObj = {
          count_per_project: configurations.count_per_project ? configurations.count_per_project.value : -1,
          storage_per_project: configurations.storage_per_project ? configurations.storage_per_project.value : -1
        };
      });
  }
  public get isSystemAdmin(): boolean {
    let account = this.session.getCurrentUser();
    return account != null && account.has_admin_role;
  }
  openModal(): void {
    this.creationProject.newProject();
  }

  createProject(created: boolean) {
    if (created) {
      this.refresh();
    }
  }

  doSearchProjects(projectName: string): void {
    this.projectName = projectName;
    this.listProject.doSearchProject(this.projectName);
  }

  doFilterProjects(): void {
    this.listProject.doFilterProject(this.selecteType);
  }

  refresh(): void {
    this.currentFilteredType = 0;
    this.projectName = "";
    this.listProject.refresh();
  }

}
