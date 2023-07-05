import { Injectable } from '@angular/core';
import {
    Router,
    ActivatedRouteSnapshot,
    RouterStateSnapshot,
} from '@angular/router';
import { Observable } from 'rxjs';
import { ErrorHandler } from '../units/error-handler';
import { UserPermissionService, UserPrivilegeServeItem } from '../services';
import { CommonRoutes } from '../entities/shared.const';

@Injectable({
    providedIn: 'root',
})
export class MemberPermissionGuard {
    constructor(
        private router: Router,
        private errorHandler: ErrorHandler,
        private userPermission: UserPermissionService
    ) {}

    canActivate(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<boolean> | boolean {
        const projectId = route.parent.params['id'];
        const permission = route.data.permissionParam as UserPrivilegeServeItem;
        return this.checkPermission(projectId, permission);
    }

    canActivateChild(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<boolean> | boolean {
        return this.canActivate(route, state);
    }
    checkPermission(
        projectId: number,
        permission: UserPrivilegeServeItem
    ): Observable<boolean> {
        return new Observable(observer => {
            this.userPermission
                .getPermission(
                    projectId,
                    permission.resource,
                    permission.action
                )
                .subscribe({
                    next: permissionRouter => {
                        if (!permissionRouter) {
                            this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
                        }
                        observer.next(permissionRouter);
                    },
                    error: error => {
                        this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
                        observer.next(false);
                        this.errorHandler.error(error);
                    },
                });
        });
    }
}
