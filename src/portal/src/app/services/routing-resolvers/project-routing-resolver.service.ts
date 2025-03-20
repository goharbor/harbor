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
import { Injectable } from '@angular/core';
import {
    Router,
    RouterStateSnapshot,
    ActivatedRouteSnapshot,
} from '@angular/router';
import { Project } from '../../base/project/project';
import { SessionService } from '../../shared/services/session.service';
import { Observable } from 'rxjs';
import { map, catchError } from 'rxjs/operators';
import { ProjectService } from '../../shared/services';
import { RoleInfo, Roles } from '../../shared/entities/shared.const';

@Injectable({
    providedIn: 'root',
})
export class ProjectRoutingResolver {
    constructor(
        private sessionService: SessionService,
        private projectService: ProjectService,
        private router: Router
    ) {}

    resolve(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<Project> {
        // Support both parameters and query parameters
        let projectId = route.params['id'];
        if (!projectId) {
            projectId = route.queryParams['project_id'];
        }
        return this.projectService.getProjectFromCache(projectId).pipe(
            map(
                (project: Project) => {
                    if (project) {
                        let currentUser = this.sessionService.getCurrentUser();
                        if (currentUser) {
                            if (currentUser.has_admin_role) {
                                project.has_project_admin_role = true;
                                project.is_member = true;
                                project.role_name = 'MEMBER.SYS_ADMIN';
                            } else {
                                project.has_project_admin_role =
                                    project.current_user_role_id ===
                                    Roles.PROJECT_ADMIN;
                                project.is_member =
                                    project.current_user_role_id > 0;
                                project.role_name =
                                    RoleInfo[project.current_user_role_id];
                            }
                        }
                        return project;
                    } else {
                        this.router.navigate(['/harbor', 'projects']);
                        return null;
                    }
                },
                catchError(error => {
                    this.router.navigate(['/harbor', 'projects']);
                    return null;
                })
            )
        );
    }
}
