import { Injectable } from '@angular/core';
import {
  CanActivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  CanActivateChild
} from '@angular/router';
import { SessionService } from '../../shared/session.service';
import { harborRootRoute } from '../../shared/shared.const';

@Injectable()
export class SystemAdminGuard implements CanActivate, CanActivateChild {
  constructor(private authService: SessionService, private router: Router) { }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): boolean {
    let sessionUser = this.authService.getCurrentUser();

    let validation = sessionUser != null && sessionUser.has_admin_role > 0;
    if (!validation) {
      this.router.navigateByUrl(harborRootRoute);
    }

    return validation;
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): boolean {
    return this.canActivate(route, state);
  }
}
