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
    ActivatedRouteSnapshot,
    Router,
    RouterStateSnapshot,
    UrlTree,
} from '@angular/router';
import { Observable, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { ErrorHandler } from '../../../shared/units/error-handler';
import {
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../shared/services';
import { CommonRoutes } from '../../../shared/entities/shared.const';

export const TAG_RETENTION_ROUTE = 'tag-retention';
export const IMMUTABLE_TAG_ROUTE = 'immutable-tag';

interface TagFeaturePermissions {
    hasTagRetentionPermission: boolean;
    hasTagImmutablePermission: boolean;
}

/**
 * The Policy (tag-strategy) page hosts both the tag retention and the tag
 * immutability features, so a user holding either permission must be able
 * to open it.
 */
@Injectable({
    providedIn: 'root',
})
export class TagStrategyGuardService {
    constructor(
        private router: Router,
        private errorHandler: ErrorHandler,
        private userPermission: UserPermissionService
    ) {}

    canActivate(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<boolean | UrlTree> {
        const projectId = route.parent.params['id'];
        return this.getPermissions(projectId).pipe(
            map(permissions =>
                permissions.hasTagRetentionPermission ||
                permissions.hasTagImmutablePermission
                    ? true
                    : this.router.parseUrl(CommonRoutes.HARBOR_DEFAULT)
            )
        );
    }

    getPermissions(projectId: number): Observable<TagFeaturePermissions> {
        return this.userPermission
            .hasProjectPermissions(projectId, [
                {
                    resource: USERSTATICPERMISSION.TAG_RETENTION.KEY,
                    action: USERSTATICPERMISSION.TAG_RETENTION.VALUE.READ,
                },
                {
                    resource: USERSTATICPERMISSION.IMMUTABLE_TAG.KEY,
                    action: USERSTATICPERMISSION.IMMUTABLE_TAG.VALUE.LIST,
                },
            ])
            .pipe(
                map(
                    ([
                        hasTagRetentionPermission,
                        hasTagImmutablePermission,
                    ]) => ({
                        hasTagRetentionPermission: !!hasTagRetentionPermission,
                        hasTagImmutablePermission: !!hasTagImmutablePermission,
                    })
                ),
                catchError(error => {
                    this.errorHandler.error(error);
                    return of({
                        hasTagRetentionPermission: false,
                        hasTagImmutablePermission: false,
                    });
                })
            );
    }
}

/**
 * Redirects the empty tag-strategy child route to the first tab the user is
 * allowed to see, instead of always sending them to tag retention.
 */
@Injectable({
    providedIn: 'root',
})
export class TagStrategyDefaultTabGuardService {
    constructor(
        private router: Router,
        private tagStrategyGuard: TagStrategyGuardService
    ) {}

    canActivate(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<UrlTree> {
        const projectId = route.parent.parent.parent.params['id'];
        const segments = route.pathFromRoot
            .map(snapshot => snapshot.url.map(segment => segment.path))
            .reduce((acc, paths) => acc.concat(paths), []);
        return this.tagStrategyGuard.getPermissions(projectId).pipe(
            map(permissions => {
                if (permissions.hasTagRetentionPermission) {
                    return this.router.createUrlTree([
                        '/',
                        ...segments,
                        TAG_RETENTION_ROUTE,
                    ]);
                }
                if (permissions.hasTagImmutablePermission) {
                    return this.router.createUrlTree([
                        '/',
                        ...segments,
                        IMMUTABLE_TAG_ROUTE,
                    ]);
                }
                return this.router.parseUrl(CommonRoutes.HARBOR_DEFAULT);
            })
        );
    }
}
