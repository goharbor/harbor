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
  CanActivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  CanActivateChild
} from '@angular/router';
import { SessionService } from '../../shared/session.service';
import { Observable, of } from 'rxjs';
import { map, catchError } from 'rxjs/operators';
import { ProjectService } from "../../../lib/services";
import { CommonRoutes } from "../../../lib/entities/shared.const";

@Injectable()
export class MemberGuard implements CanActivate, CanActivateChild {
  constructor(
    private sessionService: SessionService,
    private projectService: ProjectService,
    private router: Router) {}

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    const projectId = route.params['id'];
    this.sessionService.setProjectMembers([]);

    const user = this.sessionService.getCurrentUser();
    if (user !== null) {
      return this.hasProjectPerm(state.url, projectId);
    }

    return this.sessionService.retrieveUser().pipe(
      () => {
        return this.hasProjectPerm(state.url, projectId);
      },
      catchError(err => {
        this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
        return of(false);
      })
    );
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    return this.canActivate(route, state);
  }

  hasProjectPerm(url: string, projectId: number): Observable<boolean> {
    // Note: current user will have the permission to visit the project when the user can get response from GET /projects/:id API.
    return this.projectService.getProject(projectId).pipe(
      map(() => {
        return true;
      }),
      catchError(err => {
        this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
        return of(false);
      })
    );
  }
}
