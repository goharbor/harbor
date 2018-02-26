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
import {
    Component,
    Output,
    Input,
    ChangeDetectionStrategy,
    ChangeDetectorRef,
    OnDestroy, EventEmitter
} from '@angular/core';
import { Router, NavigationExtras } from '@angular/router';

import { SessionService } from '../../../shared/session.service';
import { SearchTriggerService } from '../../../base/global-search/search-trigger.service';
import { ProjectTypes, RoleInfo} from '../../../shared/shared.const';
import { CustomComparator, doFiltering, doSorting, calculatePage } from '../../../shared/shared.utils';

import { Comparator, State } from 'clarity-angular';
import { MessageHandlerService } from '../../../shared/message-handler/message-handler.service';
import { StatisticHandler } from '../../../shared/statictics/statistic-handler.service';
import { Subscription } from 'rxjs/Subscription';
import { ConfirmationDialogService } from '../../../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../../../shared/confirmation-dialog/confirmation-message';
import { ConfirmationTargets, ConfirmationState, ConfirmationButtons } from '../../../shared/shared.const';
import {ProjectService} from "../../../project/project.service";
import {Project} from "../../../project/project";

@Component({
    selector: 'list-project-model',
    templateUrl: 'list-project-model.component.html',
    styleUrls: ['list-project-model.component.css'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListProjectModelComponent {
    projectTypes = ProjectTypes;
    loading: boolean = true;
    projects: Project[] = [];
    filteredType: number = 0;//All projects
    searchKeyword: string = "";
    ismodelOpen: boolean ;
    currentFilteredType: number = 0;//all projects
    projectName: string = "";
    selectedProject: Project;

    roleInfo = RoleInfo;
    repoCountComparator: Comparator<Project> = new CustomComparator<Project>("repo_count", "number");
    timeComparator: Comparator<Project> = new CustomComparator<Project>("creation_time", "date");
    accessLevelComparator: Comparator<Project> = new CustomComparator<Project>("public", "number");
    roleComparator: Comparator<Project> = new CustomComparator<Project>("current_user_role_id", "number");
    currentPage: number = 1;
    totalCount: number = 0;
    pageSize: number = 10;
    currentState: State;
    @Output() selectedPro = new EventEmitter<Project>();

    constructor(
        private session: SessionService,
        private router: Router,
        private searchTrigger: SearchTriggerService,
        private proService: ProjectService,
        private msgHandler: MessageHandlerService,
        private statisticHandler: StatisticHandler,
        private deletionDialogService: ConfirmationDialogService,
        private ref: ChangeDetectorRef) {
    }

    get selecteType(): number {
        return this.currentFilteredType;
    }
    set selecteType(_project: number) {
        this.currentFilteredType = _project;
        if (window.sessionStorage) {
            window.sessionStorage['projectTypeValue'] = _project;
        }
    }


    clrLoad(state: State) {
        this.selectedProject = null;
        //Keep state for future filtering and sorting
        this.currentState = state;

        let pageNumber: number = calculatePage(state);
        if (pageNumber <= 0) { pageNumber = 1; }

        this.loading = true;

        let passInFilteredType: number = undefined;
        if (this.filteredType > 0) {
            passInFilteredType = this.filteredType - 1;
        }
        this.proService.listProjects(this.searchKeyword, passInFilteredType, pageNumber, this.pageSize).toPromise()
            .then(response => {
                //Get total count
                if (response.headers) {
                    let xHeader: string = response.headers.get("X-Total-Count");
                    if (xHeader) {
                        this.totalCount = parseInt(xHeader, 0);
                    }
                }

                this.projects = response.json() as Project[];
                //Do customising filtering and sorting
                this.projects = doFiltering<Project>(this.projects, state);
                this.projects = doSorting<Project>(this.projects, state);

                this.loading = false;
            })
            .catch(error => {
                this.loading = false;
                this.msgHandler.handleError(error);
            });

        //Force refresh view
        let hnd = setInterval(() => this.ref.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 3000);
    }

    openModel(): void {
        this.selectedProject = null;
        this.ismodelOpen = true;
        //Force refresh view
        let hnd = setInterval(() => this.ref.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 2000);
    }

    refresh(): void {
        this.currentPage = 1;
        this.filteredType = 0;
        this.searchKeyword = '';

        this.reload();
    }

    doFilterProject(): void {
        this.currentPage = 1;
        this.filteredType = this.selecteType;
        this.reload();
    }

    doSearchProject(proName: string): void {
        this.projectName = proName;
        this.currentPage = 1;
        this.searchKeyword = proName;
        this.reload();
    }

    reload(): void {
        let st: State = this.currentState;
        if (!st) {
            st = {
                page: {}
            };
        }
        st.page.from = 0;
        st.page.to = this.pageSize - 1;
        st.page.size = this.pageSize;

        this.clrLoad(st);
    }

    oKModel() {
        this.ismodelOpen = false;
        this.selectedPro.emit(this.selectedProject);
    }

    closeModel(): void {
        this.ismodelOpen = false;
    }
}
