import { Injectable } from '@angular/core';
import {
  CanActivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  CanActivateChild,
  NavigationExtras
} from '@angular/router';
import { SessionService } from '../../shared/session.service';
import { CommonRoutes } from '../../shared/shared.const';
import { AppConfigService } from '../../app-config.service';

@Injectable()
export class SystemAdminGuard implements CanActivate, CanActivateChild {
  constructor(
    private authService: SessionService,
    private router: Router,
    private appConfigService: AppConfigService) { }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return new Promise((resolve, reject) => {
      let user = this.authService.getCurrentUser();
      if (!user) {
        this.authService.retrieveUser()
          .then(() => {
            //updated user
            user = this.authService.getCurrentUser();
            if (user.has_admin_role > 0) {
              return resolve(true);
            } else {
              this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
              return resolve(false);
            }
          })
          .catch(error => {
            //Session retrieving failed then redirect to sign-in
            //no matter what status code is.
            //Please pay attention that route 'harborRootRoute' support anonymous user
            if (state.url != CommonRoutes.HARBOR_ROOT && !state.url.startsWith(CommonRoutes.EMBEDDED_SIGN_IN)) {
              let navigatorExtra: NavigationExtras = {
                queryParams: { "redirect_url": state.url }
              };
              this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], navigatorExtra);
              return resolve(false);
            } else {
              return resolve(true);
            }
          });
      } else {
        if (user.has_admin_role > 0) {
          return resolve(true);
        } else {
          this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
          return resolve(false);
        }
      }
    });
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return this.canActivate(route, state);
  }
}
