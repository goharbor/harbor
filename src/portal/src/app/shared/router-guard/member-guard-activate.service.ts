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
    CanActivate,
    Router,
    ActivatedRouteSnapshot,
    RouterStateSnapshot,
    CanActivateChild,
} from '@angular/router';
import { SessionService } from '../services/session.service';
import { Observable, of } from 'rxjs';
import { map, catchError } from 'rxjs/operators';
import { ProjectService } from '../services';
import { CommonRoutes } from '../entities/shared.const';
import { HttpStatusCode } from '@angular/common/http';
import { delUrlParam } from '../units/utils';
import { UN_LOGGED_PARAM, YES } from 'src/app/account/sign-in/sign-in.service';

@Injectable({
    providedIn: 'root',
})
export class MemberGuard implements CanActivate, CanActivateChild {
    constructor(
        private sessionService: SessionService,
        private projectService: ProjectService,
        private router: Router
    ) {}

    canActivate(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<boolean> | boolean {
        const projectId = route.params['id'];
        this.sessionService.setProjectMembers([]);

        const user = this.sessionService.getCurrentUser();
        if (user !== null) {
            return this.hasProjectPerm(state.url, projectId, route);
        }

        return this.sessionService.retrieveUser().pipe(
            () => {
                return this.hasProjectPerm(state.url, projectId, route);
            },
            catchError(err => {
                this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
                return of(false);
            })
        );
    }

    canActivateChild(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<boolean> | boolean {
        return this.canActivate(route, state);
    }

    hasProjectPerm(
        url: string,
        projectId: number,
        route: ActivatedRouteSnapshot
    ): Observable<boolean> {
        // Note: current user will have the permission to visit the project when the user can get response from GET /projects/:id API.
        return this.projectService.getProject(projectId).pipe(
            map(() => {
                return true;
            }),
            catchError(err => {
                // User session timed out, then redirect to sign-in page
                if (
                    err.status === HttpStatusCode.Unauthorized &&
                    route.queryParams[UN_LOGGED_PARAM] !== YES
                ) {
                    this.sessionService.clear(); // because of SignInGuard, must clear user session before navigating to sign-in page
                    this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], {
                        queryParams: {
                            redirect_url: delUrlParam(
                                this.router.url,
                                UN_LOGGED_PARAM
                            ),
                        },
                    });
                } else {
                    this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
                }
                return of(false);
            })
        );
    }
}
