import { Injectable } from "@angular/core";
import {
  CanActivate,
  Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  CanActivateChild,
} from "@angular/router";
import { Observable } from "rxjs";
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { UserPermissionService, UserPrivilegeServeItem } from "../../../lib/services";
import { CommonRoutes } from "../../../lib/entities/shared.const";

@Injectable()
export class MemberPermissionGuard implements CanActivate, CanActivateChild {
  constructor(
    private router: Router,
    private errorHandler: ErrorHandler,
    private userPermission: UserPermissionService
  ) {}

  canActivate(
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot
  ): Observable<boolean> | boolean {
    const projectId = route.parent.params["id"];
    const permission = route.data.permissionParam as UserPrivilegeServeItem;
    return new Observable(observer => {
      this.userPermission
        .getPermission(projectId, permission.resource, permission.action)
        .subscribe(
          permissionRouter => {
            if (!permissionRouter) {
              this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
            }
            observer.next(permissionRouter);
          },
          error => {
            this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
            observer.next(false);
            this.errorHandler.error(error);
          }
        );
    });
  }

  canActivateChild(
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot
  ): Observable<boolean> | boolean {
    return this.canActivate(route, state);
  }
}
