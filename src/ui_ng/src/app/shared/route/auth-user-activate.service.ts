import { Injectable } from '@angular/core';
import {
  CanActivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  CanActivateChild,
  NavigationExtras
} from '@angular/router';
import { SessionService } from '../../shared/session.service';
import { harborRootRoute, signInRoute } from '../../shared/shared.const';

@Injectable()
export class AuthCheckGuard implements CanActivate, CanActivateChild {
  constructor(private authService: SessionService, private router: Router) { }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return new Promise((resolve, reject) => {
      let user = this.authService.getCurrentUser();
      if (!user) {
        this.authService.retrieveUser()
          .then(() => resolve(true))
          .catch(error => {
            //Session retrieving failed then redirect to sign-in
            //no matter what status code is.
            //Please pay attention that route 'harborRootRoute' support anonymous user
            if (state.url != harborRootRoute) {
              let navigatorExtra: NavigationExtras = {
                queryParams: { "redirect_url": state.url }
              };
              this.router.navigate([signInRoute], navigatorExtra);
              return resolve(false);
            } else {
              return resolve(true);
            }
          });
      } else {
        return resolve(true);
      }
    });
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return this.canActivate(route, state);
  }
}
