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
import { Component } from '@angular/core';
import { Project } from '../../../../../../ng-swagger-gen/models/project';
import { Router } from '@angular/router';
import { PermissionsKinds } from '../system-robot-util';
import { RobotPermission } from '../../../../../../ng-swagger-gen/models/robot-permission';
import { PermissionSelectPanelModes } from '../../../../shared/components/robot-permissions-panel/robot-permissions-panel.component';
import { ProjectService } from '../../../../../../ng-swagger-gen/services/project.service';
import { ClrDatagridStateInterface } from '@clr/angular';
import { finalize } from 'rxjs/operators';
import { PAGE_SIZE_OPTIONS } from 'src/app/shared/entities/shared.const';

@Component({
    selector: 'app-projects-modal',
    templateUrl: './projects-modal.component.html',
    styleUrls: ['./projects-modal.component.scss'],
})
export class ProjectsModalComponent {
    clrPageSizeOptions: number[] = PAGE_SIZE_OPTIONS;
    projectsModalOpened: boolean = false;
    robotName: string;
    cachedAllProjects: Project[];
    permissions: RobotPermission[] = [];
    pageSize: number = 10;
    loading: boolean = false;
    constructor(
        private router: Router,
        private projectService: ProjectService
    ) {}

    close() {
        this.projectsModalOpened = false;
    }
    clrDgRefresh(state?: ClrDatagridStateInterface) {
        if (this.permissions.length) {
            if (state) {
                this.pageSize = state.page.size;
                this.getProjectFromBackend(
                    this.permissions.slice(state.page.from, state.page.to + 1)
                );
            } else {
                this.getProjectFromBackend(
                    this.permissions.slice(0, this.pageSize)
                );
            }
        }
    }
    getProjectFromBackend(permissions: RobotPermission[]) {
        const projectNames: string[] = [];
        permissions?.forEach(item => {
            if (item?.kind === PermissionsKinds.PROJECT) {
                projectNames.push(item?.namespace);
            }
        });
        this.loading = true;
        this.projectService
            .listProjects({
                withDetail: false,
                page: 1,
                pageSize: permissions?.length,
                q: encodeURIComponent(`name={${projectNames.join(' ')}}`),
            })
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(res => {
                if (res?.length) {
                    this.cachedAllProjects = res;
                }
            });
    }
    getProject(p: RobotPermission): Project {
        if (this.cachedAllProjects && this.cachedAllProjects.length) {
            for (let i = 0; i < this.cachedAllProjects.length; i++) {
                if (p.namespace === this.cachedAllProjects[i].name) {
                    return this.cachedAllProjects[i];
                }
            }
        }
        return null;
    }
    goToLink(proId: number): void {
        this.router.navigate(['harbor', 'projects', proId]);
    }

    protected readonly PermissionSelectPanelModes = PermissionSelectPanelModes;
}
