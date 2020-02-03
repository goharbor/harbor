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
import { Component, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { CreateProjectComponent } from './create-project/create-project.component';
import { ListProjectComponent } from './list-project/list-project.component';
import { ProjectTypes } from '../shared/shared.const';
import { ConfigurationService } from '../config/config.service';
import { SessionService } from "../shared/session.service";
import { ProjectService, QuotaHardInterface, Repository, RequestQueryParams } from "../../lib/services";
import { Configuration } from "../../lib/components/config/config";
import { FilterComponent } from '../../lib/components/filter/filter.component';
import { Subscription } from 'rxjs';
import { debounceTime, distinctUntilChanged, finalize, switchMap } from 'rxjs/operators';
import { calculatePage, doFiltering, doSorting } from '../../lib/utils/utils';
import { Project } from './project';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';

@Component({
  selector: 'project',
  templateUrl: 'project.component.html',
  styleUrls: ['./project.component.scss']
})
export class ProjectComponent implements OnInit, OnDestroy {
  projectTypes = ProjectTypes;
  quotaObj: QuotaHardInterface;
  @ViewChild(CreateProjectComponent, {static: false})
  creationProject: CreateProjectComponent;

  @ViewChild(ListProjectComponent, {static: false})
  listProject: ListProjectComponent;

  currentFilteredType: number = 0; // all projects
  projectName: string = "";

  loading: boolean = true;

  get selecteType(): string {
    return this.currentFilteredType + "";
  }
  set selecteType(_project: string) {
    this.currentFilteredType = +_project;
    if (window.sessionStorage) {
      window.sessionStorage['projectTypeValue'] = +_project;
    }
  }
  @ViewChild(FilterComponent, {static: true})
  filterComponent: FilterComponent;
  searchSub: Subscription;

  constructor(
    public configService: ConfigurationService,
    private session: SessionService,
    private proService: ProjectService,
    private msgHandler: MessageHandlerService,
  ) { }

  ngOnInit(): void {
    if (window.sessionStorage && window.sessionStorage['projectTypeValue'] && window.sessionStorage['fromDetails']) {
      this.currentFilteredType = +window.sessionStorage['projectTypeValue'];
      window.sessionStorage.removeItem('fromDetails');
    }
    if (this.isSystemAdmin) {
      this.getConfigration();
    }
    if (!this.searchSub) {
      this.searchSub = this.filterComponent.filterTerms.pipe(
          debounceTime(500),
          distinctUntilChanged(),
          switchMap(projectName => {
            // reset project list
              this.listProject.currentPage = 1;
              this.listProject.searchKeyword = projectName;
              this.listProject.selectedRow = [];
              this.loading = true;
              let passInFilteredType: number = undefined;
              if (this.listProject.filteredType > 0) {
                  passInFilteredType = this.listProject.filteredType - 1;
              }
            return this.proService.listProjects( this.listProject.searchKeyword,
                passInFilteredType,  this.listProject.currentPage, this.listProject.pageSize)
                .pipe(finalize(() => {
                  this.loading = false;
                }));
          })).subscribe(response => {
                // Get total count
                if (response.headers) {
                    let xHeader: string = response.headers.get("X-Total-Count");
                    if (xHeader) {
                        this.listProject.totalCount = parseInt(xHeader, 0);
                    }
                }
                this.listProject.projects = response.body as Project[];
            }, error => {
                this.msgHandler.handleError(error);
            });
    }
  }

  ngOnDestroy() {
    if (this.searchSub) {
      this.searchSub.unsubscribe();
      this.searchSub = null;
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

  doFilterProjects(): void {
    this.listProject.doFilterProject(+this.selecteType);
  }

  refresh(): void {
    this.currentFilteredType = 0;
    this.projectName = "";
    this.listProject.refresh();
  }

}
